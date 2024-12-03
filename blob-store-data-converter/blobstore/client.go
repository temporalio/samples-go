package blobstore

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Client struct {
	dir                    string
	simulateNetworkLatency time.Duration
}

func NewClient() *Client {
	return &Client{
		dir:                    "/tmp/temporal-sample/blob-store-data-converter/blobs",
		simulateNetworkLatency: 1 * time.Second,
	}
}

func NewTestClient() *Client {
	return &Client{
		dir:                    "/tmp/temporal-sample/blob-store-data-converter/test-blobs",
		simulateNetworkLatency: 0,
	}
}

func (b *Client) SaveBlob(key string, data []byte) error {
	err := os.MkdirAll(b.dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", b.dir, err)
	}

	path := fmt.Sprintf(b.dir + "/" + strings.ReplaceAll(key, "/", "_"))
	fmt.Println("saving blob to: ", path)
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to save blob: %w", err)
	}

	time.Sleep(b.simulateNetworkLatency)

	return nil
}

func (b *Client) GetBlob(key string) ([]byte, error) {
	path := fmt.Sprintf(b.dir + "/" + strings.ReplaceAll(key, "/", "_"))
	fmt.Println("reading blob from: ", path)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read blob: %w", err)
	}

	time.Sleep(b.simulateNetworkLatency)

	return data, nil
}
