package youtube

import (
	"context"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

type filesCache interface {
	Add(path string)
	Remove(path string)
}

type dlp struct {
	loaded filesCache
}

func NewYouTubeDL(cache filesCache) *dlp {
	return &dlp{
		loaded: cache,
	}
}

func (dl *dlp) Download(ctx context.Context, id, outputDir string) (string, error) {
	ytdl := exec.CommandContext(ctx, "./yt-dlp",
		"-f", "ba[ext=m4a][abr<200]",
		"-q",
		"--print", "after_move:filepath",
		"-o", outputDir+"/%(id)s.%(ext)s",
		id)
	output, err := ytdl.Output()
	if err != nil {
		return "", errors.Wrap(err, "cmd yt-dlp")
	}
	path := strings.TrimSuffix(string(output), "\n")
	dl.loaded.Add(path)
	return path, nil
}
