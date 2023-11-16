package main

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	pb "github.com/amoeba/grpc_go_bench/dataservice"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	port    = flag.Int("port", 50051, "The server port")
	useTLS  = flag.Bool("tls", false, "Whether to use TLS or not")
	useMTLS = flag.Bool("mtls", false, "Whether to use mTLS or not")
)

type dataServer struct {
	pb.UnimplementedDataServiceServer
	ChunkSize int
	data      []byte
}

func makeServer(chunkSize int, dataSize int) *dataServer {
	log.Printf("makeServer() called, creating %d bytes of test data...\n", dataSize)
	s := &dataServer{ChunkSize: chunkSize}

	// Make/assign our data
	blob := make([]byte, dataSize)
	rand.Read(blob)
	s.data = blob

	log.Println("...done")

	return s
}

func (s *dataServer) GiveMeData(req *pb.DataRequest, stream pb.DataService_GiveMeDataServer) error {
	log.Printf("Streaming %d bytes of data\n", len(s.data))

	resp := &pb.DataResponse{}

	for currentByte := 0; currentByte < len(s.data); currentByte += s.ChunkSize {
		if currentByte+s.ChunkSize > len(s.data) {
			resp.Data = s.data[currentByte:len(s.data)]
		} else {
			resp.Data = s.data[currentByte : currentByte+s.ChunkSize]
		}
		log.Printf("Sending %d bytes...\n", len(resp.Data))
		if err := stream.Send(resp); err != nil {
			return err
		}
	}

	return nil
}

func runMainTLS() {
	log.Println("Setting up server in TLS mode.")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	creds, err := credentials.NewServerTLSFromFile("tls/server_cert.pem", "tls/server_key.pem")

	if err != nil {
		log.Fatalf("failed to create credentials: %v", err)
	}

	s := grpc.NewServer(grpc.Creds(creds))

	pb.RegisterDataServiceServer(s, makeServer(4*1000*1000, (1*1024+512)*1024*1024))

	log.Printf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func runMainMTLS() {
	log.Println("Setting up server in mTLS mode.")

	cert, err := tls.LoadX509KeyPair("tls/server_cert.pem", "tls/server_key.pem")

	if err != nil {
		log.Fatalf("failed to load key pair: %s", err)
	}

	ca := x509.NewCertPool()
	caFilePath := "tls/client_ca_cert.pem"
	caBytes, err := os.ReadFile(caFilePath)

	if err != nil {
		log.Fatalf("failed to read ca cert %q: %v", caFilePath, err)
	}

	if ok := ca.AppendCertsFromPEM(caBytes); !ok {
		log.Fatalf("failed to parse %q", caFilePath)
	}

	tlsConfig := &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{cert},
		ClientCAs:    ca,
	}

	s := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig)))
	pb.RegisterDataServiceServer(s, makeServer(4*1000*1000, (1*1024+512)*1024*1024))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	if err != nil {
		log.Fatalf("failed to create credentials: %v", err)
	}

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func runMain() {
	log.Println("Setting up server w/o TLS.")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	pb.RegisterDataServiceServer(s, makeServer(4*1000*1000, (1*1024+512)*1024*1024))

	log.Printf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
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
