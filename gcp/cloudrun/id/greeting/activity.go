package greeting

import (
	"context"
	"fmt"
)

// HelloActivity returns a greeting string.
func HelloActivity(ctx context.Context, name string) (string, error) {
	return fmt.Sprintf("Hello, %s!", name), nil
}
