go-grpc-utils
===
gRPC utils for Go

Utils
---
### WaitForHealth(c healthpb.HealthClient, service string, timeout time.Duration) (*healthpb.HealthCheckResponse, error)
It waits for a health of a gRPC service until time's up.

### WaitForHealthy(c healthpb.HealthClient, service string, timeout time.Duration) error
It waits for the gRPC service is healthy until time's up.
