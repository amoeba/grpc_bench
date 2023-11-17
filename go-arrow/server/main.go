// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
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
	log.Println("DoGet...")
	writer := flight.NewRecordWriter(fs, ipc.WithSchema(f.Table.Schema()))

	reader := array.NewTableReader(f.Table, 1)
	for {
		if !reader.Next() {
			break
		}
		writer.Write(reader.Record())
	}

	log.Println("...DoGet.")

	return nil
}

func main() {
	s := grpc.NewServer()

	tbl := CreateTable(1_000_000)
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

	data, err := fc.DoGet(context.Background(), &flight.Ticket{})

	if err != nil {
		log.Fatalf("Error starting DoGet: %s", err)
	}

	reader, err := flight.NewRecordReader(data)
	if err != nil {
		log.Fatalf("Failed to create new RecordReader: %s", err)
	}

	var numRows int64 = 0

	for {
		if !reader.Next() {
			break
		}
		rec := reader.Record()

		numRows += rec.NumRows()
		log.Printf("%d", numRows)
	}

	fmt.Println(reader.Schema())
	fmt.Printf("Read %d rows\n", numRows)
}
