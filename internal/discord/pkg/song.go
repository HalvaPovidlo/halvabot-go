package pkg

import "github.com/bwmarrin/discordgo"

type SongRequest struct {
	Metadata     *Metadata       `json:"metadata"`
	ServiceName  string          `json:"service_name"`  // Name of service used for this queue entry
	ServiceColor int             `json:"service_color"` // Color of service used for this queue entry
	Requester    *discordgo.User `json:"-"`             // TODO: decide pass requester or not
}

type Metadata struct {
	Artists      []MetadataArtist `json:"artists,omitempty"`
	Title        string           `json:"title,omitempty"`
	DisplayURL   string           `json:"display_url,omitempty"`
	StreamURL    string           `json:"-"`
	Duration     float64          `json:"duration,omitempty"`
	ArtworkURL   string           `json:"artwork_url,omitempty"`
	ThumbnailURL string           `json:"thumbnail_url,omitempty"`
}

// MetadataArtist stores the data about an artist
type MetadataArtist struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}
