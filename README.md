# grpc_go_bench

Runnable tests of GRPC single stream throughput performance.

The basic questions the lead to this repo are:

1. What's GRPC's actual single-stream maximum throughput for large (GiB+) messages?
2. What settings might impact that throughput?

    For example:

    - GRPC message size limits (i.e. you can try to send a large message but GRPC will error unless you chunk the message)
    - Use of TLS, mTLS, and any related settings. Does TLS have an impact?

GRPC was designed for sending lots of smaller message (RPCs) over many streams and may not necessarily be optimized for sending large, multi gigabyte or lager messages over a single stream.
GRPC also isn't just one implementation so a real test would test various implementations.

## Implementations

- [ ] Python (wraps C++)
- [x] Go
    - [x] No TLS
    - [x] TLS
    - [x] mTLS

## Running tests

### Pre-requisites

- Generate TLS certs, CA, requires OpenSSL

    ```sh
    sh scripts/gen_certs.sh
    ```
- Generate protobuf code, requires protoc and possibly a Golang GRPC package or two

    ```
    sh scripts/gen_proto.sh
    ```

### Running

#### Go

This tests a GRPC client streaming a single RPC from the server containing a variable-size payload of bytes.

Run the server and client commands in separate terminals. When each server starts up, it will create a `[]byte` (and fill it with random data) so beware of how much memory you want to use.

w/o TLS, 1 GiB test size

- `go run server/main.go -port 5000 -size 1073741824`
- `go run client/main.go -addr localhost:5000`

w/ TLS, 1 GiB test size

- `go run server/main.go -tls -port 5001 -size 1073741824`
- `go run client/main.go -tls -addr localhost:5001`

w/ mTLS, 1 GiB test size

- `go run server/main.go -mtls -port 5002 -size 1073741824`
- `go run client/main.go -mtls -addr localhost:5002`

### Results

Tests were run over localhost, and each RPC was run ten times and average throughput was calculated from that, chunk size was 4MB. Warmup would help the throughput numbers a bit.

| Stream Size | Throughput (GiB/s) | Config  |
|-------------|--------------------|---------|
| 512 MiB     | 3.4                | w/o TLS |
| 512 MiB     | 1.7                | TLS     |
| 512 MiB     | 1.6                | mTLS    |
| 1 GiB       | 3.0                | w/o TLS |
| 1 GiB       | 1.7                | TLS     |
| 1 GiB       | 1.7                | mTLS    |
| 10 GiB      | 2.7                | w/o TLS |
| 10 GiB      | 1.6                | TLS     |
| 10 GiB      | 1.5                | mTLS    |
