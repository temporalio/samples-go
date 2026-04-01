package snappycompress

import (
	"github.com/golang/snappy"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
	"google.golang.org/protobuf/proto"
)

// AlwaysCompressDataConverter is a converter that will always perform
// compression even if the compressed size is larger than the original.
var AlwaysCompressDataConverter = NewDataConverter(converter.GetDefaultDataConverter(), Options{AlwaysEncode: true})

// Options are options for Snappy compression.
type Options struct {
	// If true, will always "compress" even if the compression results in a larger
	// sized payload.
	AlwaysEncode bool
}

// NewDataConverter creates a new data converter that wraps the given data
// converter with snappy compression.
func NewDataConverter(underlying converter.DataConverter, options Options) converter.DataConverter {
	return converter.NewCodecDataConverter(underlying, &Encoder{Options: options})
}

// Encoder implements converter.PayloadCodec for snappy compression.
type Encoder struct {
	Options Options
}

// Encode implements converter.PayloadCodec.Encode.
func (e *Encoder) Encode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	result := make([]*commonpb.Payload, len(payloads))
	for i, p := range payloads {
		// Marshal proto
		origBytes, err := proto.Marshal(p)
		if err != nil {
			return payloads, err
		}
		// Compress
		b := snappy.Encode(nil, origBytes)
		// Only apply if the compression is smaller or always encode is set
		if len(b) < len(origBytes) || e.Options.AlwaysEncode {
			result[i] = &commonpb.Payload{
				Metadata: map[string][]byte{converter.MetadataEncoding: []byte("binary/snappy")},
				Data:     b,
			}
		} else {
			result[i] = p
		}
	}
	return result, nil
}

// Decode implements converter.PayloadCodec.Decode.
func (*Encoder) Decode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	result := make([]*commonpb.Payload, len(payloads))
	for i, p := range payloads {
		// Only if it's our encoding
		if string(p.Metadata[converter.MetadataEncoding]) != "binary/snappy" {
			result[i] = p
			continue
		}
		// Uncompress
		b, err := snappy.Decode(nil, p.Data)
		if err != nil {
			return payloads, err
		}
		// Unmarshal proto
		result[i] = &commonpb.Payload{}
		if err := proto.Unmarshal(b, result[i]); err != nil {
			return payloads, err
		}
	}
	return result, nil
}
