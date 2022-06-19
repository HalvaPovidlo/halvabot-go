package pkg

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

type ServiceName string

const (
	ServiceYouTube ServiceName = "youtube"
)

type SongID struct {
	ID      string
	Service ServiceName
}

type PlayDate struct {
	time.Time
}

type Song struct {
	Title        string      `firestore:"title,omitempty" csv:"title" json:"title,omitempty"`
	URL          string      `firestore:"url,omitempty" csv:"url,omitempty" json:"url,omitempty"`
	Service      ServiceName `firestore:"service,omitempty" csv:"service,omitempty" json:"service,omitempty"`
	ArtistName   string      `firestore:"artist_name,omitempty" csv:"artist_name,omitempty" json:"artist_name,omitempty"`
	ArtistURL    string      `firestore:"artist_url,omitempty" csv:"artist_url,omitempty" json:"artist_url,omitempty"`
	ArtworkURL   string      `firestore:"artwork_url,omitempty" csv:"artwork_url,omitempty" json:"artwork_url,omitempty"`
	ThumbnailURL string      `firestore:"thumbnail_url,omitempty" csv:"thumbnail_url,omitempty" json:"thumbnail_url,omitempty"`
	Playbacks    int         `firestore:"playbacks,omitempty" csv:"playbacks" json:"playbacks,omitempty"`
	LastPlay     PlayDate    `firestore:"last_play,omitempty" csv:"last_play,omitempty" json:"last_play,omitempty"`

	ID        SongID          `firestore:"-" csv:"-" json:"-"`
	Requester *discordgo.User `firestore:"-" csv:"-" json:"-"`
	StreamURL string          `firestore:"-" csv:"-" json:"-"`
	Duration  float64         `firestore:"-" csv:"-" json:"-"`
}

type User struct {
	Name  string
	Songs []Song
}

type PlayerStatus struct {
	Loop  bool  `json:"loop"`
	Radio bool  `json:"radio"`
	Now   *Song `json:"now,omitempty"`
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

func (s *Song) MergeNoOverride(new *Song) {
	if new == nil {
		return
	}
	if s.Title == "" {
		s.Title = new.Title
	}
	if s.URL == "" {
		s.URL = new.URL
	}
	if s.Service == "" {
		s.Service = new.Service
	}
	if s.ArtistName == "" {
		s.ArtistName = new.ArtistName
	}
	if s.ArtistURL == "" {
		s.ArtistURL = new.ArtistURL
	}
	if s.ArtworkURL == "" {
		s.ArtworkURL = new.ArtworkURL
	}
	if s.ThumbnailURL == "" {
		s.ThumbnailURL = new.ThumbnailURL
	}
	if s.Playbacks == 0 {
		s.Playbacks = new.Playbacks
	}
	if s.LastPlay.IsZero() {
		s.LastPlay = new.LastPlay
	}
}

func GetIDFromURL(url string) SongID {
	var id SongID
	if TestYoutubeURL(url) {
		id.Service = ServiceYouTube
		// TODO: trim all urls https://stackoverflow.com/questions/19377262/regex-for-youtube-url
		url = strings.TrimPrefix(url, `https://www.youtube.com/watch?v=`)
		url = strings.TrimPrefix(url, `https://youtube.com/watch?v=`)
		id.ID = url
		return id
	}
	return id
}

func TestYoutubeURL(url string) bool {
	test, _ := regexp.MatchString("^((?:https?:)?\\/\\/)?((?:www|m)\\.)?((?:youtube(-nocookie)?\\.com|youtu.be))(\\/(?:[\\w\\-]+\\?v=|embed\\/|v\\/)?)([\\w\\-]+)(\\S+)?$", url)
	return test
}
