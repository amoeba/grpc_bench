package main

import (
	"context"
	"flag"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
)

func getOne(client pb.DataServiceClient, ctx context.Context) float64 {
	start := time.Now().UnixMilli()

	stream, err := client.GiveMeData(ctx, &pb.DataRequest{})

	if err != nil {
		log.Fatalf("request failed: %v", err)
	}

	totalDataSize := 0

	for {
		d, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Fatal error while streaming response to client: %s", err)
		}
		totalDataSize += len(d.Data)
	}

	end := time.Now().UnixMilli()

	// Stats
	elapsed := end - start
	total := totalDataSize
	total_GB := float64(total) / 1024 / 1024 / 1024
	elpased_s := float64(end-start) / 1000
	throughput := total_GB / elpased_s

	log.Printf("got %d b in %d ms", total, elapsed)
	log.Printf("throughput: %f GiB/s", throughput)

	return throughput
}

func doStats(samples []float64) float64 {
	sum := float64(0)

	for _, sample := range samples {
		sum += sample
	}

	return sum / float64(len(samples))
}
func main() {
	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewDataServiceClient(conn)

	// Contact the server and print out its response.
	ctx := context.Background()

	n := 10
	samples := make([]float64, n)

	for i := range samples {
		samples[i] = getOne(c, ctx)
	}

	stat := doStats(samples)

	log.Printf("mean: %f GB/s", stat)

}
