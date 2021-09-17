package cryptconverter

import (
	"fmt"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/interceptors"
)

const (
	// MetadataEncryptionKeyId is "encryption-key-id"
	MetadataEncryptionKeyId = "encryption-key-id"

	// MetadataEncodingEncrypted is "binary/encrypted"
	MetadataEncodingEncrypted = "binary/encrypted"
)

func NewCryptInputResultsInterceptor() interceptors.ServiceInterceptor {
	return interceptors.NewInputsResultsServiceInterceptor(
		interceptors.InterceptorEncoder{
			Encode: encryptPayloads,
			Decode: decryptPayloads,
		},
	)
}

func NewCryptHeartbeatDetailsInterceptor() interceptors.ServiceInterceptor {
	return interceptors.NewHeartbeatDetailsServiceInterceptor(
		interceptors.InterceptorEncoder{
			Encode: encryptPayloads,
			Decode: decryptPayloads,
		},
	)
}

func getKeyID() []byte {
	return []byte("test")
}

func getKey(_ []byte) (key []byte) {
	// Key must be fetched from secure storage in production (such as a KMS).
	// For testing here we just hard code a key.
	return []byte("test-key-test-key-test-key-test!")
}

func encryptPayloads(payloads *commonpb.Payloads) (*commonpb.Payloads, error) {
	keyId := getKeyID()
	key := getKey(keyId)

	result := commonpb.Payloads{}

	for _, payload := range payloads.Payloads {
		encryptedPayload, err := encryptPayload(payload, keyId, key)
		if err != nil {
			return &result, err
		}
		result.Payloads = append(result.Payloads, encryptedPayload)
	}

	return &result, nil
}

func decryptPayloads(payloads *commonpb.Payloads) (*commonpb.Payloads, error) {
	result := commonpb.Payloads{}

	for _, payload := range payloads.Payloads {
		decryptedPayload, err := decryptPayload(payload)
		if err != nil {
			return &result, err
		}
		result.Payloads = append(result.Payloads, decryptedPayload)
	}

	return &result, nil
}

func encryptPayload(payload *commonpb.Payload, keyId []byte, key []byte) (*commonpb.Payload, error) {
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
			MetadataEncryptionKeyId:    []byte(keyId),
		},
		Data: encryptedPayload,
	}, nil
}

func decryptPayload(payload *commonpb.Payload) (*commonpb.Payload, error) {
	metadata := payload.GetMetadata()
	if metadata == nil {
		return nil, converter.ErrMetadataIsNotSet
	}

	encoding, ok := metadata[converter.MetadataEncoding]
	if !ok || string(encoding) != MetadataEncodingEncrypted {
		return payload, nil
	}

	keyId, ok := metadata[MetadataEncryptionKeyId]
	if !ok {
		return nil, fmt.Errorf("%w: %s metadata not set", converter.ErrUnableToDecode, MetadataEncryptionKeyId)
	}

	key := getKey(keyId)
	serializedPayload, err := decrypt(payload.GetData(), key)
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
