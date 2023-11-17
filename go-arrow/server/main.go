package main

import (
	"context"
	"flag"
	"log"
	"net"

	"github.com/apache/arrow/go/v14/arrow"
	"github.com/apache/arrow/go/v14/arrow/array"
	"github.com/apache/arrow/go/v14/arrow/flight"
	"github.com/apache/arrow/go/v14/arrow/ipc"
	"github.com/apache/arrow/go/v14/arrow/memory"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	nrows     = flag.Int64("nrows", 1024, "Number of rows of test data to use")
	chunkSize = flag.Int64("chunkSize", 1048576, "Chunk size to split test data into")
)

type MyFlightServer struct {
	flight.BaseFlightServer
	Table arrow.Table
}

func CreateTable(nrows int64) arrow.Table {
	log.Printf("Creating table with %d rows...", nrows)

	pool := memory.NewGoAllocator()

	schema := arrow.NewSchema(
		[]arrow.Field{
			{Name: "a", Type: arrow.PrimitiveTypes.Int64},
		},
		nil,
	)

	builder := array.NewInt64Builder(pool)
	defer builder.Release()

	for i := int64(0); i < nrows; i++ {
		builder.Append(i)
	}

	arr := builder.NewInt64Array()
	defer arr.Release()
	col := arrow.NewColumnFromArr(schema.Field(0), arr)
	defer col.Release()
	tbl := array.NewTable(schema, []arrow.Column{col}, nrows)
	defer tbl.Release()

	// Is this right? Retain to bump the ref count so the deferred release
	// doesn't free the table?
	tbl.Retain()

	log.Println("...done creating table.")
	return tbl
}

func (f *MyFlightServer) DoGet(ticket *flight.Ticket, fs flight.FlightService_DoGetServer) error {
	log.Println("Starting handling DoGet...")
	writer := flight.NewRecordWriter(fs, ipc.WithSchema(f.Table.Schema()))

	// TODO: What do with chunkSize param here
	// -1 or 0 seemingly isn't good
	// Something smaller like 1024 is fine
	reader := array.NewTableReader(f.Table, *chunkSize)
	nchunks := 0
	for reader.Next() {
		writer.Write(reader.Record())
		nchunks += 1
	}

	log.Printf("Wrote Table out in %d chunk(s).", nchunks)
	log.Println("...done serving DoGet.")

	return nil
}

func main() {
	flag.Parse()

	s := grpc.NewServer()

	// Create a Table for streaming to clients
	tbl := CreateTable(*nrows) // 100mil rows
	defer tbl.Release()

	flightServer := MyFlightServer{Table: tbl}
	flight.RegisterFlightServiceServer(s, &flightServer)

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Fatalf("Failed to listen: %s", err)
	}
	go s.Serve(lis)
	defer s.Stop()

	conn, err := grpc.DialContext(context.Background(), lis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial: %s", err)
	}
	defer conn.Close()

	fc := flight.NewClientFromConn(conn, nil)

	log.Println("Client requesting DoGet()...")
	stream, err := fc.DoGet(context.Background(), &flight.Ticket{})

	if err != nil {
		log.Fatalf("Error starting DoGet: %s", err)
	}

	reader, err := flight.NewRecordReader(stream)
	if err != nil {
		log.Fatalf("Failed to create new RecordReader: %s", err)
	}
	defer reader.Release()

	var numRecords int64 = 0
	var numRows int64 = 0

	// TODO: When chunkSize is 0 or -1, reader.Next() never evaluates to true
	for reader.Next() {
		record := reader.Record()
		numRecords += 1
		numRows += record.NumRows()
	}

	log.Println("...client finished requesting DoGet().")
	log.Printf("Read %d row(s) over %d records(s)\n", numRows, numRecords)
}
