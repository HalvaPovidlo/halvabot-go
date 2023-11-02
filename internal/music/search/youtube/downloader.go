package youtube

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/kkdai/youtube/v2"

	"github.com/kkdai/youtube/v2/downloader"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/HalvaPovidlo/halvabot-go/pkg/contexts"
)

type Downloader struct {
	loaded filesCache
	ytdl   *youtube.Client
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

func (dl *Downloader) Download(ctx context.Context, id, outputDir string) (string, error) {
	videoInfo, err := dl.ytdl.GetVideo(id)
	if err != nil {
		return "", errors.Wrapf(err, "loag video metadata by url %s", id)
	}
	formats := videoInfo.Formats.WithAudioChannels().Type(videoType)
	if len(formats) == 0 {
		return "", errors.New("unable to get list of formats")
	}

	fileName := id + videoFormat
	dl.OutputDir = outputDir
	file, err := dl.getOutputFile(fileName)
	if err != nil {
		return "", err
	}

	logger := contexts.GetLogger(ctx)
	dl.loaded.Add(file)

	var out *os.File
	_, err = os.Stat(file)
	switch err {
	case nil:
		logger.Info("file is already downloaded", zap.String("name", file))
		return file, nil
	default:
		if errors.Is(err, os.ErrNotExist) {
			out, err = os.Create(file)
			if err != nil {
				dl.loaded.Remove(file)
				return "", err
			}
		} else {
			dl.loaded.Remove(file)
			return "", err
		}
	}
	defer out.Close()

	formats.Sort()
	format := formats[len(formats)-1]
	logger.Info("downloading video", zap.String("codec", format.MimeType), zap.String("filename", file))
	if err := dl.videoDLWorker(ctx, out, videoInfo, &format); err != nil {
		dl.loaded.Remove(file)
		return "", err
	}
	return file, nil
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
