package youtube

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/kkdai/youtube/v2"
	"github.com/kkdai/youtube/v2/downloader"

	"github.com/HalvaPovidlo/discordBotGo/pkg/zap"
)

type Downloader struct {
	logger zap.Logger
	downloader.Downloader
}

func (dl *Downloader) Download(ctx context.Context, v *youtube.Video, format *youtube.Format, outputFile string) error {
	dl.logger.Infof("Video '%s'- Codec '%s'", v.Title, format.MimeType)
	destFile, err := dl.getOutputFile(outputFile)
	if err != nil {
		return err
	}

	out, err := os.Create(destFile)
	if err != nil {
		return err
	}
	defer out.Close()

	dl.logger.Infof("Download to file=%s", destFile)
	return dl.videoDLWorker(ctx, out, v, format)
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
