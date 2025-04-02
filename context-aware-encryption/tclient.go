package contextawareencryption

import (
	"context"
	sdkclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/workflow"
	"sync"
)

var once sync.Once
var defaultTemporalClient sdkclient.Client

func GetDefaultOptions() sdkclient.Options {
	// If you intend to let the dataConverter to decide encryption key for all workflows
	// you can set the KeyID for the encryption encoder like so:
	//
	//   DataConverter: encryption.NewEncryptionDataConverter(
	// 	  converter.GetDefaultDataConverter(),
	// 	  encryption.DataConverterOptions{KeyID: "test", Compress: true},
	//   ),
	//
	// In this case you do not need to use a ContextPropagator.
	// You also can implement the dataConverter to decide the encryption key
	// dynamically so that it's not always the same key.
	//
	// If you need to let the workflow starter to decide the encryption key per workflow,
	// you can instead leave the KeyID unset for the encoder and supply it via the workflow
	// context as shown below. For this use case you will also need to use a
	// ContextPropagator so that KeyID is also available in the context for activities.
	//
	// Set DataConverter to ensure that workflow inputs and results are
	// encrypted/decrypted as required.
	dataConverter := NewEncryptionDataConverter(
		converter.GetDefaultDataConverter(),
		DataConverterOptions{Compress: true},
	)

	// Use a ContextPropagator so that the KeyID value set in the workflow context is
	// also availble in the context for activities.
	ctxProp := NewContextPropagator()
	options := sdkclient.Options{
		DataConverter:      dataConverter,
		ContextPropagators: []workflow.ContextPropagator{ctxProp},
	}
	return options
}
func GetTemporalClient(ctx context.Context, opts sdkclient.Options) (sdkclient.Client, error) {
	c, err := sdkclient.Dial(opts)
	return c, err
}
func MustGetDefaultTemporalClient(ctx context.Context, opts *sdkclient.Options) sdkclient.Client {
	once.Do(func() {
		if opts == nil {
			o := GetDefaultOptions()
			opts = &o
		}

		var err error
		defaultTemporalClient, err = GetTemporalClient(ctx, *opts)
		if err != nil {
			panic("failed to create default temporal client: " + err.Error())
		}
	})
	return defaultTemporalClient
}
