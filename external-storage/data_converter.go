package externalstorage

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
)

// NewClient dials Temporal with the data converter and external storage
// configuration shared by every process in this sample (worker, starter,
// codec server). They must all agree on:
//
//   - the codec chain that wraps each payload (zlib here), and
//   - the external storage driver and threshold that decides when a payload
//     is offloaded instead of stored inline.
//
// If any one of them diverges, payloads written by one side will not be
// readable by the other.
func NewClient(ctx context.Context, options client.Options) (client.Client, error) {
	driver, err := NewS3Driver(ctx)
	if err != nil {
		return nil, fmt.Errorf("new s3 driver: %w", err)
	}

	if options.DataConverter == nil {
		options.DataConverter = converter.NewCodecDataConverter(
			converter.GetDefaultDataConverter(),
			converter.NewZlibCodec(converter.ZlibCodecOptions{AlwaysEncode: true}),
		)
	}
	options.ExternalStorage = converter.ExternalStorage{
		Drivers: []converter.StorageDriver{driver},
	}

	return client.Dial(options)
}
