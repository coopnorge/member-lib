package grpcdetails

import (
	"context"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

// GetClientAddress extracts the IP address and port of the client from the given context.
// It returns the address as a string in the format "ip:port" or empty string if peer information not present in the context.
func GetClientAddress(ctx context.Context) string {
	if peerInfo, ok := peer.FromContext(ctx); ok {
		return peerInfo.Addr.String()
	}

	return ""
}

// GetMetadataValue retrieves a specific value associated with a given key from the request's metadata.
// If the metadata is available and the key is found, the first value associated with the key is returned.
// If the metadata is not present or the key is not found, an empty string will be returned.
func GetMetadataValue(ctx context.Context, key string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	values, found := md[key]
	if !found || len(values) == 0 {
		return ""
	}

	return values[0]
}
