package storage

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type SongID struct {
	ID      string
	Service ServiceName
}

type ServiceName string

const (
	ServiceYouTube ServiceName = "youtube"
)

type PlayDate struct {
	time.Time
}

type Song struct {
	Title        string   `firestore:"title,omitempty" csv:"title" json:"title,omitempty"`
	URL          string   `firestore:"url,omitempty" csv:"url,omitempty" json:"url,omitempty"`
	Service      string   `firestore:"service,omitempty" csv:"service,omitempty" json:"service,omitempty"`
	ArtistName   string   `firestore:"artist_name,omitempty" csv:"artist_name,omitempty" json:"artist_name,omitempty"`
	ArtistURL    string   `firestore:"artist_url,omitempty" csv:"artist_url,omitempty" json:"artist_url,omitempty"`
	ArtworkURL   string   `firestore:"artwork_url,omitempty" csv:"artwork_url,omitempty" json:"artwork_url,omitempty"`
	ThumbnailURL string   `firestore:"thumbnail_url,omitempty" csv:"thumbnail_url,omitempty" json:"thumbnail_url,omitempty"`
	Playbacks    int      `firestore:"playbacks,omitempty" csv:"playbacks" json:"playbacks,omitempty"`
	LastPlay     PlayDate `firestore:"last_play,omitempty" csv:"last_play,omitempty" json:"last_play,omitempty"`
	ID           SongID   `firestore:"-" csv:"-" json:"-"`
}

type User struct {
	Name  string
	Songs []Song
}

func (date *PlayDate) UnmarshalCSV(csv string) error {
	in := strings.Split(csv, "/")
	if len(in) < 3 {
		return errors.New("wrong time format")
	}
	toParse := fmt.Sprintf("%s/%s/%s", in[1], in[0], in[2])
	t, err := time.Parse("01/02/2006", toParse)
	if err != nil {
		return err
	}
	*date = PlayDate{t}
	return nil
}

func (date PlayDate) String() string {
	return date.Time.String()
}

func (id SongID) String() string {
	return string(id.Service) + "_" + id.ID
}
