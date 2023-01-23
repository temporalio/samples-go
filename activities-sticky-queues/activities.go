package activities_sticky_queues

import (
	"context"
	"crypto/sha256"
	"io"
	"os"
	"time"

	"go.temporal.io/sdk/activity"
)

// DownloadFile creates a file on the host with some test data.
func DownloadFile(ctx context.Context, url, path string) (err error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Downloading file", "url", url, "path", path)
	// Create the file
	out, err := os.Create(path)
	if err != nil {
		return
	}
	defer func() {
		err = out.Close()
	}()
	time.Sleep(3 * time.Second)
	_, err = out.WriteString("downloaded body")
	return
}

// ProcessFile is a stub function to processes a file.
func ProcessFile(ctx context.Context, path string) (err error) {
	logger := activity.GetLogger(ctx)
	// Open file
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer func() {
		err = f.Close()
	}()

	// Calculate checksum
	h := sha256.New()
	if _, err = io.Copy(h, f); err != nil {
		return
	}
	checksum := h.Sum(nil)

	logger.Info("Did some work", "path", path, "checksum", checksum)
	time.Sleep(3 * time.Second)
	return
}

// DeleteFile deletes path file on the host
func DeleteFile(ctx context.Context, path string) error {
	logger := activity.GetLogger(ctx)
	time.Sleep(3 * time.Second)
	logger.Info("Removing file", "path", path)
	return os.Remove(path)
}
