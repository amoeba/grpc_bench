#!/bin/sh

cd go || exit

protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    dataservice/dataservice.proto