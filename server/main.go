package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
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

// SayHello implements helloworld.GreeterServer
func (s *dataServer) GiveMeData(req *helloworld.DataRequest, stream pb.DataService_GiveMeDataServer) error {
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

func main() {
	flag.Parse()
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
