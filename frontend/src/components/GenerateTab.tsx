import { useState, useEffect } from 'react'
import { Button } from './ui/button'
import { Input } from './ui/input'
import { Label } from './ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select'
import { Checkbox } from './ui/checkbox'
import { Music, Loader2, Trash2 } from 'lucide-react'
import type { LogEntry, Playlist } from '../App'

const DATA_SOURCES = [
  { value: 'lastfm', label: 'Last.fm' },
  { value: 'spotify', label: 'Spotify' },
  { value: 'musicbrainz', label: 'MusicBrainz' },
]

const PLAYLIST_TYPES = [
  { value: 'top_songs', label: 'Top Songs' },
  { value: 'mixed_songs', label: 'Mixed Songs' },
  { value: 'similar', label: 'Similar Songs' },
  { value: 'tag', label: 'Tag/Genre Radio' },
]

export function GenerateTab() {
  const [dataSource, setDataSource] = useState('lastfm')
  const [playlistType, setPlaylistType] = useState('top_songs')
  const [artist, setArtist] = useState('')
  const [tag, setTag] = useState('')
  const [limit, setLimit] = useState(50)
  const [useAlbumArtist, setUseAlbumArtist] = useState(false)
  const [isGenerating, setIsGenerating] = useState(false)
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [playlists, setPlaylists] = useState<Playlist[]>([])
  const [artists, setArtists] = useState<string[]>([])
  const [enabledSources, setEnabledSources] = useState<string[]>([])

  useEffect(() => {
    loadInitialData()
    const interval = setInterval(refreshLogs, 2000)
    return () => clearInterval(interval)
  }, [])

  const loadInitialData = async () => {
    if (!window.go?.cmd?.App) return
    
    try {
      const [artistList, playlistList, sources] = await Promise.all([
        window.go.cmd.App.GetUniqueArtists(),
        window.go.cmd.App.GetAllPlaylists(),
        window.go.cmd.App.GetEnabledSources(),
      ])
      setArtists(artistList || [])
      setPlaylists(playlistList || [])
      setEnabledSources(sources || [])
    } catch (error) {
      console.error('Failed to load data:', error)
    }
  }

  const refreshLogs = async () => {
    if (!window.go?.cmd?.App) return
    try {
      const newLogs = await window.go.cmd.App.GetLogs()
      setLogs(newLogs)
    } catch (error) {
      console.error('Failed to refresh logs:', error)
    }
  }

  const handleGenerate = async () => {
    if (!window.go?.cmd?.App) return
    
    setIsGenerating(true)
    try {
      await window.go.cmd.App.GeneratePlaylist(
        dataSource,
        playlistType,
        artist,
        tag,
        limit,
        useAlbumArtist
      )
      
      const playlistList = await window.go.cmd.App.GetAllPlaylists()
      setPlaylists(playlistList || [])
      
      await refreshLogs()
    } catch (error) {
      console.error('Generate failed:', error)
    } finally {
      setIsGenerating(false)
    }
  }

  const handleDeletePlaylist = async (id: number) => {
    if (!window.go?.cmd?.App) return
    
    try {
      await window.go.cmd.App.DeletePlaylist(id)
      const playlistList = await window.go.cmd.App.GetAllPlaylists()
      setPlaylists(playlistList || [])
    } catch (error) {
      console.error('Delete failed:', error)
    }
  }

  const needsArtist = ['top_songs', 'mixed_songs', 'similar'].includes(playlistType)
  const needsTag = playlistType === 'tag'
  const isSourceEnabled = enabledSources.includes(dataSource)

  return (
    <div className="space-y-6">
      <div className="rounded-lg border bg-card p-6">
        <h2 className="text-lg font-semibold mb-4">Generate Playlist</h2>
        
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label>Data Source</Label>
            <Select value={dataSource} onValueChange={setDataSource}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {DATA_SOURCES.map((source) => (
                  <SelectItem 
                    key={source.value} 
                    value={source.value}
                    disabled={!enabledSources.includes(source.value)}
                  >
                    {source.label}
                    {!enabledSources.includes(source.value) && ' (Not configured)'}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label>Playlist Type</Label>
            <Select value={playlistType} onValueChange={setPlaylistType}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {PLAYLIST_TYPES.map((type) => (
                  <SelectItem key={type.value} value={type.value}>
                    {type.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {needsArtist && (
            <div className="space-y-2">
              <Label>Artist</Label>
              <Input
                placeholder="Enter artist name..."
                value={artist}
                onChange={(e) => setArtist(e.target.value)}
                list="artist-list"
              />
              <datalist id="artist-list">
                {artists.map((a) => (
                  <option key={a} value={a} />
                ))}
              </datalist>
            </div>
          )}

          {needsTag && (
            <div className="space-y-2">
              <Label>Tag / Genre</Label>
              <Input
                placeholder="e.g., Death Metal, Jazz, Progressive Rock..."
                value={tag}
                onChange={(e) => setTag(e.target.value)}
              />
            </div>
          )}

          <div className="space-y-2">
            <Label>Max Songs</Label>
            <Input
              type="number"
              min={1}
              max={200}
              value={limit}
              onChange={(e) => setLimit(parseInt(e.target.value) || 50)}
            />
          </div>
        </div>

        {/* Use Album Artist Option */}
        <div className="mt-4 flex items-start space-x-3">
          <Checkbox
            id="useAlbumArtist"
            checked={useAlbumArtist}
            onCheckedChange={(checked) => setUseAlbumArtist(checked === true)}
          />
          <div className="grid gap-1.5 leading-none">
            <Label
              htmlFor="useAlbumArtist"
              className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 cursor-pointer"
            >
              Use Album Artist if available
            </Label>
            <p className="text-sm text-muted-foreground">
              When enabled, songs will be matched using the Album Artist field instead of the Artist field.
              This is useful for compilations and albums with multiple artists. If a song has no Album Artist,
              the regular Artist field will be used as a fallback.
            </p>
          </div>
        </div>

        <div className="mt-4">
          <Button 
            onClick={handleGenerate} 
            disabled={isGenerating || !isSourceEnabled || (needsArtist && !artist) || (needsTag && !tag)}
          >
            {isGenerating ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Generating...
              </>
            ) : (
              <>
                <Music className="mr-2 h-4 w-4" />
                Generate Playlist
              </>
            )}
          </Button>
          
          {!isSourceEnabled && (
            <p className="text-sm text-destructive mt-2">
              Please configure {DATA_SOURCES.find(s => s.value === dataSource)?.label} API credentials in Settings.
            </p>
          )}
        </div>
      </div>

      {/* Generated Playlists */}
      {playlists.length > 0 && (
        <div className="rounded-lg border bg-card p-6">
          <h3 className="text-lg font-semibold mb-4">Generated Playlists</h3>
          <div className="space-y-2">
            {playlists.map((playlist) => (
              <div key={playlist.ID} className="flex items-center justify-between p-3 rounded bg-secondary/50">
                <div>
                  <p className="font-medium">{playlist.name}</p>
                  <p className="text-sm text-muted-foreground">
                    {playlist.song_count} songs • {playlist.data_source}
                  </p>
                </div>
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => handleDeletePlaylist(playlist.ID)}
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Logs Section */}
      <div className="rounded-lg border bg-card p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold">Logs</h3>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => window.go?.cmd?.App?.ClearLogs()}
          >
            Clear
          </Button>
        </div>
        
        <div className="log-container h-48 overflow-y-auto rounded bg-secondary/50 p-3">
          {logs.length === 0 ? (
            <p className="text-muted-foreground text-sm">No logs yet.</p>
          ) : (
            logs.map((log, index) => (
              <div key={index} className={`log-entry ${log.level}`}>
                <span className="log-time">
                  {new Date(log.time).toLocaleTimeString()}
                </span>
                <span>{log.message}</span>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  )
}
