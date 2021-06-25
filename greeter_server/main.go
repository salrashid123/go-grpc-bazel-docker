package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/salrashid123/go-grpc-bazel-docker/echo"

	"github.com/google/uuid"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

var (
	grpcport = flag.String("grpcport", ":50051", "grpcport")
	tlsCert  = flag.String("tlsCert", "", "tls Certificate")
	tlsKey   = flag.String("tlsKey", "", "tls Key")
	insecure = flag.Bool("insecure", false, "startup without TLS")
	hs       *health.Server
)

// server is used to implement echo.EchoServer.
type server struct{}
type healthServer struct{}

func (s *healthServer) Check(ctx context.Context, in *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	log.Printf("Handling grpc Check request: " + in.Service)
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

func (s *healthServer) Watch(in *healthpb.HealthCheckRequest, srv healthpb.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "Watch is not implemented")
}

func (s *server) SayHelloUnary(ctx context.Context, in *echo.EchoRequest) (*echo.EchoReply, error) {
	log.Println("Got Unary Request: ")
	uid, _ := uuid.NewUUID()
	return &echo.EchoReply{Message: "SayHelloUnary Response " + uid.String()}, nil
}

func (s *server) SayHelloServerStream(in *echo.EchoRequest, stream echo.EchoServer_SayHelloServerStreamServer) error {
	log.Println("Got SayHelloServerStream: Request ")
	for i := 0; i < 5; i++ {
		stream.Send(&echo.EchoReply{Message: "SayHelloServerStream Response"})
	}
	return nil
}

func (s server) SayHelloBiDiStream(srv echo.EchoServer_SayHelloBiDiStreamServer) error {
	//ctx := srv.Context()
	for {
		req, err := srv.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Printf("receive error %v", err)
			continue
		}
		log.Printf("Got SayHelloBiDiStream %s", req.Name)
		resp := &echo.EchoReply{Message: fmt.Sprintf("SayHelloBiDiStream Server Response for request [%s]", req.Name)}
		if err := srv.Send(resp); err != nil {
			log.Printf("send error %v", err)
		}
	}
}

func (s server) SayHelloClientStream(stream echo.EchoServer_SayHelloClientStreamServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&echo.EchoReply{Message: "SayHelloClientStream  Response"})
		}
		if err != nil {
			return err
		}
		log.Printf("Got SayHelloClientStream Request: %s", req.Name)
	}
}

func main() {
	flag.Parse()
	if *grpcport == "" {
		flag.Usage()
		log.Fatalf("missing -grpcport flag (:50051)")
	}

	lis, err := net.Listen("tcp", *grpcport)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	sopts := []grpc.ServerOption{}
	if *insecure == false {
		if *tlsCert == "" || *tlsKey == "" {
			log.Fatalf("Must set --tlsCert and tlsKey if --insecure flags is not set")
		}
		ce, err := credentials.NewServerTLSFromFile(*tlsCert, *tlsKey)
		if err != nil {
			log.Fatalf("Failed to generate credentials %v", err)
		}
		sopts = append(sopts, grpc.Creds(ce))
	}

	s := grpc.NewServer(sopts...)
	echo.RegisterEchoServerServer(s, &server{})

	healthpb.RegisterHealthServer(s, &healthServer{})

	log.Printf("Starting server...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
