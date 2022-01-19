package main

import (
	"context"
	"fmt"
	"log"
	"net"

	// importing generated stubs
	grpc_gateway_demo "grpc-gateway-demo/gen/go/hello"

	"google.golang.org/grpc"
)

// GreeterServerImpl will implement the service defined in protocol buffer definitions
type GreeterServerImpl struct {
	grpc_gateway_demo.UnimplementedGreeterServer
}

// SayHello is the implementation of RPC call defined in protocol definitions.
// This will take HelloRequest message and return HelloReply
func (g *GreeterServerImpl) SayHello(ctx context.Context, request *grpc_gateway_demo.HelloRequest) (*grpc_gateway_demo.HelloReply, error) {
	return &grpc_gateway_demo.HelloReply{
		Message: fmt.Sprintf("hello %s",request.Name),
	},nil
}

func main() {
	// create new gRPC server
	server := grpc.NewServer()
	// register the GreeterServerImpl on the gRPC server
	grpc_gateway_demo.RegisterGreeterServer(server, &GreeterServerImpl{})
	// start listening on port :8080 for a tcp connection
	if l, err := net.Listen("tcp", ":8080"); err != nil {
		log.Fatal("error in listening on port :8080", err)
	} else {
		// the gRPC server
		if err:=server.Serve(l);err!=nil {
			log.Fatal("unable to start server",err)
		}
	}
}
