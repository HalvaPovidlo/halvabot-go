package search

import (
	"regexp"
	"sort"

	iso "github.com/channelmeter/iso8601duration"
	ytdl "github.com/kkdai/youtube/v2"
	"github.com/pkg/errors"
	"google.golang.org/api/youtube/v3"

	"github.com/HalvaPovidlo/discordBotGo/internal/discord/voice"
)

const videoPrefix = "https://youtube.com/watch?v="

// YouTube exports the methods required to access the YouTube service
type YouTube struct {
	ytdl    *ytdl.Client
	youtube *youtube.Service
}

func NewYouTubeClient(ytdl *ytdl.Client, yt *youtube.Service) *YouTube {
	return &YouTube{
		ytdl:    ytdl,
		youtube: yt,
	}
}

// GetName returns the service's name
func (*YouTube) GetName() string {
	return "YouTube"
}

// GetColor returns the service's color
func (*YouTube) GetColor() int {
	return 0xFF0000
}

// TestURL tests if the given URL is a YouTube video URL
func (*YouTube) TestURL(url string) (bool, error) {
	test, err := regexp.MatchString("(?:https?://)?(?:www\\.)?youtu\\.?be(?:\\.com)?/?.*(?:watch|embed)?(?:.*v=|v/|/)[\\w-_]+", url)
	return test, err
}

// GetMetadata returns the metadata for a given YouTube video URL
func (y *YouTube) GetMetadata(url string) (*voice.Metadata, error) {
	videoInfo, err := y.ytdl.GetVideo(url)
	if err != nil {
		return nil, err
	}

	formats := videoInfo.Formats
	if len(formats) == 0 {
		return nil, errors.New("unable to get list of formats")
	}

	sort.SliceStable(formats, func(i, j int) bool {
		return formats[i].ItagNo < formats[j].ItagNo
	})

	format := formats[0]

	videoURL, err := y.ytdl.GetStreamURL(videoInfo, &format)
	if err != nil {
		return nil, err
	}

	ytCall := youtube.NewVideosService(y.youtube).
		List([]string{"snippet", "contentDetails"}).
		Id(videoInfo.ID)

	ytResponse, err := ytCall.Do()
	if err != nil {
		return nil, err
	}

	duration, err := iso.FromString(ytResponse.Items[0].ContentDetails.Duration)
	if err != nil {
		return nil, err
	}

	metadata := &voice.Metadata{
		Title:      videoInfo.Title,
		DisplayURL: url,
		StreamURL:  videoURL,
		Duration:   duration.ToDuration().Seconds(),
	}

	videoAuthor := &voice.MetadataArtist{
		Name: ytResponse.Items[0].Snippet.ChannelTitle,
		URL:  "https://youtube.com/channel/" + ytResponse.Items[0].Snippet.ChannelId,
	}
	metadata.Artists = append(metadata.Artists, *videoAuthor)

	return metadata, nil
}

// getQuery returns YouTube search results
func (y *YouTube) getQuery(query string) (string, error) {
	call := y.youtube.Search.List([]string{"id"}).
		Q(query).
		MaxResults(50)

	response, err := call.Do()
	if err != nil {
		return "", errors.Wrap(err, "could not find any results for the specified query")
	}

	for _, item := range response.Items {
		if item.Id.Kind == "youtube#video" {
			url := videoPrefix + item.Id.VideoId
			return url, nil
		}
	}

	return "", errors.New("could not find a video result for the specified query")
}

func (y *YouTube) FindSong(query string) (*voice.QueueEntry, error) {
	url, err := y.getQuery(query)
	if err != nil {
		return nil, err
	}

	test, err := y.TestURL(url)
	if err != nil {
		return nil, err
	}
	if test {
		metadata, err := y.GetMetadata(url)
		if err != nil {
			return nil, errors.Wrap(err, "Find song")
		}
		queueEntry := &voice.QueueEntry{
			Metadata:     metadata,
			ServiceName:  y.GetName(),
			ServiceColor: y.GetColor(),
		}
		return queueEntry, nil
	}

	return nil, errors.New("wrong youtube url")
}
