package externalstorage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GenerateOrders_DeterministicForSameBatchID(t *testing.T) {
	first := generateOrders("BATCH-DETERMINISM", 5)
	second := generateOrders("BATCH-DETERMINISM", 5)
	require.Equal(t, first, second, "same batch ID must produce identical orders")
}

func Test_GenerateOrders_DifferentBatchIDsDiffer(t *testing.T) {
	a := generateOrders("BATCH-A", 5)
	b := generateOrders("BATCH-B", 5)
	require.NotEqual(t, a, b, "different batch IDs must produce different orders")
}
