package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
	"google.golang.org/protobuf/proto"
)

const (
	EncodingAESGCM   = "binary/nexus-aes-gcm"
	MetadataKeyID    = "encryption-key-id"
	MetadataEndpoint = "nexus-endpoint-name"
)

// aesGCMCodec follows the repository's encryption sample: encrypt the entire serialized
// Payload so its original encoding metadata is restored after decryption.
type aesGCMCodec struct {
	keyID    string
	key      []byte
	endpoint string
}

func (c *aesGCMCodec) gcm() (cipher.AEAD, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func (c *aesGCMCodec) Encode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	gcm, err := c.gcm()
	if err != nil {
		return nil, err
	}
	result := make([]*commonpb.Payload, len(payloads))
	for i, payload := range payloads {
		plaintext, err := proto.Marshal(payload)
		if err != nil {
			return nil, err
		}
		nonce := make([]byte, gcm.NonceSize())
		if _, err := rand.Read(nonce); err != nil {
			return nil, err
		}
		metadata := map[string][]byte{
			converter.MetadataEncoding: []byte(EncodingAESGCM),
			MetadataKeyID:              []byte(c.keyID),
		}
		if c.endpoint != "" {
			metadata[MetadataEndpoint] = []byte(c.endpoint)
		}
		result[i] = &commonpb.Payload{Metadata: metadata, Data: gcm.Seal(nonce, nonce, plaintext, []byte(c.endpoint))}
	}
	return result, nil
}

func (c *aesGCMCodec) Decode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	gcm, err := c.gcm()
	if err != nil {
		return nil, err
	}
	result := make([]*commonpb.Payload, len(payloads))
	for i, payload := range payloads {
		if string(payload.Metadata[converter.MetadataEncoding]) != EncodingAESGCM {
			result[i] = payload
			continue
		}
		if string(payload.Metadata[MetadataKeyID]) != c.keyID || string(payload.Metadata[MetadataEndpoint]) != c.endpoint {
			return nil, fmt.Errorf("payload key or endpoint does not match codec for %q", c.keyID)
		}
		if len(payload.Data) < gcm.NonceSize() {
			return nil, fmt.Errorf("encrypted payload is shorter than its nonce")
		}
		plaintext, err := gcm.Open(nil, payload.Data[:gcm.NonceSize()], payload.Data[gcm.NonceSize():], []byte(c.endpoint))
		if err != nil {
			return nil, fmt.Errorf("decrypt payload with key %q: %w", c.keyID, err)
		}
		result[i] = &commonpb.Payload{}
		if err := proto.Unmarshal(plaintext, result[i]); err != nil {
			return nil, err
		}
	}
	return result, nil
}
