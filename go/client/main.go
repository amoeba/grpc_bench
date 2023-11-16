package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io"
	"log"
	"os"
	"time"

	pb "github.com/amoeba/grpc_go_bench/dataservice"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr     = flag.String("addr", "localhost:50051", "the address to connect to")
	nSamples = flag.Int("nsamples", 10, "Number of samples to take")
	useTLS   = flag.Bool("tls", false, "Whether to use TLS or not")
	useMTLS  = flag.Bool("mtls", false, "Whether to use mTLS or not")
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

func runMainTLS() {
	log.Println("Running client in TLS mode.")

	creds, err := credentials.NewClientTLSFromFile("tls/ca_cert.pem", "x.test.example.com")

	if err != nil {
		log.Fatalf("failed to load credentials: %v", err)
	}

	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewDataServiceClient(conn)

	ctx := context.Background()

	n := 10
	samples := make([]float64, n)

	for i := range samples {
		samples[i] = getOne(c, ctx)
	}

	stat := doStats(samples)
	log.Printf("mean: %f GB/s", stat)
}

func runMainMTLS() {
	log.Println("Running client in mTLS mode.")

	cert, err := tls.LoadX509KeyPair("tls/client_cert.pem", "tls/client_key.pem")
	if err != nil {
		log.Fatalf("failed to load client cert: %v", err)
	}

	ca := x509.NewCertPool()
	caFilePath := "tls/ca_cert.pem"
	caBytes, err := os.ReadFile(caFilePath)

	if err != nil {
		log.Fatalf("failed to read ca cert %q: %v", caFilePath, err)
	}

	if ok := ca.AppendCertsFromPEM(caBytes); !ok {
		log.Fatalf("failed to parse %q", caFilePath)
	}

	tlsConfig := &tls.Config{
		ServerName:   "x.test.example.com",
		Certificates: []tls.Certificate{cert},
		RootCAs:      ca,
	}

	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewDataServiceClient(conn)

	ctx := context.Background()

	n := 10
	samples := make([]float64, n)

	for i := range samples {
		samples[i] = getOne(c, ctx)
	}

	stat := doStats(samples)
	log.Printf("mean: %f GB/s", stat)
}

func runMain() {
	log.Println("Running client w/o TLS.")

	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer conn.Close()

	c := pb.NewDataServiceClient(conn)

	ctx := context.Background()
	samples := make([]float64, *nSamples)

	for i := range samples {
		samples[i] = getOne(c, ctx)
	}

	stat := doStats(samples)
	log.Printf("mean: %f GB/s", stat)
}

func main() {
	flag.Parse()

	if *useTLS {
		runMainTLS()
	} else if *useMTLS {
		runMainMTLS()
	} else {
		runMain()
	}

}
