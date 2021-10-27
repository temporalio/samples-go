package snappycompress

import (
	"github.com/golang/snappy"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
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
	return converter.NewEncodingDataConverter(underlying, &Encoder{Options: options})
}

// Encoder implements converter.PayloadEncoder for snappy compression.
type Encoder struct {
	Options Options
}

// Encode implements converter.PayloadEncoder.Encode.
func (e *Encoder) Encode(p *commonpb.Payload) error {
	// Marshal proto
	origBytes, err := p.Marshal()
	if err != nil {
		return err
	}
	// Compress
	b := snappy.Encode(nil, origBytes)
	// Only apply if the compression is smaller or always encode is set
	if len(b) < len(origBytes) || e.Options.AlwaysEncode {
		p.Metadata = map[string][]byte{converter.MetadataEncoding: []byte("binary/snappy")}
		p.Data = b
	}
	return nil
}

// Decode implements converter.PayloadEncoder.Decode.
func (*Encoder) Decode(p *commonpb.Payload) error {
	// Only if it's our encoding
	if string(p.Metadata[converter.MetadataEncoding]) != "binary/snappy" {
		return nil
	}
	// Uncompress
	b, err := snappy.Decode(nil, p.Data)
	if err != nil {
		return err
	}
	// Unmarshal proto
	p.Reset()
	return p.Unmarshal(b)
}
