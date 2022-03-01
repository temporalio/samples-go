package main

import (
	"log"
	"net"

	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"google.golang.org/grpc"

	"github.com/temporalio/samples-go/encryption"
)

const ServerEndpoint = "localhost:7234"

func main() {
	interceptor, err := converter.NewPayloadEncoderGRPCClientInterceptor(
		converter.PayloadEncoderGRPCClientInterceptorOptions{
			Encoders: encryption.NewEncoders(encryption.DataConverterOptions{Compress: true}),
		},
	)
	if err != nil {
		log.Fatalf("unable to create interceptor: %v", err)
	}

	grpcClient, err := grpc.Dial(
		"localhost:7233",
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(interceptor),
	)
	defer func() { _ = grpcClient.Close() }()

	workflowClient := workflowservice.NewWorkflowServiceClient(grpcClient)

	if err != nil {
		log.Fatalf("unable to create client: %v", err)
	}

	listener, err := net.Listen("tcp", ServerEndpoint)
	if err != nil {
		log.Fatalf("unable to create listener: %v", err)
	}

	server := grpc.NewServer()
	handler, err := client.NewWorkflowServiceProxyServer(
		client.WorkflowServiceProxyOptions{Client: workflowClient},
	)
	if err != nil {
		log.Fatalf("unable to create service proxy: %v", err)
	}

	workflowservice.RegisterWorkflowServiceServer(server, handler)

	err = server.Serve(listener)
	if err != nil {
		log.Fatalf("unable to serve: %v", err)
	}
}
