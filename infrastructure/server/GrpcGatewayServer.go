package server

import (
	core "authz/api/gen/v1alpha"
	"authz/domain/contracts"
	"authz/domain/model"
	"context"
	"errors"
	"github.com/golang/glog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net"
	"net/http"
	"os"
	"sync"
)

// GrpcGatewayServer represents a GrpcServer host service
type GrpcGatewayServer struct{}

// NewServer creates a new Server object to use.
func (r *GrpcGatewayServer) NewServer() contracts.Server {
	return &GrpcGatewayServer{}
}

// Serve exposes a GRPC endpoint and blocks until processing ends, at which point the waitgroup is signalled. This should be run as a goroutine.
func (r *GrpcGatewayServer) Serve(host string, handler http.HandlerFunc, wait *sync.WaitGroup) error {
	defer wait.Done()

	ls, err := net.Listen("tcp", ":"+host)

	if err != nil {
		glog.Errorf("Error opening TCP port: %s", err)
		return err
	}

	var creds credentials.TransportCredentials = nil
	if _, err = os.Stat("/etc/tls/tls.crt"); err == nil {
		if _, err := os.Stat("/etc/tls/tls.key"); err == nil { //Cert and key exists start server in TLS mode
			glog.Info("TLS cert and Key found  - Starting gRPC server in secure TLS mode")

			creds, err = credentials.NewServerTLSFromFile("/etc/tls/tls.crt", "/etc/tls/tls.key")
			if err != nil {
				glog.Errorf("Error loading certs: %s", err)
				return err
			}
		}
	} else { // For all cases of error - we start a plain HTTP server
		glog.Info("TLS cert or Key not found  - Starting gRPC server in insecure mode")
	}

	srv := grpc.NewServer(grpc.Creds(creds))
	core.RegisterCheckPermissionServer(srv, r)
	err = srv.Serve(ls)
	if err != nil {
		glog.Errorf("Error hosting gRPC service: %s", err)
		return err
	}
	return nil
}

// CheckPermission processes an authorization check and returns whether or not the operation would be allowed
func (r *GrpcGatewayServer) CheckPermission(ctx context.Context, rpcReq *core.CheckPermissionRequest) (*core.CheckPermissionResponse, error) {

	return &core.CheckPermissionResponse{Result: true}, nil
}

func getRequestorIdentityFromContext(ctx context.Context) model.Principal {
	for _, name := range []string{"grpcgateway-authorization", "bearer-token"} {
		if metadata, ok := metadata.FromIncomingContext(ctx); ok {
			headers := metadata.Get(name)
			if len(headers) > 0 {
				return model.NewPrincipal(headers[0])
			}
		}
	}

	return model.NewAnonymousPrincipal()
}

func convertDomainErrorToGrpc(err error) error {
	switch {
	case errors.Is(err, ErrNotAuthenticated):
		return status.Error(codes.Unauthenticated, "Anonymous access is not allowed.")
	case errors.Is(err, ErrNotAuthorized):
		return status.Error(codes.PermissionDenied, "Access denied.")
	default:
		return status.Error(codes.Unknown, "Internal server error.")
	}
}
