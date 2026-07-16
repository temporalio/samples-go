package externalstorage

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
)

// NewSampleDataConverter is the single source of truth for the codec chain applied to
// every payload in this sample. The worker, starter, and codec server must all
// agree on it.
func NewSampleDataConverter() converter.DataConverter {
	return converter.NewCodecDataConverter(
		converter.GetDefaultDataConverter(),
		converter.NewZlibCodec(converter.ZlibCodecOptions{AlwaysEncode: true}),
	)
}

// NewClient dials Temporal with the data converter and external storage
// configuration shared by every process in this sample.
func NewClient(ctx context.Context, options client.Options) (client.Client, error) {
	driver, err := NewS3Driver(ctx)
	if err != nil {
		return nil, fmt.Errorf("new s3 driver: %w", err)
	}

	if options.DataConverter == nil {
		options.DataConverter = NewSampleDataConverter()
	}
	options.ExternalStorage = converter.ExternalStorage{
		Drivers: []converter.StorageDriver{driver},
	}

	return client.Dial(options)
}
