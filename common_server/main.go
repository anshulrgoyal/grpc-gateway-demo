package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/felixge/httpsnoop"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/soheilhy/cmux"
	// importing generated stubs
	gen "grpc-gateway-demo/gen/go/hello"

	"google.golang.org/grpc"
)

// GreeterServerImpl will implement the service defined in protocol buffer definitions
type GreeterServerImpl struct {
	gen.UnimplementedGreeterServer
}

// SayHello is the implementation of RPC call defined in protocol definitions.
// This will take HelloRequest message and return HelloReply
func (g *GreeterServerImpl) SayHello(ctx context.Context, request *gen.HelloRequest) (*gen.HelloReply, error) {
	if err:=request.Validate();err!=nil {
		return nil,err
	}
	return &gen.HelloReply{
		Message: fmt.Sprintf("hello %s %s",request.Name,request.LastName),
	},nil
}

func main() {
	// create new gRPC server
	grpcSever := grpc.NewServer()
	// register the GreeterServerImpl on the gRPC server
	gen.RegisterGreeterServer(grpcSever, &GreeterServerImpl{})
	// creating mux for gRPC gateway. This will multiplex or route request different gRPC service
	mux:=runtime.NewServeMux()
	// setting up a dail up for gRPC service by specifying endpoint/target url
	err := gen.RegisterGreeterHandlerFromEndpoint(context.Background(), mux, "localhost:8081", []grpc.DialOption{grpc.WithInsecure()})
	if err != nil {
		log.Fatal(err)
	}
	// Creating a normal HTTP server
	server:=http.Server{
		Handler: withLogger(mux),
	}

	// creating a listener for server
	l,err:=net.Listen("tcp",":8081")
	if err!=nil {
		log.Fatal(err)
	}
	m := cmux.New(l)
	httpL := m.Match(cmux.HTTP1Fast())
	grpcL := m.Match(cmux.HTTP2())
	// start server
	go server.Serve(httpL)
	go grpcSever.Serve(grpcL)
	m.Serve()
}

func withLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		m:=httpsnoop.CaptureMetrics(handler,writer,request)
		log.Printf("http[%d]-- %s -- %s\n",m.Code,m.Duration,request.URL.Path)
	})
}
