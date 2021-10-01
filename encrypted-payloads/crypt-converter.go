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
	// MetadataEncodingEncrypted is "binary/encrypted"
	MetadataEncodingEncrypted = "binary/encrypted"
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

func encrypt(plaintext []byte, key []byte) ([]byte, error) {
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

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func (dc *CryptDataConverter) encryptPayload(payload *commonpb.Payload, key []byte) (*commonpb.Payload, error) {
	serializedPayload, err := payload.Marshal()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", converter.ErrUnableToEncode, err)
	}

	encryptedPayload, err := encrypt(serializedPayload, key)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", converter.ErrUnableToEncode, err)
	}

	return &commonpb.Payload{
		Metadata: map[string][]byte{
			converter.MetadataEncoding: []byte(MetadataEncodingEncrypted),
		},
		Data: encryptedPayload,
	}, nil
}

func isEncryptedPayload(payload *commonpb.Payload) bool {
	metadata := payload.GetMetadata()
	if metadata == nil {
		return false
	}

	if encoding, ok := metadata[converter.MetadataEncoding]; ok {
		return string(encoding) == MetadataEncodingEncrypted
	}

	return false
}

func decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short: %v", ciphertext)
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func (dc *CryptDataConverter) decryptPayload(payload *commonpb.Payload) (*commonpb.Payload, error) {
	serializedPayload, err := decrypt(payload.GetData(), dc.getKey())
	if err != nil {
		return nil, fmt.Errorf("%w: %v", converter.ErrUnableToDecode, err)
	}

	result := &commonpb.Payload{}
	err = result.Unmarshal(serializedPayload)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", converter.ErrUnableToDecode, err)
	}

	return result, nil
}

// ToPayloads converts a list of values.
func (dc *CryptDataConverter) ToPayloads(values ...interface{}) (*commonpb.Payloads, error) {
	payloads := &commonpb.Payloads{}

	for i, value := range values {
		payload, err := dc.ToPayload(value)
		if err != nil {
			return nil, fmt.Errorf("values[%d]: %w", i, err)
		}

		encryptedPayload, err := dc.encryptPayload(payload, dc.getKey())
		if err != nil {
			return nil, fmt.Errorf("values[%d]: %w", i, err)
		}

		payloads.Payloads = append(payloads.Payloads, encryptedPayload)
	}

	return payloads, nil
}

// ToPayload converts single value to payload.
func (dc *CryptDataConverter) ToPayload(value interface{}) (*commonpb.Payload, error) {
	return dc.dataConverter.ToPayload(value)
}

// FromPayloads converts to a list of values of different types.
func (dc *CryptDataConverter) FromPayloads(payloads *commonpb.Payloads, valuePtrs ...interface{}) error {
	for i, payload := range payloads.GetPayloads() {
		if i >= len(valuePtrs) {
			break
		}

		var err error

		if isEncryptedPayload(payload) {
			payload, err = dc.decryptPayload(payload)
			if err != nil {
				return fmt.Errorf("args[%d]: %w", i, err)
			}
		}

		err = dc.dataConverter.FromPayload(payload, valuePtrs[i])
		if err != nil {
			return fmt.Errorf("args[%d]: %w", i, err)
		}
	}

	return nil
}

// FromPayload converts single value from payload.
func (dc *CryptDataConverter) FromPayload(payload *commonpb.Payload, valuePtr interface{}) error {
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
	if !isEncryptedPayload(payload) {
		return dc.dataConverter.ToString(payload)
	}

	decryptedPayload, err := dc.decryptPayload(payload)
	if err != nil {
		return err.Error()
	}

	return dc.dataConverter.ToString(decryptedPayload)
}
