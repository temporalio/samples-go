package contextawareencryption

import (
	"context"
	"fmt"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/workflow"
)

const (
	// MetadataEncodingEncrypted is "binary/encrypted"
	MetadataEncodingEncrypted = "binary/encrypted"

	// MetadataEncryptionKeyID is "encryption-key-id"
	MetadataEncryptionKeyID = "encryption-key-id"
	MetadataTenant          = "tenant"
)

type DataConverterOptions struct {
	KeyID string
	// Enable ZLib compression before encryption.
	Compress bool
}

type DataConverter struct {
	// Until EncodingDataConverter supports workflow.ContextAware we'll store parent here.
	parent converter.DataConverter
	converter.DataConverter
	options DataConverterOptions
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
		fmt.Println("dataConverter.WithWorkflowContext forwarding key:", val.KeyID)
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
		fmt.Println("dataConverter.WithContext forwarding key:", val.KeyID)

		return NewEncryptionDataConverter(parent, options)
	}

	return dc
}

// NewEncryptionDataConverter creates a new instance of EncryptionDataConverter wrapping a DataConverter
func NewEncryptionDataConverter(dataConverter converter.DataConverter, options DataConverterOptions) *DataConverter {
	codec := GetTenantCodec(options.KeyID)
	codecs := []converter.PayloadCodec{codec}
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
