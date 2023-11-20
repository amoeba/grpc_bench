FROM golang:1.21.4-bookworm

RUN apt update && apt install -y protobuf-compiler \
    protoc-gen-go \
    protoc-gen-go-grpc

ADD ./ /work
WORKDIR /work
RUN scripts/gen_certs.sh
RUN scripts/gen_proto.sh
WORKDIR /work/go-grpc
