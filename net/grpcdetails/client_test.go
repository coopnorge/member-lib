package grpcdetails

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

// TestGetClientAddress tests the GetClientAddress function.
func TestGetClientAddress(t *testing.T) {
	peerAddrStub := &peer.Peer{
		Addr: &net.TCPAddr{
			IP:   net.IPv4(127, 0, 0, 1),
			Port: 8080,
		},
	}

	ctx := peer.NewContext(context.Background(), peerAddrStub)

	address := GetClientAddress(ctx)
	expectedAddress := "127.0.0.1:8080"
	if address != expectedAddress {
		t.Errorf("GetClientAddress() = %v, want %v", address, expectedAddress)
	}

	// Test with a context without peer information
	emptyCtx := context.Background()
	if addr := GetClientAddress(emptyCtx); addr != "" {
		t.Errorf("GetClientAddress() = %v, want empty string", addr)
	}
}

// TestGetMetadataValue tests the GetMetadataValue function.
func TestGetMetadataValue(t *testing.T) {
	metaStub := metadata.MD{
		"key1": []string{"value1"},
	}

	ctx := metadata.NewIncomingContext(context.Background(), metaStub)

	value := GetMetadataValue(ctx, "key1")
	expectedValue := "value1"
	if value != expectedValue {
		t.Errorf("GetMetadataValue() = %v, want %v", value, expectedValue)
	}

	// Test with a key that doesn't exist
	if val := GetMetadataValue(ctx, "nonexistent"); val != "" {
		t.Errorf("GetMetadataValue() = %v, want empty string", val)
	}

	// Test with a context without metadata
	emptyCtx := context.Background()
	if val := GetMetadataValue(emptyCtx, "key1"); val != "" {
		t.Errorf("GetMetadataValue() = %v, want empty string", val)
	}
}
