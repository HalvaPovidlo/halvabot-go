package youtube

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/HalvaPovidlo/discordBotGo/pkg/contexts"
	"github.com/kkdai/youtube/v2"
	"github.com/kkdai/youtube/v2/downloader"
	"go.uber.org/zap"
)

type filesCache interface {
	Add(path string)
	Remove(path string)
}

type Downloader struct {
	loaded filesCache
	downloader.Downloader
}

func NewDownloader(client youtube.Client, outputDir string, cache filesCache) *Downloader {
	return &Downloader{
		loaded: cache,
		Downloader: downloader.Downloader{
			Client:    client,
			OutputDir: outputDir,
		},
	}
}

func (dl *Downloader) Download(ctx context.Context, v *youtube.Video, format *youtube.Format, outputFile string) error {
	destinationFile, err := dl.getOutputFile(outputFile)
	if err != nil {
		return err
	}
	logger := contexts.GetLogger(ctx)
	dl.loaded.Add(destinationFile)

	var out *os.File
	_, err = os.Stat(destinationFile)
	switch err {
	case nil:
		logger.Info("file is already downloaded", zap.String("name", destinationFile))
		return nil
	case os.ErrNotExist:
		out, err = os.Create(destinationFile)
		if err != nil {
			dl.loaded.Remove(destinationFile)
			return err
		}
	default:
		dl.loaded.Remove(destinationFile)
		return err
	}
	defer out.Close()

	logger.Info("downloading video",
		zap.String("title", v.Title),
		zap.String("codec", format.MimeType),
		zap.String("filename", destinationFile))
	if err := dl.videoDLWorker(ctx, out, v, format); err != nil {
		dl.loaded.Remove(destinationFile)
		return err
	}
	return nil
}

func (dl *Downloader) getOutputFile(outputFile string) (string, error) {
	if dl.OutputDir != "" {
		if err := os.MkdirAll(dl.OutputDir, 0o755); err != nil {
			return "", err
		}
		outputFile = filepath.Join(dl.OutputDir, outputFile)
	}
	return outputFile, nil
}

func (dl *Downloader) videoDLWorker(ctx context.Context, out *os.File, video *youtube.Video, format *youtube.Format) error {
	stream, size, err := dl.GetStreamContext(ctx, video, format)
	if err != nil {
		return err
	}

	reader := stream
	prog := &progress{contentLength: float64(size)}
	mw := io.MultiWriter(out, prog)
	_, err = io.Copy(mw, reader)
	if err != nil {
		return err
	}
	return nil
}

type progress struct {
	contentLength     float64
	totalWrittenBytes float64
	downloadLevel     float64
}

func (dl *progress) Write(p []byte) (n int, err error) {
	n = len(p)
	dl.totalWrittenBytes += float64(n)
	currentPercent := (dl.totalWrittenBytes / dl.contentLength) * 100
	if (dl.downloadLevel <= currentPercent) && (dl.downloadLevel < 100) {
		dl.downloadLevel++
	}
	return
}
