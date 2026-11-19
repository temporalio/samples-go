package encryption

import (
	"fmt"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"

	"github.com/temporalio/samples-go/nexus-serialization-context/service"
)

// These are demonstration keys, not production secrets. Load real keys from a secure key store.
var (
	compressedEncryptedKey1 = []byte("handler-a-key-01")
	compressedEncryptedKey2 = []byte("handler-a-key-02")
	encryptedKey1           = []byte("handler-b-key-01")
	encryptedKey2           = []byte("handler-b-key-02")
	callerWorkflowKey       = []byte("sample-key-C-123")
)

// NewCallerDataConverter configures all endpoint keys and a separate workflow default.
func NewCallerDataConverter() converter.DataConverter {
	return newDataConverter(map[string]converter.PayloadCodec{
		service.CompressedEncrypted1Endpoint: endpointCodec(service.CompressedEncrypted1Endpoint, "compressed-encrypted-1", compressedEncryptedKey1, true),
		service.CompressedEncrypted2Endpoint: endpointCodec(service.CompressedEncrypted2Endpoint, "compressed-encrypted-2", compressedEncryptedKey2, true),
		service.Encrypted1Endpoint:           endpointCodec(service.Encrypted1Endpoint, "encrypted-1", encryptedKey1, false),
		service.Encrypted2Endpoint:           endpointCodec(service.Encrypted2Endpoint, "encrypted-2", encryptedKey2, false),
	}, &aesGCMCodec{keyID: "caller-workflow", key: callerWorkflowKey})
}

func NewCompressedEncryptedDataConverter() converter.DataConverter {
	return newDataConverter(map[string]converter.PayloadCodec{
		service.CompressedEncrypted1Endpoint: endpointCodec(service.CompressedEncrypted1Endpoint, "compressed-encrypted-1", compressedEncryptedKey1, true),
		service.CompressedEncrypted2Endpoint: endpointCodec(service.CompressedEncrypted2Endpoint, "compressed-encrypted-2", compressedEncryptedKey2, true),
	}, &aesGCMCodec{keyID: "compressed-encrypted-1", key: compressedEncryptedKey1})
}

func NewEncryptedDataConverter() converter.DataConverter {
	return newDataConverter(map[string]converter.PayloadCodec{
		service.Encrypted1Endpoint: endpointCodec(service.Encrypted1Endpoint, "encrypted-1", encryptedKey1, false),
		service.Encrypted2Endpoint: endpointCodec(service.Encrypted2Endpoint, "encrypted-2", encryptedKey2, false),
	}, &aesGCMCodec{keyID: "encrypted-1", key: encryptedKey1})
}

func endpointCodec(endpoint, keyID string, key []byte, compress bool) converter.PayloadCodec {
	encryption := &aesGCMCodec{endpoint: endpoint, keyID: keyID, key: key}
	if !compress {
		return encryption
	}
	// SDK codecs encode last-to-first: compress first, then encrypt.
	dc := converter.NewCodecDataConverter(converter.GetDefaultDataConverter(), encryption,
		converter.NewZlibCodec(converter.ZlibCodecOptions{AlwaysEncode: true}))
	return dc.(converter.PayloadCodecProvider).Codec()
}

func newDataConverter(endpoints map[string]converter.PayloadCodec, workflowCodec converter.PayloadCodec) converter.DataConverter {
	return converter.NewCodecDataConverter(converter.GetDefaultDataConverter(), &nexusCodec{
		endpoints: endpoints, workflowCodec: workflowCodec,
	})
}

// Bound copies share read-only configuration. Each operation keeps its own endpoint.
type nexusCodec struct {
	endpoints     map[string]converter.PayloadCodec
	workflowCodec converter.PayloadCodec
	endpoint      string
}

var _ converter.PayloadCodecWithSerializationContext = (*nexusCodec)(nil)

func (c *nexusCodec) WithSerializationContext(ctx converter.SerializationContext) converter.PayloadCodec {
	// Copy the struct so each call keeps its own endpoint.
	// The map and configured codecs stay shared and read-only.
	bound := *c
	bound.endpoint = ""
	if nexusCtx, ok := ctx.(converter.NexusSerializationContext); ok {
		bound.endpoint = nexusCtx.Endpoint
	}
	return &bound
}

func (c *nexusCodec) selected() (converter.PayloadCodec, error) {
	if c.endpoint == "" {
		return c.workflowCodec, nil
	}
	codec, ok := c.endpoints[c.endpoint]
	if !ok {
		return nil, fmt.Errorf("no codec configured for Nexus endpoint %q", c.endpoint)
	}
	return codec, nil
}

func (c *nexusCodec) Encode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	codec, err := c.selected()
	if err != nil {
		return nil, err
	}
	return codec.Encode(payloads)
}

func (c *nexusCodec) Decode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	codec, err := c.selected()
	if err != nil {
		return nil, err
	}
	return codec.Decode(payloads)
}
