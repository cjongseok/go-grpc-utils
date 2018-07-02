package grpcutils

import (
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"net"
	"testing"
	"time"
)

const port = 8081
const normalService = "normal service"
const stuckService = "stuck service"

type testServer struct {
	grpcServer   *grpc.Server
	healthServer *health.Server
}

func startServer(port int) (*testServer, error) {
	healthServer := health.NewServer()
	healthServer.SetServingStatus(normalService, healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus(stuckService, healthpb.HealthCheckResponse_NOT_SERVING)
	grpcServer := grpc.NewServer([]grpc.ServerOption{}...)
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("listen failure; %v", err)
	}
	s := &testServer{
		grpcServer,
		healthServer,
	}
	go func() {
		err := grpcServer.Serve(lis)
		if err != nil {
			fmt.Println("grpc server error;", err)
		}
		stopServer(s)
	}()
	return s, nil
}

func stopServer(s *testServer) {
	s.grpcServer.Stop()
}

func beforeTest(t *testing.T) (s *testServer, c healthpb.HealthClient) { // spawns a server
	var err error

	// spawns a server
	s, err = startServer(port)
	if err != nil {
		fmt.Println("server failure;", err)
		t.FailNow()
		return
	}

	// spawns a client
	var cc *grpc.ClientConn
	addr := fmt.Sprintf("localhost:%d", port)
	cc, err = grpc.Dial(addr, []grpc.DialOption{grpc.WithInsecure()}...)
	if err != nil {
		fmt.Println("dial failure:", err)
		t.FailNow()
		return
	}
	c = healthpb.NewHealthClient(cc)
	return
}

func TestWaitForHealthToHealthyServer(t *testing.T) {
	const timeout = 3 * time.Second
	s, c := beforeTest(t)
	defer stopServer(s)

	// wait for health
	start := time.Now()
	health, err := WaitForHealth(c, normalService, timeout)
	done := time.Now()
	fmt.Println("elapsed:", done.Sub(start))
	if err != nil {
		fmt.Println(err)
		t.FailNow()
		return
	}
	fmt.Println("health:", health)
}

type jitter struct {
	after  time.Duration
	status healthpb.HealthCheckResponse_ServingStatus
}

func jitterServer(s *testServer, service string, jitters chan jitter) *testServer {
	go func() {
		for j := range jitters {
			<-time.After(j.after)
			s.healthServer.SetServingStatus(service, j.status)
		}
	}()
	return s
}

func TestWaitForHealthToJitterServer(t *testing.T) {
	const newservice = "new service"
	const timeout = 10 * time.Second
	const registrationDelay = 5 * time.Second
	const desiredMinDelay = registrationDelay
	s, c := beforeTest(t)
	defer stopServer(s)
	jittering := make(chan jitter, 1)
	s = jitterServer(s, newservice, jittering)
	jittering <- jitter{registrationDelay, healthpb.HealthCheckResponse_NOT_SERVING}
	close(jittering)

	// wait for the health of new service
	start := time.Now()
	health, err := WaitForHealth(c, newservice, timeout)
	done := time.Now()
	elapsed := done.Sub(start)
	fmt.Println("elapsed:", elapsed)
	fmt.Println("health:", health)
	if err != nil {
		fmt.Println(err)
		t.FailNow()
		return
	}
	if elapsed < desiredMinDelay {
		fmt.Println("jitter timing error")
		t.FailNow()
		return
	}
}

func TestWaitForHealthyToHealthyServer(t *testing.T) {
	const timeout = 3 * time.Second
	s, c := beforeTest(t)
	defer stopServer(s)

	// wait for health
	start := time.Now()
	err := WaitForHealthy(c, normalService, timeout)
	done := time.Now()
	fmt.Println("elapsed:", done.Sub(start))
	if err != nil {
		fmt.Println(err)
		t.FailNow()
		return
	}
}

func TestWaitForHealthyToJitterServer(t *testing.T) {
	const newservice = "new service"
	const timeout = 10 * time.Second
	const desiredMinDelay = 6 * time.Second
	s, c := beforeTest(t)
	defer stopServer(s)
	jittering := make(chan jitter, 2)
	s = jitterServer(s, newservice, jittering)
	jittering <- jitter{3 * time.Second, healthpb.HealthCheckResponse_NOT_SERVING}
	jittering <- jitter{3 * time.Second, healthpb.HealthCheckResponse_SERVING}
	close(jittering)

	// wait for the health of new service
	start := time.Now()
	err := WaitForHealthy(c, newservice, timeout)
	done := time.Now()
	elapsed := done.Sub(start)
	fmt.Println("elapsed:", elapsed)
	if err != nil {
		fmt.Println(err)
		t.FailNow()
		return
	}
	if elapsed < desiredMinDelay {
		fmt.Println("jitter timing error")
		t.FailNow()
		return
	}
}
