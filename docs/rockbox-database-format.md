# Rockbox TagCache Database Format

This document describes the Rockbox TagCache database format used by Rocklist to parse music library information from Rockbox devices.

## Overview

Rockbox stores its music library metadata in a binary database called TagCache. The database consists of multiple files stored in the `.rockbox` directory on the device:

- `database_idx.tcd` - Master index file containing entry metadata and numeric tags
- `database_0.tcd` through `database_8.tcd` - Tag files containing string data

## File Locations

All database files are located in:
```
<device_root>/.rockbox/
├── database_idx.tcd    # Master index
├── database_0.tcd      # Artist tags
├── database_1.tcd      # Album tags
├── database_2.tcd      # Genre tags
├── database_3.tcd      # Title tags
├── database_4.tcd      # Filename tags
├── database_5.tcd      # Composer tags
├── database_6.tcd      # Comment tags
├── database_7.tcd      # Album Artist tags
└── database_8.tcd      # Grouping tags
```

## Magic Number

All TagCache files use the same magic number in their header:

```
TAGCACHE_MAGIC = 0x54434810
```

This is ASCII "TCH" followed by version byte `0x10`. The magic number can appear in either little-endian or big-endian format depending on the device's CPU architecture:
- **Little-endian**: ARM-based players (most modern devices)
- **Big-endian**: Coldfire, SH1-based players (older devices)

## Common Header Structure

All TagCache files share a common 12-byte header:

```c
struct tagcache_header {
    int32_t magic;       // Magic number (0x54434810)
    int32_t datasize;    // Total data size in bytes (excluding header)
    int32_t entry_count; // Number of entries in this file
};
```

| Offset | Size | Field | Description |
|--------|------|-------|-------------|
| 0x00 | 4 | magic | Magic number `0x54434810` |
| 0x04 | 4 | datasize | Size of data following the header |
| 0x08 | 4 | entry_count | Number of entries in the file |

## Master Index File (database_idx.tcd)

The master index file has an extended header followed by index entries for each song.

### Master Header (24 bytes)

```c
struct master_header {
    struct tagcache_header tch;  // Common header (12 bytes)
    int32_t serial;              // Increasing counting number
    int32_t commitid;            // Number of commits so far
    int32_t dirty;               // Dirty flag
};
```

| Offset | Size | Field | Description |
|--------|------|-------|-------------|
| 0x00 | 12 | tch | Common TagCache header |
| 0x0C | 4 | serial | Serial number |
| 0x10 | 4 | commitid | Commit ID |
| 0x14 | 4 | dirty | Database dirty flag |

### Index Entry Structure

Following the master header, there's one `index_entry` for each song:

```c
struct index_entry {
    int32_t tag_seek[TAG_COUNT]; // Seek positions for each tag
    int32_t flag;                // Status flags
};
```

The `tag_seek` array contains:
- For **string tags** (0-8): Byte offset into the corresponding tag file
- For **numeric tags** (9+): The actual numeric value

### Tag Types

| Index | Tag Name | Type | Description |
|-------|----------|------|-------------|
| 0 | tag_artist | String | Artist name |
| 1 | tag_album | String | Album name |
| 2 | tag_genre | String | Genre |
| 3 | tag_title | String | Track title |
| 4 | tag_filename | String | File path |
| 5 | tag_composer | String | Composer |
| 6 | tag_comment | String | Comment |
| 7 | tag_albumartist | String | Album artist |
| 8 | tag_grouping | String | Grouping |
| 9 | tag_year | Numeric | Year |
| 10 | tag_discnumber | Numeric | Disc number |
| 11 | tag_tracknumber | Numeric | Track number |
| 12 | tag_bitrate | Numeric | Bitrate (kbps) |
| 13 | tag_length | Numeric | Duration (ms) |
| 14 | tag_playcount | Numeric | Play count |
| 15 | tag_rating | Numeric | Rating (0-10) |
| 16 | tag_playtime | Numeric | Total play time |
| 17 | tag_lastplayed | Numeric | Last played timestamp |
| 18 | tag_commitid | Numeric | Commit ID |
| 19 | tag_mtime | Numeric | File modification time |

