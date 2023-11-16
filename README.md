# grpc_go_bench

Runnable tests of GRPC single stream throughput performance.

The basic questions the lead to this repo are:

1. What's GRPC's actual single-stream maximum throughput for large (GiB+) messages?
2. What settings might impact that throughput?

    For example:

    - GRPC message size limits (i.e. you can try to send a large message but GRPC will always chunk it for you)
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
