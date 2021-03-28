package cryptconverter

import (
	"fmt"

	commonpb "go.temporal.io/api/common/v1"

	"go.temporal.io/sdk/converter"
)

const (
	// MetadataEncryptionKeyId is "encryption-key-id"
	MetadataEncryptionKeyId = "encryption-key-id"

	// MetadataContentEncoding is "content-encoding"
	MetadataContentEncoding = "content-encoding"
)

// CryptDataConverter implements DataConverter using AES Crypt.
type CryptDataConverter struct {
	dataConverter converter.DataConverter
	context       CryptContext
}

type Context interface {
	Value(interface{}) interface{}
}

func (dc *CryptDataConverter) WithContext(c interface{}) converter.DataConverter {
	ctx, ok := c.(Context)
	if !ok {
		return dc
	}

	if val := ctx.Value(PropagateKey); val != nil {
		return &CryptDataConverter{
			dataConverter: dc.dataConverter,
			context:       val.(CryptContext),
		}
	}

	return dc
}

func (dc *CryptDataConverter) getKey(keyId string) (key []byte) {
	// Key can be fetched from KMS or other secure storage.
	return []byte("test-key-test-key-test-key-test!")
}

// NewCryptDataConverter created new instance of CryptDataConverter wrapping a DataConverter
func NewCryptDataConverter(dataConverter converter.DataConverter) *CryptDataConverter {
	return &CryptDataConverter{
		dataConverter: dataConverter,
	}
}

// ToPayloads converts a list of values.
func (dc *CryptDataConverter) ToPayloads(values ...interface{}) (*commonpb.Payloads, error) {
	result := &commonpb.Payloads{}

	for i, value := range values {
		payload, err := dc.ToPayload(value)
		if err != nil {
			return nil, fmt.Errorf("values[%d]: %w", i, err)
		}

		result.Payloads = append(result.Payloads, payload)
	}

	return result, nil
}

func (dc *CryptDataConverter) EncryptPayload(payload *commonpb.Payload, keyId string) error {
	key := dc.getKey(keyId)

	metadata := payload.GetMetadata()
	if metadata == nil {
		return converter.ErrMetadataIsNotSet
	}

	encoding, ok := metadata[converter.MetadataEncoding]
	if !ok {
		return converter.ErrEncodingIsNotSet
	}
	metadata[converter.MetadataEncoding] = []byte(converter.MetadataEncodingBinary)
	metadata[MetadataContentEncoding] = encoding
	metadata[MetadataEncryptionKeyId] = []byte(keyId)

	encryptedData, err := encrypt(payload.GetData(), key)
	if err != nil {
		return fmt.Errorf("%w: %v", converter.ErrUnableToEncode, err)
	}

	payload.Data = encryptedData

	return nil
}

// ToPayload converts single value to payload.
func (dc *CryptDataConverter) ToPayload(value interface{}) (*commonpb.Payload, error) {
	if dc.context.KeyId == "" {
		return dc.dataConverter.ToPayload(value)
	}

	payload, err := dc.dataConverter.ToPayload(value)
	if err != nil {
		return nil, err
	}

	if payload == nil {
		return payload, nil
	}

	err = dc.EncryptPayload(payload, dc.context.KeyId)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

// FromPayloads converts to a list of values of different types.
func (dc *CryptDataConverter) FromPayloads(payloads *commonpb.Payloads, valuePtrs ...interface{}) error {
	for i, payload := range payloads.GetPayloads() {
		err := dc.FromPayload(payload, valuePtrs[i])
		if err != nil {
			return fmt.Errorf("args[%d]: %w", i, err)
		}
	}

	return nil
}

func (dc *CryptDataConverter) DecryptPayload(payload *commonpb.Payload) error {
	metadata := payload.GetMetadata()
	if metadata == nil {
		return converter.ErrMetadataIsNotSet
	}

	keyId, ok := metadata[MetadataEncryptionKeyId]
	if !ok {
		return nil
	}

	encoding, ok := metadata[MetadataContentEncoding]
	if !ok {
		return fmt.Errorf("%w: %s", converter.ErrUnableToDecode, "no content encoding")
	}

	metadata[converter.MetadataEncoding] = encoding
	delete(metadata, MetadataContentEncoding)
	delete(metadata, MetadataEncryptionKeyId)

	key := dc.getKey(string(keyId))
	decryptData, err := decrypt(payload.GetData(), key)
	if err != nil {
		return fmt.Errorf("%w: %v", converter.ErrUnableToDecode, err)
	}

	payload.Data = decryptData

	return nil
}

// FromPayload converts single value from payload.
func (dc *CryptDataConverter) FromPayload(payload *commonpb.Payload, valuePtr interface{}) error {
	err := dc.DecryptPayload(payload)
	if err != nil {
		return err
	}

	return dc.dataConverter.FromPayload(payload, valuePtr)
}

// ToStrings converts payloads object into human readable strings.
func (dc *CryptDataConverter) ToStrings(payloads *commonpb.Payloads) []string {
	var result []string
	for _, payload := range payloads.GetPayloads() {
		result = append(result, dc.ToString(payload))
	}

	return result
}

// ToString converts payload object into human readable string.
func (dc *CryptDataConverter) ToString(payload *commonpb.Payload) string {
	err := dc.DecryptPayload(payload)
	if err != nil {
		return err.Error()
	}

	return dc.dataConverter.ToString(payload)
}
