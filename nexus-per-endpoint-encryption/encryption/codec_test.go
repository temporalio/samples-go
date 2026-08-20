package encryption

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/converter"

	"github.com/temporalio/samples-go/nexus-per-endpoint-encryption/service"
)

func TestCodecUsesDifferentKeysForEachEndpoint(t *testing.T) {
	require.NotEqual(t, endpointKeys[service.EndpointA], endpointKeys[service.EndpointB])

	dataConverter := NewDataConverter()
	endpointAConverter := converter.WithDataConverterSerializationContext(
		dataConverter,
		converter.NexusSerializationContext{Endpoint: service.EndpointA},
	)
	endpointBConverter := converter.WithDataConverterSerializationContext(
		dataConverter,
		converter.NexusSerializationContext{Endpoint: service.EndpointB},
	)

	endpointAPayload, err := endpointAConverter.ToPayload("secret")
	require.NoError(t, err)
	endpointBPayload, err := endpointBConverter.ToPayload("secret")
	require.NoError(t, err)
	require.NotEqual(t, endpointAPayload.Data, endpointBPayload.Data)

	var value string
	require.NoError(t, endpointAConverter.FromPayload(endpointAPayload, &value))
	require.Equal(t, "secret", value)
	require.Error(t, endpointBConverter.FromPayload(endpointAPayload, &value))
}

func TestCodecLeavesNonNexusPayloadsUnchanged(t *testing.T) {
	dataConverter := converter.WithDataConverterSerializationContext(
		NewDataConverter(),
		converter.WorkflowSerializationContext{Namespace: "namespace", WorkflowID: "workflow-id"},
	)

	payload, err := dataConverter.ToPayload("plain")
	require.NoError(t, err)
	require.NotEqual(t, metadataEncodingEncrypted, string(payload.Metadata[converter.MetadataEncoding]))
}
