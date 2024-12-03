package blobstore_data_converter

import (
	"context"
	"github.com/temporalio/samples-go/blob-store-data-converter/blobstore"
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/converter"
)

func Test_DataConverter(t *testing.T) {
	defaultDc := converter.GetDefaultDataConverter()

	ctx := context.Background()
	ctx = context.WithValue(ctx, PropagatedValuesKey, PropagatedValues{
		TenantID:       "t1",
		BlobNamePrefix: []string{"t1", "starter"},
	})

	blobDc := NewDataConverter(
		converter.GetDefaultDataConverter(),
		blobstore.NewTestClient(),
	)
	blobDcCtx := blobDc.WithContext(ctx)

	defaultPayloads, err := defaultDc.ToPayloads("small payload")
	require.NoError(t, err)
	require.Equal(t, string(defaultPayloads.Payloads[0].GetData()), `"small payload"`)

	const largePayload = "really really really large giant payload"
	require.Greater(t, len([]byte(largePayload)), payloadSizeLimit, "payload size should be larger than the limit in the example")

	offloadedPayloads, err := blobDcCtx.ToPayloads(largePayload)
	require.NoError(t, err)
	require.Contains(t, string(offloadedPayloads.Payloads[0].GetData()), "blob://")

	var result string
	err = blobDc.FromPayloads(offloadedPayloads, &result)
	require.NoError(t, err)

	require.Equal(t, largePayload, result)
}
