# gRPC Details

Package provides utilities for extracting detailed client information in gRPC contexts.
It includes various functions to obtain additional information form gRPC or work with gRPC requests.

## Example

You can have a middleware where from gRPC request can be extracted client details.

```go
package middleware

import (
  "context"

  "github.com/coopnorge/member-lib/net/grpcdetails"
  "google.golang.org/grpc"
)

func (m *MyMiddleware) UnaryServerInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, nextHandler grpc.UnaryHandler) (resp any, err error) {
  cIP := grpcdetails.GetClientAddress(ctx)
  cAgent := grpcdetails.GetMetadataValue(ctx, "User-Agent")
  
  // another internal code logic to use it...

  return nextHandler(ctx, req)
}
```
