package youtube

import (
	"context"
	"path/filepath"
	"sort"

	ytdl "github.com/kkdai/youtube/v2"
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
	Download(ctx context.Context, v *ytdl.Video, format *ytdl.Format, outputFile string) error
}

type Config struct {
	Download  bool   `json:"download"`
	OutputDir string `json:"output"`
}

type YouTube struct {
	ytdl    *ytdl.Client
	youtube *youtube.Service
	loader  loader
	config  Config
}

func NewYouTubeClient(ytdl *ytdl.Client, yt *youtube.Service, loader *Downloader, config Config) *YouTube {
	return &YouTube{
		ytdl:    ytdl,
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

func getYTDLImages(ts ytdl.Thumbnails) (string, string) {
	if len(ts) == 0 {
		return "", ""
	}
	thumbnails := []ytdl.Thumbnail(ts)
	var maxHeight uint
	maxIter := 0
	for i := range thumbnails {
		t := &thumbnails[i]
		if t.Height > maxHeight {
			maxHeight = t.Height
			maxIter = i
		}
	}
	return thumbnails[maxIter].URL, thumbnails[maxIter].URL
}

func (y *YouTube) findSong(ctx context.Context, query string) (*item.Song, error) {
	y.youtube.Videos.List()
	call := y.youtube.Search.List([]string{"id, snippet"}).
		Q(query).
		MaxResults(maxSearchResult)
	call.Context(ctx)
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
	videoInfo, err := y.ytdl.GetVideo(song.URL)
	if err != nil {
		return nil, errors.Wrapf(err, "loag video metadata by url %s", song.URL)
	}
	formats := videoInfo.Formats.WithAudioChannels().Type(videoType)
	if len(formats) == 0 {
		return nil, errors.New("unable to get list of formats")
	}

	if y.config.Download {
		formats.Sort()
		format := formats[len(formats)-1]
		fileName := videoInfo.ID + videoFormat
		song.StreamURL = filepath.Join(y.config.OutputDir, fileName)
		err := y.loader.Download(ctx, videoInfo, &format, fileName)
		if err != nil {
			return nil, err
		}
	} else {
		sort.SliceStable(formats, func(i, j int) bool {
			return formats[i].ItagNo < formats[j].ItagNo
		})
		format := formats[0]
		streamURL, err := y.ytdl.GetStreamURLContext(ctx, videoInfo, &format)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to get streamURL %s", videoInfo.Title)
		}
		song.StreamURL = streamURL
	}

	additionalSongInfo := songFromInfo(videoInfo)
	song.MergeNoOverride(additionalSongInfo)
	return song, nil
}

func songFromInfo(v *ytdl.Video) *item.Song {
	art, thumb := getYTDLImages(v.Thumbnails)
	return &item.Song{
		Title:        v.Title,
		URL:          videoPrefix + v.ID,
		Service:      item.ServiceYouTube,
		ArtistName:   v.Author,
		ArtworkURL:   art,
		ThumbnailURL: thumb,
		ID: item.SongID{
			ID:      v.ID,
			Service: item.ServiceYouTube,
		},
		Duration: v.Duration.Seconds(),
	}
}

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
