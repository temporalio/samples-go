package encryption

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"

	"github.com/temporalio/samples-go/nexus-per-endpoint-encryption/service"
)

const metadataEncodingEncrypted = "binary/encrypted"

var endpointKeys = map[string][]byte{
	service.EndpointA: []byte("endpoint-a-key-32-bytes-long!!!!"),
	service.EndpointB: []byte("endpoint-b-key-32-bytes-long!!!!"),
}

type Codec struct {
	endpoint string
}

func (c *Codec) WithSerializationContext(ctx converter.SerializationContext) converter.PayloadCodec {
	nexusCtx, ok := ctx.(converter.NexusSerializationContext)
	if !ok {
		return &Codec{}
	}
	return &Codec{endpoint: nexusCtx.Endpoint}
}

func (c *Codec) Encode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	if c.endpoint == "" {
		return payloads, nil
	}
	gcm, err := c.gcm()
	if err != nil {
		return nil, err
	}

	encoded := make([]*commonpb.Payload, len(payloads))
	for i, payload := range payloads {
		plainText, err := payload.Marshal()
		if err != nil {
			return nil, fmt.Errorf("marshal payload: %w", err)
		}
		nonce := make([]byte, gcm.NonceSize())
		if _, err := rand.Read(nonce); err != nil {
			return nil, fmt.Errorf("generate nonce: %w", err)
		}
		cipherText := gcm.Seal(nonce, nonce, plainText, []byte(c.endpoint))
		encoded[i] = &commonpb.Payload{
			Metadata: map[string][]byte{
				converter.MetadataEncoding: []byte(metadataEncodingEncrypted),
				"endpoint":                 []byte(c.endpoint),
			},
			Data: cipherText,
		}
	}
	return encoded, nil
}

func (c *Codec) Decode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	decoded := make([]*commonpb.Payload, len(payloads))
	for i, payload := range payloads {
		if !bytes.Equal(payload.Metadata[converter.MetadataEncoding], []byte(metadataEncodingEncrypted)) {
			decoded[i] = payload
			continue
		}
		if c.endpoint == "" {
			return nil, fmt.Errorf("encrypted Nexus payload is missing serialization context")
		}
		if payloadEndpoint := string(payload.Metadata["endpoint"]); payloadEndpoint != c.endpoint {
			return nil, fmt.Errorf("payload endpoint %q does not match serialization context endpoint %q", payloadEndpoint, c.endpoint)
		}
		gcm, err := c.gcm()
		if err != nil {
			return nil, err
		}
		if len(payload.Data) < gcm.NonceSize() {
			return nil, fmt.Errorf("encrypted payload is shorter than its nonce")
		}
		nonce, cipherText := payload.Data[:gcm.NonceSize()], payload.Data[gcm.NonceSize():]
		plainText, err := gcm.Open(nil, nonce, cipherText, []byte(c.endpoint))
		if err != nil {
			return nil, fmt.Errorf("decrypt payload for endpoint %q: %w", c.endpoint, err)
		}
		decoded[i] = &commonpb.Payload{}
		if err := decoded[i].Unmarshal(plainText); err != nil {
			return nil, fmt.Errorf("unmarshal payload: %w", err)
		}
	}
	return decoded, nil
}

func (c *Codec) gcm() (cipher.AEAD, error) {
	key, ok := endpointKeys[c.endpoint]
	if !ok {
		return nil, fmt.Errorf("no encryption key configured for Nexus endpoint %q", c.endpoint)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create AES-GCM cipher: %w", err)
	}
	return gcm, nil
}

func NewDataConverter() converter.DataConverter {
	return converter.NewCodecDataConverter(converter.GetDefaultDataConverter(), &Codec{})
}
