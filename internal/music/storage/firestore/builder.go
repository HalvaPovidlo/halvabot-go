package firestore

import (
	"time"

	"cloud.google.com/go/firestore"
	"github.com/pkg/errors"

	"github.com/HalvaPovidlo/halvabot-go/internal/pkg"
)

type playDate struct {
	time.Time
}

type oldSong struct {
	Title        string          `firestore:"title,omitempty" csv:"title" json:"title,omitempty"`
	URL          string          `firestore:"url,omitempty" csv:"url,omitempty" json:"url,omitempty"`
	Service      pkg.ServiceName `firestore:"service,omitempty" csv:"service,omitempty" json:"service,omitempty"`
	ArtistName   string          `firestore:"artist_name,omitempty" csv:"artist_name,omitempty" json:"artist_name,omitempty"`
	ArtistURL    string          `firestore:"artist_url,omitempty" csv:"artist_url,omitempty" json:"artist_url,omitempty"`
	ArtworkURL   string          `firestore:"artwork_url,omitempty" csv:"artwork_url,omitempty" json:"artwork_url,omitempty"`
	ThumbnailURL string          `firestore:"thumbnail_url,omitempty" csv:"thumbnail_url,omitempty" json:"thumbnail_url,omitempty"`
	Playbacks    int             `firestore:"playbacks,omitempty" csv:"playbacks" json:"playbacks,omitempty"`
	LastPlay     playDate        `firestore:"last_play,omitempty" csv:"last_play,omitempty" json:"last_play,omitempty"`
}

func parseSongDoc(doc *firestore.DocumentSnapshot) (pkg.Song, error) {
	var s pkg.Song
	err := doc.DataTo(&s)
	if err != nil {
		var old *oldSong
		err = doc.DataTo(&old)
		if err != nil {
			return pkg.Song{}, errors.Wrap(err, "unable to marshal song data")
		}
		s = buildNewSong(old)
	}
	return s, nil
}

func buildNewSong(s *oldSong) pkg.Song {
	return pkg.Song{
		Title:        s.Title,
		URL:          s.URL,
		Service:      s.Service,
		ArtistName:   s.ArtistName,
		ArtistURL:    s.ArtistURL,
		ArtworkURL:   s.ArtworkURL,
		ThumbnailURL: s.ThumbnailURL,
		Playbacks:    s.Playbacks,
		LastPlay:     s.LastPlay.Time,
	}
}
