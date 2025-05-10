package models

// TopTrack represents a top track from an API response
type TopTrack struct {
	Name     string
	Artist   string
	Album    string
	Rank     int
	Filename string
	Found    bool // indicates if the track was found in the Rockbox database
} 