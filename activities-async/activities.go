package activities_async

import (
	"context"
)

func SayHello(ctx context.Context, param string) (string, error) {
	return "Hello " + param + "!", nil
}

func SayGoodbye(ctx context.Context, param string) (string, error) {
	return "Goodbye " + param + "!", nil
}
