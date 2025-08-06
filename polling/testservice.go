package polling

import (
	"context"

	"go.temporal.io/sdk/temporal"
)

type TestService struct {
	tryAttempts   int
	errorAttempts int
}

func NewTestService(errorAttempts int) TestService {
	return TestService{
		tryAttempts:   0,
		errorAttempts: errorAttempts,
	}
}

func (testService *TestService) GetServiceResult(ctx context.Context) (string, error) {
	testService.tryAttempts += 1
	if testService.tryAttempts%testService.errorAttempts == 0 {
		return "OK", nil
	}
	return "", temporal.NewApplicationErrorWithOptions(
		"service is down", "ServiceError", temporal.ApplicationErrorOptions{
			// This error is expected so we set it as benign to avoid excessive logging
			Category: temporal.ApplicationErrorCategoryBenign,
		})
}
