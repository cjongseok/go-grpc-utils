package grpcutils

import (
	"context"
	"fmt"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"time"
)

// normalRetryDelay is a desired delay among HTTP requests to the same API
const normalRetryDelay = 4 * time.Second

// minRetryDelay is a min delay among HTTP requests to the same API
const minRetryDelay = 1 * time.Second

func checkHealth(c healthpb.HealthClient, service string, timeout time.Duration) (health *healthpb.HealthCheckResponse, err error) {
	//c := healthpb.NewHealthClient(cc)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	health, err = c.Check(ctx, &healthpb.HealthCheckRequest{Service: service})
	if err != nil {
		health = nil
	}
	return
}

type healthCondition func(response *healthpb.HealthCheckResponse) bool

func checkHealthUntilDesiredCondition(c healthpb.HealthClient, service string, timeout time.Duration, stop healthCondition) (health *healthpb.HealthCheckResponse, err error) {
	deadline := time.Now().Add(timeout)
	for {
		now := time.Now()
		if now.Before(deadline) {
			health, err = checkHealth(c, service, deadline.Sub(now))
		} else { // time's up
			if err == nil {
				err = fmt.Errorf("timeout")
			}
			health = nil
			return
		}

		// delay the next try
		if err == nil && stop(health) {
			return
		}
		remain := deadline.Sub(time.Now())
		if remain > normalRetryDelay {
			time.Sleep(normalRetryDelay)
		} else if remain > minRetryDelay {
			time.Sleep(minRetryDelay)
		}
	}
}

// WaitForHealth awaits for a health of a gRPC service until time's up.
func WaitForHealth(c healthpb.HealthClient, service string, timeout time.Duration) (health *healthpb.HealthCheckResponse, err error) {
	stop := func(response *healthpb.HealthCheckResponse) bool {
		if response == nil {
			return false
		}
		return true
	}
	return checkHealthUntilDesiredCondition(c, service, timeout, stop)
}

// WaitForHealthy awaits for a gRPC service healthy
func WaitForHealthy(c healthpb.HealthClient, service string, timeout time.Duration) error {
	stop := func(response *healthpb.HealthCheckResponse) bool {
		if response != nil && response.Status == healthpb.HealthCheckResponse_SERVING {
			return true
		}
		return false
	}
	_, err := checkHealthUntilDesiredCondition(c, service, timeout, stop)
	return err
}
