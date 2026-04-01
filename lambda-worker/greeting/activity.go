package greeting

import (
	"context"
	"fmt"
)

// SampleActivity is a basic activity function
func HelloActivity(ctx context.Context, name string) (string, error) {
	return fmt.Sprintf("Hello, %s!", name), nil
}
