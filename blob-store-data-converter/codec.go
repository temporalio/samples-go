package blobstore_data_converter

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/temporalio/samples-go/blob-store-data-converter/blobstore"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
	"strings"
)

const (
	MetadataEncodingBlobStorePlain = "blobstore/plain"

	// gRPC has a 4MB limit.
	// To save some space for other metadata, we should stay around half that.
	//
	// For this example, as a proof of concept, we'll use much smaller size limit.
	payloadSizeLimit = 37
)

// BlobCodec knows where to store the blobs from the PropagatedValues
// Note, see readme for details on missing values
type BlobCodec struct {
	client     *blobstore.Client
	bucket     string
	tenant     string
	pathPrefix []string
}

var _ = converter.PayloadCodec(&BlobCodec{}) // Ensure that BlobCodec implements converter.PayloadCodec

// NewBlobCodec is aware of where of the propagated context values from the data converter
func NewBlobCodec(c *blobstore.Client, values PropagatedValues) *BlobCodec {
	return &BlobCodec{
		client:     c,
		bucket:     "blob://mybucket",
		tenant:     values.TenantID,
		pathPrefix: values.BlobNamePrefix,
	}
}

// Encode knows where to store the blobs from values stored in the context
func (c *BlobCodec) Encode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	result := make([]*commonpb.Payload, len(payloads))
	for i, p := range payloads {
		// if the payload is small enough, just send it as is
		fmt.Printf("encoding payload with len(%s): %d\n", string(p.Data), len(p.Data))
		if len(p.Data) < payloadSizeLimit {
			result[i] = &commonpb.Payload{Metadata: p.Metadata, Data: p.Data}
			continue
		}

		origBytes, err := p.Marshal()
		if err != nil {
			return payloads, err
		}

		// save the data in our blob store db
		objectName := strings.Join(c.pathPrefix, "_") + "__" + uuid.New().String() // ensures each blob is unique
		path := fmt.Sprintf("%s/%s/%s", c.bucket, c.tenant, objectName)
		err = c.client.SaveBlob(path, origBytes)
		if err != nil {
			return payloads, err
		}

		result[i] = &commonpb.Payload{
			Metadata: map[string][]byte{
				"encoding": []byte(MetadataEncodingBlobStorePlain),
			},
			Data: []byte(path),
		}
	}

	return result, nil
}

// Decode does not need to be context aware because it can fetch the blobs via the payload path
func (c *BlobCodec) Decode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	result := make([]*commonpb.Payload, len(payloads))
	for i, p := range payloads {
		if string(p.Metadata["encoding"]) != MetadataEncodingBlobStorePlain {
			result[i] = &commonpb.Payload{Metadata: p.Metadata, Data: p.Data}
			continue
		}

		// fetch it from our blob store db
		data, err := c.client.GetBlob(string(p.Data))
		if err != nil {
			return payloads, err
		}

		result[i] = &commonpb.Payload{}
		err = result[i].Unmarshal(data)
		if err != nil {
			return payloads, err
		}
	}

	return result, nil
}
