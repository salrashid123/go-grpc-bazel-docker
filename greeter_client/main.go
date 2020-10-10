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
	"os"
	"time"

	pb "helloworld"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

const (
	address     = "127.0.0.1:50051"
	defaultName = "world"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}

	ctx := context.Background()

	// ******** HealthCheck

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	resp, err := healthpb.NewHealthClient(conn).Check(ctx, &healthpb.HealthCheckRequest{Service: "helloworld.GreeterServer"})
	if err != nil {
		log.Fatalf("HealthCheck failed %+v", err)
	}
	if resp.GetStatus() != healthpb.HealthCheckResponse_SERVING {
		log.Fatalf("service not in serving state: ", resp.GetStatus().String())
	}
	log.Printf("RPC HealthChekStatus:%v", resp.GetStatus())

	// ******** Unary Request
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Unary Request Response:  %s", r.Message)

	// ******** CLIENT Streaming

	cstream, err := c.SayHelloClientStream(context.Background())

	if err != nil {
		log.Fatalf("%v.SayHelloClientStream(_) = _, %v", c, err)
	}

	for i := 1; i < 5; i++ {
		if err := cstream.Send(&pb.HelloRequest{Name: fmt.Sprintf("client stream RPC %d ", i)}); err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("%v.Send(%v) = %v", cstream, i, err)
		}
	}

	creply, err := cstream.CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v, want %v", cstream, err, nil)
	}
	log.Printf(" Got SayHelloClientStream  [%s]", creply.Message)

	/// ***** SERVER Streaming
	stream, err := c.SayHelloServerStream(ctx, &pb.HelloRequest{Name: "Stream RPC msg"})
	if err != nil {
		log.Fatalf("SayHelloStream(_) = _, %v", err)
	}
	for {
		m, err := stream.Recv()
		if err == io.EOF {
			t := stream.Trailer()
			log.Println("Stream Trailer: ", t)
			break
		}
		if err != nil {
			log.Fatalf("SayHelloStream(_) = _, %v", err)
		}

		log.Printf("Message: [%s]", m.Message)
	}

	/// ********** BIDI Streaming

	done := make(chan bool)
	stream, err = c.SayHelloBiDiStream(context.Background())
	if err != nil {
		log.Fatalf("openn stream error %v", err)
	}
	ctx = stream.Context()

	go func() {
		for i := 1; i <= 10; i++ {
			req := pb.HelloRequest{Name: "Bidirectional CLient RPC msg "}
			if err := stream.SendMsg(&req); err != nil {
				log.Fatalf("can not send %v", err)
			}
		}
		if err := stream.CloseSend(); err != nil {
			log.Println(err)
		}
	}()

	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				close(done)
				return
			}
			if err != nil {
				log.Fatalf("can not receive %v", err)
			}
			log.Printf("Response: [%s] ", resp.Message)
		}
	}()

	go func() {
		<-ctx.Done()
		if err := ctx.Err(); err != nil {
			log.Println(err)
		}
		close(done)
	}()

	<-done

}
