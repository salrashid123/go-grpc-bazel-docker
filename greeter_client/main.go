package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/salrashid123/go-grpc-bazel-docker/echo"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

var (
	address         = flag.String("host", "localhost:50051", "host:port of gRPC server")
	skipHealthCheck = flag.Bool("skipHealthCheck", false, "Skip Initial Healthcheck")
)

const (
	defaultName = "world"
)

func main() {
	flag.Parse()

	// Set up a connection to the server.
	var err error
	var conn *grpc.ClientConn

	conn, err = grpc.Dial(*address, grpc.WithInsecure())

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := echo.NewEchoServerClient(conn)
	ctx := context.Background()

	// how to perform healthcheck request manually:
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if !*skipHealthCheck {
		resp, err := healthpb.NewHealthClient(conn).Check(ctx, &healthpb.HealthCheckRequest{Service: "echo.EchoServer"})
		if err != nil {
			log.Fatalf("HealthCheck failed %+v", err)
		}

		if resp.GetStatus() != healthpb.HealthCheckResponse_SERVING {
			log.Fatalf("service not in serving state: ", resp.GetStatus().String())
		}
		log.Printf("RPC HealthChekStatus: %v\n", resp.GetStatus())
	}

	// ******** Unary Request
	r, err := c.SayHelloUnary(ctx, &echo.EchoRequest{Name: defaultName})
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
		if err := cstream.Send(&echo.EchoRequest{Name: fmt.Sprintf("client stream RPC %d ", i)}); err != nil {
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
	stream, err := c.SayHelloServerStream(ctx, &echo.EchoRequest{Name: "Stream RPC msg"})
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
		log.Fatalf("open stream error %v", err)
	}
	ctx = stream.Context()

	var wg sync.WaitGroup

	go func() {
		for i := 1; i <= 10; i++ {
			wg.Add(1)
			go func(ctr int) {
				defer wg.Done()
				req := echo.EchoRequest{Name: fmt.Sprintf("CLient RPC msg %d", ctr)}
				if err := stream.SendMsg(&req); err != nil {
					log.Fatalf("can not send %v", err)
				}
			}(i)
		}
		wg.Wait()
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
