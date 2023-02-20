package polling

import (
	"context"
	"errors"
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
	return "", errors.New("service is down")
}
