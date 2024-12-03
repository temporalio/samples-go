package blobstore_data_converter

import (
	"context"
	"github.com/temporalio/samples-go/blob-store-data-converter/blobstore"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/workflow"
)

type DataConverter struct {
	client *blobstore.Client

	parent converter.DataConverter // Until EncodingDataConverter supports workflow.ContextAware we'll store parent here.

	converter.DataConverter // embeds converter.DataConverter
}

var _ = workflow.ContextAware(&DataConverter{}) // Ensure that DataConverter implements workflow.ContextAware

// NewDataConverter returns DataConverter, which embeds converter.DataConverter
func NewDataConverter(parent converter.DataConverter, client *blobstore.Client) *DataConverter {
	next := []converter.PayloadCodec{
		NewBlobCodec(client, UnknownTenant()),
	}

	return &DataConverter{
		client:        client,
		parent:        parent,
		DataConverter: converter.NewCodecDataConverter(parent, next...),
	}
}

// WithContext will create a BlobCodec used to store and retrieve payloads from the blob storage
//
// This is called when payloads needs to be passed between the Clients/Activity and the Temporal Server. e.g.
// - From starter to encode/decode Workflow Input and Result
// - For each Activity to encode/decode it's Input and Result
func (dc *DataConverter) WithContext(ctx context.Context) converter.DataConverter {
	if vals, ok := ctx.Value(PropagatedValuesKey).(PropagatedValues); ok {
		parent := dc.parent
		if parentWithContext, ok := parent.(workflow.ContextAware); ok {
			parent = parentWithContext.WithContext(ctx)
		}

		return converter.NewCodecDataConverter(parent, NewBlobCodec(dc.client, vals))
	}

	return dc
}

// WithWorkflowContext will create a BlobCodec used to store payloads in blob storage
//
// This is called inside the Workflow to decode/encode the Workflow Input and Result
func (dc *DataConverter) WithWorkflowContext(ctx workflow.Context) converter.DataConverter {
	if vals, ok := ctx.Value(PropagatedValuesKey).(PropagatedValues); ok {
		parent := dc.parent
		if parentWithContext, ok := parent.(workflow.ContextAware); ok {
			parent = parentWithContext.WithWorkflowContext(ctx)
		}

		return converter.NewCodecDataConverter(parent, NewBlobCodec(dc.client, vals))
	}

	return dc
}
