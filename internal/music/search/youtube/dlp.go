package youtube

import (
	"context"

	"github.com/kkdai/youtube/v2"
)

type dlp struct {
	loaded filesCache
}

func NewYouTubeDLP(client youtube.Client, outputDir string, cache filesCache) *dlp {
	return &dlp{
		loaded: cache,
	}
}

func (dl *dlp) Download(ctx context.Context, v *youtube.Video, format *youtube.Format, outputFile string) error {

}
