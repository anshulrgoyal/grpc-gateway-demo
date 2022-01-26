package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/felixge/httpsnoop"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc/metadata"

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
	mux:=runtime.NewServeMux(
		// convert header in response(going from gateway) from metadata received.
		runtime.WithOutgoingHeaderMatcher(isHeaderAllowed),
		runtime.WithMetadata(func(ctx context.Context, request *http.Request) metadata.MD {
			header:=request.Header.Get("Authorization")
			// send all the headers received from the client
			md:=metadata.Pairs("auth",header)
			return md
		}),
		runtime.WithErrorHandler(func(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, writer http.ResponseWriter, request *http.Request, err error) {
			//creating a new HTTTPStatusError with a custom status, and passing error
			newError:=runtime.HTTPStatusError{
				HTTPStatus: 400,
				Err:        err,
			}
			// using default handler to do the rest of heavy lifting of marshaling error and adding headers
			runtime.DefaultHTTPErrorHandler(ctx,mux,marshaler,writer,request,&newError)
		}))
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

var allowedHeaders=map[string]struct{}{
	"x-request-id": {},
}

func isHeaderAllowed(s string)( string,bool) {
	// check if allowedHeaders contain the header
	if _,isAllowed:=allowedHeaders[s];isAllowed {
		// send uppercase header
		return strings.ToUpper(s),true
	}
	// if not in allowed header, don't send the header
	return s,false
}
