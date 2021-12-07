package encryption

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/converter"
)

func Test_DataConverter(t *testing.T) {
	defaultDc := converter.GetDefaultDataConverter()

	ctx := context.Background()
	ctx = context.WithValue(ctx, PropagateKey, CryptContext{KeyID: "test"})

	cryptDc := NewEncryptionDataConverter(
		converter.GetDefaultDataConverter(),
		DataConverterOptions{},
	)
	cryptDcWc := cryptDc.WithContext(ctx)

	defaultPayloads, err := defaultDc.ToPayloads("Testing")
	require.NoError(t, err)

	encryptedPayloads, err := cryptDcWc.ToPayloads("Testing")
	require.NoError(t, err)

	require.NotEqual(t, defaultPayloads.Payloads[0].GetData(), encryptedPayloads.Payloads[0].GetData())

	var result string
	err = cryptDc.FromPayloads(encryptedPayloads, &result)
	require.NoError(t, err)

	require.Equal(t, "Testing", result)
}
