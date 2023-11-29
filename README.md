# grpc_bench

Runnable tests of GRPC single stream throughput performance.

## Background

The basic question that lead to this repo was: How fast can GRPC send data and does TLS have an impact?

GRPC was designed for sending lots of smaller message (RPCs) over many streams and may not necessarily be optimized for sending large, multi gigabyte or lager messages over a single stream.
We also know that GRPC isn't just a single implementation so testing multiple implements might be useful or interesting.

## Methodology

For each implementation tested, a GRPC server and client were written implementing a single GRPC Service with a single, streaming RPC:

```
service DataService {
  rpc GiveMeData (DataRequest) returns (stream DataResponse) {}
}
```

For each implementation, before the server starts accepting requests, it pre-allocates a single data structure containing a bytes-like object full of random 64-bit integers. When the client executes the single streaming RPC, it reads from the stream, discarding the result, until the stream is exhausted.

Under each implementation, payload sizes of 512 MiB, 1 GiB, and 10 GiB were tested tested and the tests were run with TLS disabled, client-only TLS, and mutual TLS (mTLS). For each combination of payload size and TLS configuration, the average of ten runs were taken to calculate average throughput.

## Implementations

- [ ] Python (wraps C++)
  - [x] No TLS
  - [] TLS
  - [] mTLS
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
- Generate protobuf code, requires protoc (`protobuf-compiler`), `protoc-gen-go`, `protoc-gen-go-grpc`

    ```
    # On Debian
    apt install -y protobuf-compiler protoc-gen-go protoc-gen-go-grpc

    sh scripts/gen_proto.sh
    ```

### Running

#### Python GRPC

1. cd into `./python`
2. Create a virtualenv: `python -m venv .venv` an activate it
3. Install dependencies: `python -m pip install -r requirements.txt`


w/o TLS, 1GiB test size

- `python server --port 5000 --size 1073741824`
- `python client.py --address localhost:5000`

w/ TLS, 1GiB test size

TODO

w/ mTLS, 1GiB test size

TODO

#### Go GRPC

This benchmark is located in `./go-grpc`.

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

Tests were run with the followingn settings:

- Client and server both connecting over localhost
- GRPC's chunk size was left near its default of 4 MiB for all tests
- Throughput was calculated as the average of 10 RPCs


| Implementation | No TLS    | TLS       | mTLS       | Payload Size |
|----------------|-----------|-----------|------------|--------------|
| Go             | 3.4 GiB/s | 1.7 GiB/s | 1.6  GiB/s | 512 MiB      |
| Go             | 3.0 GiB/s | 1.7 GiB/s | 1.7  GiB/s | 1 GiB        |
| Go             | 2.7 GiB/s | 1.6 GiB/s | 1.5  GiB/s | 10 GiB       |
| Python         | 1.4 GiB/s | x         | x          | 512 MiB      |
| Python         | 1.4 GiB/s | x         | x          | 1 GiB        |
| Python         | 1.3       | x         | x          | 10 GiB       |
