package main

import (
	"flag"
	"net"
	"strconv"

	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/server/common/log/tag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.temporal.io/server/common/log"

	grpcproxy "github.com/temporalio/samples-go/grpc-proxy"
)

var portFlag int
var upstreamFlag string

func main() {
	logger := log.NewCLILogger()

	flag.IntVar(&portFlag, "port", 8081, "Port to listen on")
	flag.StringVar(&upstreamFlag, "upstream", "127.0.0.1:7233", "Upstream Temporal Server Endpoint")
	flag.Parse()

	clientInterceptor, err := converter.NewPayloadCodecGRPCClientInterceptor(
		converter.PayloadCodecGRPCClientInterceptorOptions{
			Codecs: []converter.PayloadCodec{grpcproxy.NewPayloadCodec()},
		},
	)
	if err != nil {
		logger.Fatal("unable to create interceptor: %v", tag.NewErrorTag(err))
	}

	grpcClient, err := grpc.Dial(
		upstreamFlag,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(clientInterceptor),
	)
	defer func() { _ = grpcClient.Close() }()

	workflowClient := workflowservice.NewWorkflowServiceClient(grpcClient)

	if err != nil {
		logger.Fatal("unable to create client: %v", tag.NewErrorTag(err))
	}

	listener, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(portFlag))
	if err != nil {
		logger.Fatal("unable to create listener: %v", tag.NewErrorTag(err))
	}

	server := grpc.NewServer()
	handler, err := client.NewWorkflowServiceProxyServer(
		client.WorkflowServiceProxyOptions{Client: workflowClient},
	)
	if err != nil {
		logger.Fatal("unable to create service proxy: %v", tag.NewErrorTag(err))
	}

	workflowservice.RegisterWorkflowServiceServer(server, handler)

	err = server.Serve(listener)
	if err != nil {
		logger.Fatal("unable to serve: %v", tag.NewErrorTag(err))
	}
}
