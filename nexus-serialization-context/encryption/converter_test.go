package encryption_test

import (
	"crypto/aes"
	"crypto/cipher"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"

	"github.com/temporalio/samples-go/nexus-serialization-context/encryption"
	"github.com/temporalio/samples-go/nexus-serialization-context/service"
)

// These tests supply context explicitly to check the sample configuration, not SDK propagation.
func TestEndpointRoundTrips(t *testing.T) {
	caller := encryption.NewCallerDataConverter()
	compressed := encryption.NewCompressedEncryptedDataConverter()
	encrypted := encryption.NewEncryptedDataConverter()
	for _, tc := range []struct {
		endpoint, keyID, key string
		handler              converter.DataConverter
	}{
		{service.CompressedEncrypted1Endpoint, "compressed-encrypted-1", "handler-a-key-01", compressed},
		{service.CompressedEncrypted2Endpoint, "compressed-encrypted-2", "handler-a-key-02", compressed},
		{service.Encrypted1Endpoint, "encrypted-1", "handler-b-key-01", encrypted},
		{service.Encrypted2Endpoint, "encrypted-2", "handler-b-key-02", encrypted},
	} {
		t.Run(tc.endpoint, func(t *testing.T) {
			ctx := converter.NexusSerializationContext{Endpoint: tc.endpoint, Service: service.Name, Operation: service.SyncOperationName}
			caller := converter.WithDataConverterSerializationContext(caller, ctx)
			handler := converter.WithDataConverterSerializationContext(tc.handler, ctx)
			for _, direction := range []struct {
				name           string
				encode, decode converter.DataConverter
			}{
				{"request", caller, handler}, {"response", handler, caller},
			} {
				t.Run(direction.name, func(t *testing.T) {
					payload, err := direction.encode.ToPayload("Hello from Nexus")
					require.NoError(t, err)
					require.Equal(t, encryption.EncodingAESGCM, string(payload.Metadata[converter.MetadataEncoding]))
					require.Equal(t, tc.keyID, string(payload.Metadata[encryption.MetadataKeyID]))
					require.Equal(t, tc.endpoint, string(payload.Metadata[encryption.MetadataEndpoint]))
					requireEncryptedWithKey(t, payload, tc.key, tc.endpoint)
					var message string
					require.NoError(t, direction.decode.FromPayload(payload, &message))
					require.Equal(t, "Hello from Nexus", message)
				})
			}
		})
	}
}

func TestCompressionBeforeEncryption(t *testing.T) {
	dc := encryption.NewCallerDataConverter()
	for _, endpoints := range [][2]string{
		{service.CompressedEncrypted1Endpoint, service.Encrypted1Endpoint},
		{service.CompressedEncrypted2Endpoint, service.Encrypted2Endpoint},
	} {
		t.Run(endpoints[0], func(t *testing.T) {
			compressed := converter.WithDataConverterSerializationContext(dc, converter.NexusSerializationContext{Endpoint: endpoints[0]})
			encrypted := converter.WithDataConverterSerializationContext(dc, converter.NexusSerializationContext{Endpoint: endpoints[1]})
			message := strings.Repeat("compress this repeated message ", 100)
			small, err := compressed.ToPayload(message)
			require.NoError(t, err)
			large, err := encrypted.ToPayload(message)
			require.NoError(t, err)
			require.Less(t, len(small.Data), len(large.Data))
			var decoded string
			require.NoError(t, compressed.FromPayload(small, &decoded))
			require.Equal(t, message, decoded)
		})
	}
}

func TestOrdinaryWorkflowDefaults(t *testing.T) {
	for _, tc := range []struct {
		name, keyID, key string
		dc               converter.DataConverter
	}{
		{"caller", "caller-workflow", "sample-key-C-123", encryption.NewCallerDataConverter()},
		{"compressed-handler", "compressed-encrypted-1", "handler-a-key-01", encryption.NewCompressedEncryptedDataConverter()},
		{"encrypted-handler", "encrypted-1", "handler-b-key-01", encryption.NewEncryptedDataConverter()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dc := converter.WithDataConverterSerializationContext(tc.dc, converter.WorkflowSerializationContext{Namespace: tc.name, WorkflowID: "echo"})
			payload, err := dc.ToPayload("workflow input")
			require.NoError(t, err)
			require.Equal(t, encryption.EncodingAESGCM, string(payload.Metadata[converter.MetadataEncoding]))
			require.Equal(t, tc.keyID, string(payload.Metadata[encryption.MetadataKeyID]))
			require.NotContains(t, payload.Metadata, encryption.MetadataEndpoint)
			requireEncryptedWithKey(t, payload, tc.key, "")
			var decoded string
			require.NoError(t, dc.FromPayload(payload, &decoded))
			require.Equal(t, "workflow input", decoded)
		})
	}
	// The caller's final result has ordinary workflow context too, even though it contains Nexus replies.
	original := service.WorkflowOutput{
		CompressedEncrypted1: service.EndpointResults{Sync: "c1 sync", Async: "c1 async"},
		CompressedEncrypted2: service.EndpointResults{Sync: "c2 sync", Async: "c2 async"},
		Encrypted1:           service.EndpointResults{Sync: "e1 sync", Async: "e1 async"},
		Encrypted2:           service.EndpointResults{Sync: "e2 sync", Async: "e2 async"},
	}
	dc := encryption.NewCallerDataConverter()
	payload, err := dc.ToPayload(original)
	require.NoError(t, err)
	require.Equal(t, "caller-workflow", string(payload.Metadata[encryption.MetadataKeyID]))
	requireEncryptedWithKey(t, payload, "sample-key-C-123", "")
	var decoded service.WorkflowOutput
	require.NoError(t, dc.FromPayload(payload, &decoded))
	require.Equal(t, original, decoded)
}

