package api

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func UnaryAuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if err := authorize(ctx, info.FullMethod); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	return handler(ctx, req)
}

func StreamAuthInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if err := authorize(stream.Context(), info.FullMethod); err != nil {
		return status.Error(codes.PermissionDenied, err.Error())
	}
	return handler(srv, stream)
}

func authorize(ctx context.Context, method string) error {
	peer, ok := peer.FromContext(ctx)
	if !ok {
		return errors.New("Error to read peer information")
	}

	tlsInfo, ok := peer.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return errors.New("Error to get auth information")
	}

	certs := tlsInfo.State.VerifiedChains
	if len(certs) == 0 || len(certs[0]) == 0 {
		return errors.New("Missing certificate chain")
	}

	var roles []string
	for _, ext := range certs[0][0].Extensions {
		if oid := OidToString(ext.Id); IsOidRole(oid) {
			roles = ParseRoles(string(ext.Value))
			break
		}
	}

	if !HasPermission(method, roles) {
		return errors.New("Unauthorized")
	}
	return nil
}
