package encryption

import (
	"context"
	"fmt"

	commonpb "go.temporal.io/api/common/v1"

	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/workflow"
)

const (
	// MetadataEncodingEncrypted is "binary/encrypted"
	MetadataEncodingEncrypted = "binary/encrypted"

	// MetadataEncryptionKeyID is "encryption-key-id"
	MetadataEncryptionKeyID = "encryption-key-id"
)

type DataConverter struct {
	// Until EncodingDataConverter supports workflow.ContextAware we'll store parent here.
	parent converter.DataConverter
	converter.DataConverter
	options DataConverterOptions
}

type DataConverterOptions struct {
	KeyID string
	// Enable ZLib compression before encryption.
	Compress bool
}

// Codec implements PayloadCodec using AES Crypt.
type Codec struct {
	KeyID string
}

// TODO: Implement workflow.ContextAware in CodecDataConverter
// Note that you only need to implement this function if you need to vary the encryption KeyID per workflow.
func (dc *DataConverter) WithWorkflowContext(ctx workflow.Context) converter.DataConverter {
	if val, ok := ctx.Value(PropagateKey).(CryptContext); ok {
		parent := dc.parent
		if parentWithContext, ok := parent.(workflow.ContextAware); ok {
			parent = parentWithContext.WithWorkflowContext(ctx)
		}

		options := dc.options
		options.KeyID = val.KeyID

		return NewEncryptionDataConverter(parent, options)
	}

	return dc
}

// TODO: Implement workflow.ContextAware in EncodingDataConverter
// Note that you only need to implement this function if you need to vary the encryption KeyID per workflow.
func (dc *DataConverter) WithContext(ctx context.Context) converter.DataConverter {
	if val, ok := ctx.Value(PropagateKey).(CryptContext); ok {
		parent := dc.parent
		if parentWithContext, ok := parent.(workflow.ContextAware); ok {
			parent = parentWithContext.WithContext(ctx)
		}

		options := dc.options
		options.KeyID = val.KeyID

		return NewEncryptionDataConverter(parent, options)
	}

	return dc
}

func (e *Codec) getKey(keyID string) (key []byte) {
	// Key must be fetched from secure storage in production (such as a KMS).
	// For testing here we just hard code a key.
	return []byte("test-key-test-key-test-key-test!")
}

// NewEncryptionDataConverter creates a new instance of EncryptionDataConverter wrapping a DataConverter
func NewEncryptionDataConverter(dataConverter converter.DataConverter, options DataConverterOptions) *DataConverter {
	codecs := []converter.PayloadCodec{
		&Codec{KeyID: options.KeyID},
	}
	// Enable compression if requested.
	// Note that this must be done before encryption to provide any value. Encrypted data should by design not compress very well.
	// This means the compression codec must come after the encryption codec here as codecs are applied last -> first.
	if options.Compress {
		codecs = append(codecs, converter.NewZlibCodec(converter.ZlibCodecOptions{AlwaysEncode: true}))
	}

	return &DataConverter{
		parent:        dataConverter,
		DataConverter: converter.NewCodecDataConverter(dataConverter, codecs...),
		options:       options,
	}
}

// Encode implements converter.PayloadCodec.Encode.
func (e *Codec) Encode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	result := make([]*commonpb.Payload, len(payloads))
	for i, p := range payloads {
		origBytes, err := p.Marshal()
		if err != nil {
			return payloads, err
		}

		key := e.getKey(e.KeyID)

		b, err := encrypt(origBytes, key)
		if err != nil {
			return payloads, err
		}

		result[i] = &commonpb.Payload{
			Metadata: map[string][]byte{
				converter.MetadataEncoding: []byte(MetadataEncodingEncrypted),
				MetadataEncryptionKeyID:    []byte(e.KeyID),
			},
			Data: b,
		}
	}

	return result, nil
}

// Decode implements converter.PayloadCodec.Decode.
func (e *Codec) Decode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	result := make([]*commonpb.Payload, len(payloads))
	for i, p := range payloads {
		// Only if it's encrypted
		if string(p.Metadata[converter.MetadataEncoding]) != MetadataEncodingEncrypted {
			result[i] = p
			continue
		}

		keyID, ok := p.Metadata[MetadataEncryptionKeyID]
		if !ok {
			return payloads, fmt.Errorf("no encryption key id")
		}

		key := e.getKey(string(keyID))

		b, err := decrypt(p.Data, key)
		if err != nil {
			return payloads, err
		}

		result[i] = &commonpb.Payload{}
		err = result[i].Unmarshal(b)
		if err != nil {
			return payloads, err
		}
	}

	return result, nil
}
