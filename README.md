go-grpc-utils
===
gRPC utils for Go

Installing
---
```
go get github.com/cjongseok/go-grpc-utils
```

Utils
---
### WaitForHealth(c healthpb.HealthClient, service string, timeout time.Duration) (*healthpb.HealthCheckResponse, error)
It waits for a health of a gRPC service until time's up, and returns the service status. HealthCheckResponse, a gRPC standard health check response, is explained at https://github.com/grpc/grpc/blob/master/doc/health-checking.md. Its Go implementation is in [google.golang.org/grpc/health/grpc_health_v1](https://godoc.org/google.golang.org/grpc/health/grpc_health_v1) package.

### WaitForHealthy(c healthpb.HealthClient, service string, timeout time.Duration) error
It waits for the gRPC service is healthy until time's up.
