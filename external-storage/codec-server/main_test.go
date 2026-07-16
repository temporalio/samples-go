package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	"github.com/stretchr/testify/require"
	externalstorage "github.com/temporalio/samples-go/external-storage"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/contrib/aws/s3driver"
	"go.temporal.io/sdk/contrib/aws/s3driver/awssdkv2"
	"go.temporal.io/sdk/converter"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	testBucket    = "test-bucket"
	testNamespace = "default"
)

// newCodecServer builds the full codec server middleware stack against an
// in-memory gofakes3, mirroring what main() does. Returned httptest.Server is
// closed by the caller.
func newCodecServer(t *testing.T) *httptest.Server {
	t.Helper()

	backend := s3mem.New()
	require.NoError(t, backend.CreateBucket(testBucket))
	s3Server := httptest.NewServer(gofakes3.New(backend).Server())
	t.Cleanup(s3Server.Close)

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
	)
	require.NoError(t, err)
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(s3Server.URL)
		o.UsePathStyle = true
	})

	driver, err := s3driver.NewDriver(s3driver.Options{
		Client: awssdkv2.NewClient(s3Client),
		Bucket: s3driver.StaticBucket(testBucket),
	})
	require.NoError(t, err)

	// Exercise the same handler stack main() builds.
	handler, err := newCodecServerHandler(driver)
	require.NoError(t, err)

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func smallPayload() *commonpb.Payload {
	return &commonpb.Payload{
		Metadata: map[string][]byte{"encoding": []byte("json/plain")},
		Data:     []byte(`{"hello":"world"}`),
	}
}

// callPayloads POSTs the given payloads to the codec server and returns the
// decoded response.
func callPayloads(t *testing.T, url, namespace string, payloads ...*commonpb.Payload) (int, []*commonpb.Payload) {
	t.Helper()
	body, err := protojson.Marshal(&commonpb.Payloads{Payloads: payloads})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if namespace != "" {
		req.Header.Set("X-Namespace", namespace)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}

	var out commonpb.Payloads
	require.NoError(t, protojson.Unmarshal(respBody, &out))
	return resp.StatusCode, out.Payloads
}

func Test_UnknownNamespaceReturns404(t *testing.T) {
	srv := newCodecServer(t)

	status, _ := callPayloads(t, srv.URL+"/encode", "unregistered-ns", smallPayload())
	require.Equal(t, http.StatusNotFound, status)
}

func Test_CORSPreflight(t *testing.T) {
	srv := newCodecServer(t)

	req, err := http.NewRequest(http.MethodOptions, srv.URL+"/decode", nil)
	require.NoError(t, err)
	req.Header.Set("Origin", webUIOrigin)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, webUIOrigin, resp.Header.Get("Access-Control-Allow-Origin"))
	require.Equal(t, "POST,OPTIONS", resp.Header.Get("Access-Control-Allow-Methods"))
	require.Contains(t, resp.Header.Get("Access-Control-Allow-Headers"), "X-Namespace")
}

func Test_CORSRejectsOtherOrigin(t *testing.T) {
	srv := newCodecServer(t)

	req, err := http.NewRequest(http.MethodOptions, srv.URL+"/decode", nil)
	require.NoError(t, err)
	req.Header.Set("Origin", "http://evil.example.com")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	// The middleware only sets CORS headers when Origin matches webUIOrigin;
	// a request from any other origin gets no CORS headers, which makes the
	// browser refuse to use the response.
	require.Empty(t, resp.Header.Get("Access-Control-Allow-Origin"))
	require.Empty(t, resp.Header.Get("Access-Control-Allow-Methods"))
	require.Empty(t, resp.Header.Get("Access-Control-Allow-Headers"))
}

func Test_CORSHeadersOnPostResponse(t *testing.T) {
	srv := newCodecServer(t)

	// CORS headers must land on the actual POST response too, not just the
	// OPTIONS preflight — the browser checks both.
	body, err := protojson.Marshal(&commonpb.Payloads{Payloads: []*commonpb.Payload{smallPayload()}})
	require.NoError(t, err)
	req, err := http.NewRequest(http.MethodPost, srv.URL+"/encode", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Namespace", testNamespace)
	req.Header.Set("Origin", webUIOrigin)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, webUIOrigin, resp.Header.Get("Access-Control-Allow-Origin"))
}

// Test_DecodePayloadCompatibility checks compatibility between the codec
// server and the DataConverter that the worker/starter use.
func Test_DecodePayloadCompatibility(t *testing.T) {
	srv := newCodecServer(t)

	// Use the exact DataConverter the worker/starter use, so a change to
	// either side breaks this test rather than passing silently.
	workerConv := externalstorage.NewSampleDataConverter()

	type sample struct {
		Greeting string `json:"greeting"`
		Count    int    `json:"count"`
	}
	original := sample{Greeting: "hello", Count: 42}

	// Encode as the worker would, then send the result to /decode.
	encoded, err := workerConv.ToPayloads(original)
	require.NoError(t, err)
	require.Equal(t, "binary/zlib", string(encoded.Payloads[0].GetMetadata()["encoding"]))

	status, decoded := callPayloads(t, srv.URL+"/decode", testNamespace, encoded.Payloads...)
	require.Equal(t, http.StatusOK, status)
	require.Len(t, decoded, 1)

	var got sample
	require.NoError(t, converter.GetDefaultDataConverter().FromPayloads(
		&commonpb.Payloads{Payloads: decoded}, &got))
	require.Equal(t, original, got)
}
