package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	gen "grpc-gateway-demo/gen/go/hello"
)

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

func main() {
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
	err := gen.RegisterGreeterHandlerFromEndpoint(context.Background(), mux, "localhost:8080", []grpc.DialOption{grpc.WithInsecure()})
	if err != nil {
		log.Fatal(err)
	}
	// Creating a normal HTTP server
	server:=gin.New()
	server.Use(gin.Logger())
	server.Group("v1/*{grpc_gateway}").Any("",gin.WrapH(mux))

	server.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK,"Ok")
	})


	// start server
	err = server.Run(":8081")
	if err != nil {
		log.Fatal(err)
	}
}

