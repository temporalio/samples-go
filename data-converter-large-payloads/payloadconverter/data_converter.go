package dataconverter

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
)

const (
	MB                        = 1_000_000
	MetadataEncodingLocalFile = "example/local_file"
)

// LargeSizePayloadConverter is a payload converter that detects if incoming
// payloads are larger than a configurable threshold and if so, stores them to
// a file instead of passing them directly to the Temporal server. The threshold
// should be configured lower than the limit imposed by Temporal, which is 2MB.
//
// Note that other storage systems such as Postgres or S3 are a much better
// choice than local files in scenarios other than development and testing.
type LargeSizePayloadConverter struct {
	// used for generating names of files
	idGenerator idGenerator
	// used to detect which payloads must be stored in files
	threshold int
}

// idGenerator is used to generate the names of the files created for storing
// the large payloads.
type idGenerator interface {
	NewString() string
}

// LargeSizePayloadConverterOption specifies configuration options for the
// payload converter.
type LargeSizePayloadConverterOption func(*LargeSizePayloadConverter)

// WithThreshold sets the threshold from which payloads start to get written to
// files. Payloads are stored as JSONs in files and the raw []byte JSON payload
// length is compared to the threshold.
func WithThreshold(sizeBytes int) LargeSizePayloadConverterOption {
	return func(lspc *LargeSizePayloadConverter) {
		lspc.threshold = sizeBytes
	}
}

// NewLargeSizePayloadConverter creates new instance of
// LargeSizePayloadConverter.
//
// Do not use this payload converter on its own! Because the converter only
// applies to payloads larger than the threshold, there must be a fallback
// converter for payloads whose size is lower than the threshold. Always use
// this inside a composite payload converter.
func NewLargeSizePayloadConverter(opts ...LargeSizePayloadConverterOption) *LargeSizePayloadConverter {
	c := LargeSizePayloadConverter{
		idGenerator: uuidv4IDGenerator{},
		threshold:   1 * MB,
	}

	for _, opt := range opts {
		opt(&c)
	}

	return &c
}

// ToPayload implements converter.PayloadConverter.
func (c *LargeSizePayloadConverter) ToPayload(value any) (*commonpb.Payload, error) {

	result, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal json: %w", err)
	}

	// will fallback on other registered converters
	if len(result) < c.threshold {
		return nil, nil
	}

	converter.NewJSONPayloadConverter()

	filename := c.idGenerator.NewString()
	err = os.WriteFile(filename, result, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write payload to file: %w", err)
	}

	return &commonpb.Payload{
		Metadata: map[string][]byte{
			converter.MetadataEncoding: []byte(MetadataEncodingLocalFile),
		},
		Data: []byte(filename),
	}, nil
}

// FromPayload implements converter.PayloadConverter.
func (c *LargeSizePayloadConverter) FromPayload(payload *commonpb.Payload, valuePtr any) error {

	// This only gets called for payloads where the encoding is
	// "example/local_file", in which case the file name is stored in the data.
	filename := string(payload.Data)
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read payload from file: %w", err)
	}

	err = json.Unmarshal(data, valuePtr)
	if err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return nil
}

// ToString implements converter.PayloadConverter.
func (c *LargeSizePayloadConverter) ToString(payload *commonpb.Payload) string {
	// same as converter.JSONPayloadConverter
	return string(payload.GetData())
}

// Encoding implements converter.PayloadConverter.
func (c *LargeSizePayloadConverter) Encoding() string {
	return MetadataEncodingLocalFile
}

var _ converter.PayloadConverter = &LargeSizePayloadConverter{}

type uuidv4IDGenerator struct{}

func (u uuidv4IDGenerator) NewString() string {
	return uuid.NewString()
}