### Entry Flags

```c
#define FLAG_DELETED     0x0001  // Entry has been removed from db
#define FLAG_DIRCACHE    0x0002  // Filename is a dircache pointer
#define FLAG_DIRTYNUM    0x0004  // Numeric data has been modified
#define FLAG_TRKNUMGEN   0x0008  // Track number has been generated
#define FLAG_RESURRECTED 0x0010  // Statistics data has been resurrected
```

## Tag Files (database_X.tcd)

Tag files store string data for each tag type. Each file contains a header followed by tag entries.

### Tag File Entry Structure

```c
struct tagfile_entry {
    int32_t tag_length;  // Length of tag data including null terminator
    int32_t idx_id;      // Index ID in master file (-1 for unique tags)
    char tag_data[0];    // Variable-length string data
};
```

| Offset | Size | Field | Description |
|--------|------|-------|-------------|
| 0x00 | 4 | tag_length | Length of tag_data including `\0` |
| 0x04 | 4 | idx_id | Master index entry this tag belongs to |
| 0x08 | N | tag_data | Null-terminated string (N = tag_length) |

### Reading Tag Files

To read a tag file:

1. Read the 12-byte header
2. Verify the magic number
3. For each entry:
   - Read 8 bytes (tag_length + idx_id)
   - Read tag_length bytes of string data
   - Map the string to idx_id for lookup

### Entry Alignment

Tag entries may be aligned to `TAGFILE_ENTRY_CHUNK_LENGTH` (8 bytes) for performance. When writing, padding bytes may be added after `tag_data`.

## Parsing Algorithm

### Step 1: Read Master Index

```
1. Open database_idx.tcd
2. Read and verify master header (magic = 0x54434810)
3. Read entry_count from header
4. For each entry, read tag_seek positions and flags
```

### Step 2: Read Tag Files

```
For each tag file (database_0.tcd through database_8.tcd):
1. Open the file
2. Read and verify header
3. Read all entries
4. Build a map: idx_id -> tag_string
```

### Step 3: Build Song Records

```
For each entry in master index:
1. Skip if FLAG_DELETED is set
2. For string tags: lookup value in corresponding tag file map
3. For numeric tags: read directly from tag_seek array
4. Create song record with all metadata
```

## Endianness Handling

Rockbox supports both native and foreign endianness. When reading:

1. Try reading magic as little-endian
2. If it doesn't match, try big-endian
3. If big-endian matches, swap all subsequent integer reads

```go
magic := binary.LittleEndian.Uint32(header[0:4])
if magic != TagCacheMagic {
    magic = binary.BigEndian.Uint32(header[0:4])
    if magic != TagCacheMagic {
        return error("invalid magic number")
    }
    // Use big-endian for all subsequent reads
}
```

## References

- [Rockbox TagCache Source Code](https://github.com/Rockbox/rockbox/blob/master/apps/tagcache.c)
- [Rockbox TagCache Header](https://github.com/Rockbox/rockbox/blob/master/apps/tagcache.h)
- [Rockbox Wiki - TagCache](https://www.rockbox.org/wiki/TagCache)

## Implementation Notes

### Fallback to Filesystem Scan

If the TagCache database cannot be read (missing files, invalid format, etc.), Rocklist falls back to scanning the filesystem for audio files and extracting metadata from filenames.

### Supported Audio Formats

When scanning the filesystem, the following extensions are recognized:
- `.mp3`, `.flac`, `.ogg`, `.m4a`, `.aac`
- `.wav`, `.wma`, `.ape`, `.mpc`, `.opus`

### Database Regeneration

The Rockbox database can be regenerated on the device via:
- Settings → General Settings → Database → Initialize Now

Or by deleting all `database_*.tcd` files and letting Rockbox rebuild them on next boot.
