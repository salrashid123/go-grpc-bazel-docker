// Forked from https://github.com/grpc/grpc-go.
/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package main

import (
	"fmt"
	"io"
	"log"
	"net"

	pb "helloworld"

	"github.com/google/uuid"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

const (
	port = ":50051"
)

var (
	hs *health.Server
)

// server is used to implement helloworld.GreeterServer.
type server struct{}
type healthServer struct{}

func (s *healthServer) Check(ctx context.Context, in *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	log.Printf("Handling grpc Check request: " + in.Service)
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

func (s *healthServer) Watch(in *healthpb.HealthCheckRequest, srv healthpb.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "Watch is not implemented")
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	uid, _ := uuid.NewUUID()
	msg := fmt.Sprintf("Hello %s  --> %s ", in.Name, uid.String())
	return &pb.HelloReply{Message: msg}, nil
}

func (s *server) SayHelloServerStream(in *pb.HelloRequest, stream pb.Greeter_SayHelloServerStreamServer) error {
	log.Println("Got SayHelloServerStream: Request ")
	for i := 0; i < 5; i++ {
		stream.Send(&pb.HelloReply{Message: "SayHelloServerStream Response"})
	}
	return nil
}

func (s server) SayHelloBiDiStream(srv pb.Greeter_SayHelloBiDiStreamServer) error {
	ctx := srv.Context()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		req, err := srv.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Printf("receive error %v", err)
			continue
		}
		log.Printf("Got SayHelloBiDiStream %s", req.Name)
		resp := &pb.HelloReply{Message: "SayHelloBiDiStream Server Response"}
		if err := srv.Send(resp); err != nil {
			log.Printf("send error %v", err)
		}
	}
}

func (s server) SayHelloClientStream(stream pb.Greeter_SayHelloClientStreamServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.HelloReply{Message: "SayHelloClientStream  Response"})
		}
		if err != nil {
			return err
		}
		log.Printf("Got SayHelloClientStream Request: %s", req.Name)
	}
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	sopts := []grpc.ServerOption{grpc.MaxConcurrentStreams(10)}
	s := grpc.NewServer(sopts...)
	pb.RegisterGreeterServer(s, &server{})

	healthpb.RegisterHealthServer(s, &healthServer{})

	log.Printf("Starting server...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
