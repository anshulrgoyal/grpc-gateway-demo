package main

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/felixge/httpsnoop"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	gen "grpc-gateway-demo/gen/go/hello"
)

func withLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		m:=httpsnoop.CaptureMetrics(handler,writer,request)
		log.Printf("http[%d]-- %s -- %s\n",m.Code,m.Duration,request.URL.Path)
	})
}

func main() {
	// creating mux for gRPC gateway. This will multiplex or route request different gRPC service
	mux:=runtime.NewServeMux(
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
	err := gen.RegisterGreeterHandlerFromEndpoint(context.Background(), mux, "localhost:8080", []grpc.DialOption{grpc.WithInsecure()})
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
	// start server
	err = server.Serve(l)
	if err != nil {
		log.Fatal(err)
	}
}
