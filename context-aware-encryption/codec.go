package contextawareencryption

import (
	"fmt"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
)

// Codec implements PayloadCodec using AES Crypt.
type Codec struct {
	KeyID  string
	Tenant string
}

func (e *Codec) getKey(keyID string) (key []byte) {
	// Key must be fetched from secure storage in production (such as a KMS).
	// For testing here we just hard code a key.
	result := keyID + "test-key-test-key-test-key-test!"
	return []byte(result[0:32])
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
		fmt.Println("codec.Encode using tenant/key:", e.KeyID)

		b, err := encrypt(origBytes, key)
		if err != nil {
			return payloads, err
		}

		result[i] = &commonpb.Payload{
			Metadata: map[string][]byte{
				converter.MetadataEncoding: []byte(MetadataEncodingEncrypted),
				MetadataEncryptionKeyID:    []byte(e.KeyID),
				MetadataTenant:             []byte(e.Tenant),
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
		tenant, ok := p.Metadata[MetadataTenant]
		if !ok {
			return payloads, fmt.Errorf("no tenant id")
		}
		key := e.getKey(string(keyID))
		fmt.Println("codec.Decode using tenant/key:", string(tenant), string(key))

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

func GetTenantCodec(key string) *Codec {
	availableCodecs := map[string]*Codec{}
	for tenant, keyId := range TenantKeysByOrganization {
		availableCodecs[keyId] = &Codec{Tenant: tenant, KeyID: keyId}
	}
	if c, exists := availableCodecs[key]; exists {
		return c
	}
	return &Codec{Tenant: "UNKNOWN TENANT", KeyID: key}
}