func TestBoundConvertersKeepTheirContext(t *testing.T) {
	base := encryption.NewEncryptedDataConverter()
	first := converter.WithDataConverterSerializationContext(base, converter.NexusSerializationContext{Endpoint: service.Encrypted1Endpoint})
	second := converter.WithDataConverterSerializationContext(base, converter.NexusSerializationContext{Endpoint: service.Encrypted2Endpoint})
	workflow := converter.WithDataConverterSerializationContext(second, converter.WorkflowSerializationContext{Namespace: "handler", WorkflowID: "echo"})
	// Use each converter after both bindings exist. Binding must also leave the base unchanged.
	for _, tc := range []struct {
		name, endpoint, key string
		dc                  converter.DataConverter
	}{
		{"first", service.Encrypted1Endpoint, "handler-b-key-01", first},
		{"second", service.Encrypted2Endpoint, "handler-b-key-02", second},
		{"base", "", "handler-b-key-01", base},
		{"rebound-workflow", "", "handler-b-key-01", workflow},
	} {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := tc.dc.ToPayload("Hello from Nexus")
			require.NoError(t, err)
			require.Equal(t, tc.endpoint, string(payload.Metadata[encryption.MetadataEndpoint]))
			requireEncryptedWithKey(t, payload, tc.key, tc.endpoint)
			var message string
			require.NoError(t, tc.dc.FromPayload(payload, &message))
			require.Equal(t, "Hello from Nexus", message)
		})
	}
}

func TestEncryptionUsesFreshNonce(t *testing.T) {
	dc := encryption.NewCallerDataConverter()
	first, err := dc.ToPayload("same input")
	require.NoError(t, err)
	second, err := dc.ToPayload("same input")
	require.NoError(t, err)
	require.NotEqual(t, first.Data[:12], second.Data[:12])
	require.NotEqual(t, first.Data, second.Data)
}

func TestRejectsDamagedCiphertext(t *testing.T) {
	dc := encryption.NewCallerDataConverter()
	payload, err := dc.ToPayload("Hello from Nexus")
	require.NoError(t, err)
	altered := append([]byte(nil), payload.Data...)
	altered[len(altered)/2] ^= 1
	for _, tc := range []struct {
		name string
		data []byte
	}{
		{"altered", altered},
		{"incomplete nonce prefix", payload.Data[:1]},
		{"truncated tag", payload.Data[:len(payload.Data)-1]},
	} {
		t.Run(tc.name, func(t *testing.T) {
			damaged := &commonpb.Payload{Metadata: payload.Metadata, Data: tc.data}
			var message string
			require.Error(t, dc.FromPayload(damaged, &message))
		})
	}
}

// Authenticate with independently specified key bytes, not the sample's decoder
// or the payload's key-ID label, so matching encode/decode mistakes cannot hide wrong keys.
func requireEncryptedWithKey(t *testing.T, payload *commonpb.Payload, key, endpoint string) {
	t.Helper()
	block, err := aes.NewCipher([]byte(key))
	require.NoError(t, err)
	gcm, err := cipher.NewGCM(block)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(payload.Data), gcm.NonceSize()+gcm.Overhead())
	_, err = gcm.Open(nil, payload.Data[:gcm.NonceSize()], payload.Data[gcm.NonceSize():], []byte(endpoint))
	require.NoError(t, err)
}

func TestUnknownEndpoint(t *testing.T) {
	for _, dc := range []converter.DataConverter{
		encryption.NewCallerDataConverter(), encryption.NewCompressedEncryptedDataConverter(), encryption.NewEncryptedDataConverter(),
	} {
		payload, err := dc.ToPayload("ordinary workflow input")
		require.NoError(t, err)
		unknown := converter.WithDataConverterSerializationContext(dc, converter.NexusSerializationContext{Endpoint: "unknown"})
		_, err = unknown.ToPayload("request")
		require.ErrorContains(t, err, "no codec configured")
		var decoded string
		require.ErrorContains(t, unknown.FromPayload(payload, &decoded), "no codec configured")
	}
}
