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

	// MetadataEncryptionKeyId is "encryption-key-id"
	MetadataEncryptionKeyId = "encryption-key-id"
)

var CompressAndEncryptDataConverter = NewEncryptionDataConverter(
	converter.GetDefaultDataConverter(),
	EncryptionDataConverterOptions{Compress: true},
)

var EncryptDataConverter = NewEncryptionDataConverter(
	converter.GetDefaultDataConverter(),
	EncryptionDataConverterOptions{Compress: false},
)

type EncryptionDataConverter struct {
	// Until EncodingDataConverter supports workflow.ContextAware we'll store parent here.
	parent converter.DataConverter
	converter.EncodingDataConverter
	options EncryptionDataConverterOptions
}

type EncryptionDataConverterOptions struct {
	KeyID string
	// Enable ZLib compression before encryption.
	Compress bool
}

// Encoder implements PayloadEncoder using AES Crypt.
type Encoder struct {
	KeyID string
}

// TODO: Implement workflow.ContextAware in EncodingDataConverter
func (dc *EncryptionDataConverter) WithWorkflowContext(ctx workflow.Context) converter.DataConverter {
	if val := ctx.Value(PropagateKey); val != nil {
		parent := dc.parent
		if parentWithContext, ok := parent.(workflow.ContextAware); ok {
			parent = parentWithContext.WithWorkflowContext(ctx)
		}

		options := dc.options
		options.KeyID = val.(CryptContext).KeyId

		return NewEncryptionDataConverter(parent, options)
	}

	return dc
}

// TODO: Implement workflow.ContextAware in EncodingDataConverter
func (dc *EncryptionDataConverter) WithContext(ctx context.Context) converter.DataConverter {
	if val := ctx.Value(PropagateKey); val != nil {
		parent := dc.parent
		if parentWithContext, ok := parent.(workflow.ContextAware); ok {
			parent = parentWithContext.WithContext(ctx)
		}

		options := dc.options
		options.KeyID = val.(CryptContext).KeyId

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
func NewEncryptionDataConverter(dataConverter converter.DataConverter, options EncryptionDataConverterOptions) *EncryptionDataConverter {
	encoders := []converter.PayloadEncoder{
		&Encoder{KeyID: options.KeyID},
	}
	if options.Compress {
		encoders = append(encoders, converter.NewZlibEncoder(converter.ZlibEncoderOptions{AlwaysEncode: true}))
	}

	return &EncryptionDataConverter{
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
		MetadataEncryptionKeyId:    []byte(e.KeyID),
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

	keyId, ok := p.Metadata[MetadataEncryptionKeyId]
	if !ok {
		return fmt.Errorf("no encryption key id")
	}

	key := e.getKey(string(keyId))

	b, err := decrypt(p.Data, key)
	if err != nil {
		return err
	}

	p.Reset()
	err = p.Unmarshal(b)
	if err != nil {
		return err
	}

	return nil
}
