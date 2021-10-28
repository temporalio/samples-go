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

var CompressAndEncryptDataConverter = NewEncryptionDataConverter(
	converter.GetDefaultDataConverter(),
	DataConverterOptions{Compress: true},
)

var EncryptDataConverter = NewEncryptionDataConverter(
	converter.GetDefaultDataConverter(),
	DataConverterOptions{Compress: false},
)

type DataConverter struct {
	// Until EncodingDataConverter supports workflow.ContextAware we'll store parent here.
	parent converter.DataConverter
	converter.EncodingDataConverter
	options DataConverterOptions
}

type DataConverterOptions struct {
	KeyID string
	// Enable ZLib compression before encryption.
	Compress bool
}

// Encoder implements PayloadEncoder using AES Crypt.
type Encoder struct {
	KeyID string
}

// TODO: Implement workflow.ContextAware in EncodingDataConverter
func (dc *DataConverter) WithWorkflowContext(ctx workflow.Context) converter.DataConverter {
	if val, ok := ctx.Value(PropagateKey).(CryptContext); ok {
		parent := dc.parent
		if parentWithContext, ok := parent.(workflow.ContextAware); ok {
			parent = parentWithContext.WithWorkflowContext(ctx)
		}

		options := dc.options
		options.KeyID = val.KeyId

		return NewEncryptionDataConverter(parent, options)
	}

	return dc
}

// TODO: Implement workflow.ContextAware in EncodingDataConverter
func (dc *DataConverter) WithContext(ctx context.Context) converter.DataConverter {
	if val, ok := ctx.Value(PropagateKey).(CryptContext); ok {
		parent := dc.parent
		if parentWithContext, ok := parent.(workflow.ContextAware); ok {
			parent = parentWithContext.WithContext(ctx)
		}

		options := dc.options
		options.KeyID = val.KeyId

		return NewEncryptionDataConverter(parent, options)
	}

	return dc
}

func (e *Encoder) getKey(keyId string) (key []byte) {
	// Key must be fetched from secure storage in production (such as a KMS).
	// For testing here we just hard code a key.
	return []byte("test-key-test-key-test-key-test!")
}

// NewEncryptionDataConverter creates a new instance of EncryptionDataConverter wrapping a DataConverter
func NewEncryptionDataConverter(dataConverter converter.DataConverter, options DataConverterOptions) *DataConverter {
	encoders := []converter.PayloadEncoder{
		&Encoder{KeyID: options.KeyID},
	}
	// Enable compression if requested.
	// Note that this must be done before encryption to provide any value. Encrypted data should by design not compress very well.
	// This means the compression encoder must come after the encryption encoder here as encoders are applied last -> first.
	if options.Compress {
		encoders = append(encoders, converter.NewZlibEncoder(converter.ZlibEncoderOptions{AlwaysEncode: true}))
	}

	return &DataConverter{
		parent:                dataConverter,
		EncodingDataConverter: *converter.NewEncodingDataConverter(dataConverter, encoders...),
		options:               options,
	}
}

// Encode implements converter.PayloadEncoder.Encode.
func (e *Encoder) Encode(p *commonpb.Payload) error {
	origBytes, err := p.Marshal()
	if err != nil {
		return err
	}

	key := e.getKey(e.KeyID)

	b, err := encrypt(origBytes, key)
	if err != nil {
		return err
	}

	p.Metadata = map[string][]byte{
		converter.MetadataEncoding: []byte(MetadataEncodingEncrypted),
		MetadataEncryptionKeyID:    []byte(e.KeyID),
	}
	p.Data = b

	return nil
}

// Decode implements converter.PayloadEncoder.Decode.
func (e *Encoder) Decode(p *commonpb.Payload) error {
	// Only if it's encrypted
	if string(p.Metadata[converter.MetadataEncoding]) != MetadataEncodingEncrypted {
		return nil
	}

	keyID, ok := p.Metadata[MetadataEncryptionKeyID]
	if !ok {
		return fmt.Errorf("no encryption key id")
	}

	key := e.getKey(string(keyID))

	b, err := decrypt(p.Data, key)
	if err != nil {
		return err
	}

	p.Reset()
	return p.Unmarshal(b)
}
