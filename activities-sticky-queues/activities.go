package activities_sticky_queues

import (
	"bytes"
	"context"
	"crypto/sha256"
	"io"
	"os"
	"time"

	"go.temporal.io/sdk/activity"
)

// DownloadFile creates a file on the host with some test data.
func DownloadFile(ctx context.Context, url, path string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Downloading file", "url", url, "path", path)
	time.Sleep(3 * time.Second)
	return os.WriteFile(url, []byte("Hello, Gophers!"), 0666)
}

// ProcessFile is a stub function to processes a file.
func ProcessFile(ctx context.Context, path string) error {
	logger := activity.GetLogger(ctx)
	// Read from file
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Calculate checksum
	h := sha256.New()
	if _, err = io.Copy(h, bytes.NewReader(data)); err != nil {
		return err
	}
	checksum := h.Sum(nil)

	logger.Info("Did some work", "path", path, "checksum", checksum)
	time.Sleep(3 * time.Second)
	return nil
}

// DeleteFile deletes path file on the host
func DeleteFile(ctx context.Context, path string) error {
	logger := activity.GetLogger(ctx)
	time.Sleep(3 * time.Second)
	logger.Info("Removing file", "path", path)
	return os.Remove(path)
}
