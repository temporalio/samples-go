package cryptconverter

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	commonpb "go.temporal.io/api/common/v1"

	"go.temporal.io/sdk/converter"
)

const (
	// MetadataWrappedEncoding is "wrapped-encoding"
	MetadataWrappedEncoding = "wrapped-encoding"
)

// CryptDataConverter implements DataConverter using AES Crypt.
type CryptDataConverter struct {
	dataConverter converter.DataConverter
}

// getKey fetches the crypt key from secure storage
func (dc *CryptDataConverter) getKey() (key []byte) {
	// Key can be fetched from KMS or other secure storage.
	return []byte("test-key-test-key-test-key-test!")
}

// NewCryptDataConverter created new instance of CryptDataConverter wrapping a DataConverter
func NewCryptDataConverter(dataConverter converter.DataConverter) *CryptDataConverter {
	return &CryptDataConverter{
		dataConverter: dataConverter,
	}
}

func encrypt(plainData []byte, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plainData, nil), nil
}

func decrypt(encryptedData []byte, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short: %v", encryptedData)
	}

	nonce, encryptedData := encryptedData[:nonceSize], encryptedData[nonceSize:]
	return gcm.Open(nil, nonce, encryptedData, nil)
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

// ToPayload converts single value to payload.
func (dc *CryptDataConverter) ToPayload(value interface{}) (*commonpb.Payload, error) {
	payload, err := dc.dataConverter.ToPayload(value)
	if err != nil {
		return nil, err
	}

	if payload != nil {
		metadata := payload.GetMetadata()
		if metadata == nil {
			return nil, converter.ErrMetadataIsNotSet
		}

		encoding := metadata[converter.MetadataEncoding]
		if encoding != nil {
			metadata[MetadataWrappedEncoding] = encoding
		}
		metadata[converter.MetadataEncoding] = []byte("binary/crypt")

		encryptedData, err := encrypt(payload.GetData(), dc.getKey())
		if err != nil {
			return nil, fmt.Errorf("%w: %v", converter.ErrUnableToEncode, err)
		}

		payload.Data = encryptedData
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

// FromPayload converts single value from payload.
func (dc *CryptDataConverter) FromPayload(payload *commonpb.Payload, valuePtr interface{}) error {
	decryptData, err := decrypt(payload.GetData(), dc.getKey())
	if err != nil {
		return fmt.Errorf("%w: %v", converter.ErrUnableToDecode, err)
	}

	metadata := payload.GetMetadata()
	if metadata == nil {
		return converter.ErrMetadataIsNotSet
	}

	encoding := metadata[MetadataWrappedEncoding]
	if encoding != nil {
		metadata[converter.MetadataEncoding] = encoding
		delete(metadata, MetadataWrappedEncoding)
	}

	payload.Data = decryptData

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
	decryptData, err := decrypt(payload.GetData(), dc.getKey())
	// No way to return an error here...
	if err != nil {
		return err.Error()
	}
	payload.Data = decryptData

	return dc.dataConverter.ToString(payload)
}
