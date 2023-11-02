package youtube

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/api/youtube/v3"

	"github.com/HalvaPovidlo/halvabot-go/internal/pkg/item"
)

const (
	videoPrefix     = "https://youtube.com/watch?v="
	channelPrefix   = "https://youtube.com/channel/"
	videoKind       = "youtube#video"
	videoFormat     = ".m4a"
	videoType       = "audio/mp4"
	maxSearchResult = 10
)

var (
	ErrSongNotFound = errors.New("song not found")
)

type loader interface {
	Download(ctx context.Context, id, outputDir string) (string, error)
}

type Config struct {
	Download  bool   `json:"download"`
	OutputDir string `json:"output"`
}

type YouTube struct {
	youtube *youtube.Service
	loader  loader
	config  Config
}

func NewYouTubeClient(yt *youtube.Service, loader loader, config Config) *YouTube {
	return &YouTube{
		youtube: yt,
		loader:  loader,
		config:  config,
	}
}

func getImages(details *youtube.ThumbnailDetails) (string, string) {
	artwork := ""
	thumbnail := ""
	if details != nil {
		if details.Default != nil {
			thumbnail = details.Default.Url
			artwork = details.Default.Url
		}
		if details.Standard != nil {
			thumbnail = details.Standard.Url
			artwork = details.Standard.Url
		}
		if details.Medium != nil {
			artwork = details.Medium.Url
		}
		if details.High != nil {
			artwork = details.High.Url
		}
		if details.Maxres != nil {
			artwork = details.Maxres.Url
		}
	}
	return artwork, thumbnail
}

//func getYTDLImages(ts ytdl.Thumbnails) (string, string) {
//	if len(ts) == 0 {
//		return "", ""
//	}
//	thumbnails := []ytdl.Thumbnail(ts)
//	var maxHeight uint
//	maxIter := 0
//	for i := range thumbnails {
//		t := &thumbnails[i]
//		if t.Height > maxHeight {
//			maxHeight = t.Height
//			maxIter = i
//		}
//	}
//	return thumbnails[maxIter].URL, thumbnails[maxIter].URL
//}

func (y *YouTube) findSong(ctx context.Context, query string) (*item.Song, error) {
	call := y.youtube.Search.List([]string{"id, snippet"}).Q(query).MaxResults(maxSearchResult).Context(ctx)

	response, err := call.Do()
	if err != nil || response.Items == nil {
		return nil, ErrSongNotFound
	}

	for _, resp := range response.Items {
		if resp.Id.Kind == videoKind {
			art, thumb := getImages(resp.Snippet.Thumbnails)
			return &item.Song{
				Title:        resp.Snippet.Title,
				URL:          videoPrefix + resp.Id.VideoId,
				Service:      item.ServiceYouTube,
				ArtistName:   resp.Snippet.ChannelTitle,
				ArtistURL:    channelPrefix + resp.Snippet.ChannelId,
				ArtworkURL:   art,
				ThumbnailURL: thumb,
				ID: item.SongID{
					ID:      resp.Id.VideoId,
					Service: item.ServiceYouTube,
				},
			}, nil
		}
	}

	return nil, ErrSongNotFound
}

func (y *YouTube) EnsureStreamInfo(ctx context.Context, song *item.Song) (*item.Song, error) {
	fileName, err := y.loader.Download(ctx, song.ID.ID, y.config.OutputDir)
	if err != nil {
		return nil, errors.Wrap(err, "download youtube song")
	}
	song.StreamURL = fileName
	return song, nil
}

//func songFromInfo(v *ytdl.Video) *item.Song {
//	art, thumb := getYTDLImages(v.Thumbnails)
//	return &item.Song{
//		Title:        v.Title,
//		URL:          videoPrefix + v.ID,
//		Service:      item.ServiceYouTube,
//		ArtistName:   v.Author,
//		ArtworkURL:   art,
//		ThumbnailURL: thumb,
//		ID: item.SongID{
//			ID:      v.ID,
//			Service: item.ServiceYouTube,
//		},
//		Duration: v.Duration.Seconds(),
//	}
//}

func (y *YouTube) FindSong(ctx context.Context, query string) (*item.Song, error) {
	song, err := y.findSong(ctx, query)
	if err != nil {
		return nil, err
	}

	song, err = y.EnsureStreamInfo(ctx, song)
	if err != nil {
		return nil, errors.Wrap(err, "ensure stream info")
	}
	return song, nil
}
