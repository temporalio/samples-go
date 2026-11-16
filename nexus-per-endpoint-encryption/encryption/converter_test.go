package encryption_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/converter"

	"github.com/temporalio/samples-go/nexus-per-endpoint-encryption/encryption"
	"github.com/temporalio/samples-go/nexus-per-endpoint-encryption/service"
)

// These tests supply context explicitly to check the sample configuration, not SDK propagation.
func TestEndpointRoundTrips(t *testing.T) {
	caller := encryption.NewCallerDataConverter()
	compressed := encryption.NewCompressedEncryptedDataConverter()
	encrypted := encryption.NewEncryptedDataConverter()
	for _, tc := range []struct {
		endpoint, keyID string
		handler         converter.DataConverter
	}{
		{service.CompressedEncrypted1Endpoint, "compressed-encrypted-1", compressed},
		{service.CompressedEncrypted2Endpoint, "compressed-encrypted-2", compressed},
		{service.Encrypted1Endpoint, "encrypted-1", encrypted},
		{service.Encrypted2Endpoint, "encrypted-2", encrypted},
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
		name, keyID string
		dc          converter.DataConverter
	}{
		{"caller", "caller-workflow", encryption.NewCallerDataConverter()},
		{"compressed-handler", "compressed-encrypted-1", encryption.NewCompressedEncryptedDataConverter()},
		{"encrypted-handler", "encrypted-1", encryption.NewEncryptedDataConverter()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dc := converter.WithDataConverterSerializationContext(tc.dc, converter.WorkflowSerializationContext{Namespace: tc.name, WorkflowID: "echo"})
			payload, err := dc.ToPayload("workflow input")
			require.NoError(t, err)
			require.Equal(t, encryption.EncodingAESGCM, string(payload.Metadata[converter.MetadataEncoding]))
			require.Equal(t, tc.keyID, string(payload.Metadata[encryption.MetadataKeyID]))
			require.NotContains(t, payload.Metadata, encryption.MetadataEndpoint)
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
	var decoded service.WorkflowOutput
	require.NoError(t, dc.FromPayload(payload, &decoded))
	require.Equal(t, original, decoded)
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
