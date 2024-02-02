# src/cpp/grpc

Benchmark for streaming performance of GRPC C++.
This codebase produces two executables, a GRPC server and a GRPC client.
When the server is running and the client is run, the client makes repeated streaming RPCs to the server and prints throughput numbers to stdout and finishes by calculating and printing an average throughput.

## Pre-requisites

- A relatively recent C++ compiler toolchain
- CMake
- Protobuf
- OpenSSL
- Abseil

For example, on a Debian system, you'll need the following packages:

- build-essential
- cmake
- libgrpc++-dev
- libprotobuf-dev
- openssl
- protobuf-compiler-grpc

which can be installed by running:

`apt-get install build-essential cmake libgrpc++-dev libprotobuf-dev openssl protobuf-compiler-grpc`

## Building

Assuming you've cloned or downloaded this repository and are in the `cpp-grpc` subdirectory, run:

```sh
mkdir build
cd build
cmake .. -DCMAKE_BUILD_TYPE=Release
make -j8
```

## Running

The server binds 0.0.0.0 on the host it's running on and listens on port 5000 by default.

The following commands run the benchmark with a payload size of 1GiB:

In one terminal, run:

`./server --size=1073741824`

In another terminal, change `hostname` below to the DNS name or IP address of the host where the server is running and run:

`./client --target hostname:5000`

When it runs, the client runs the benchmark a default of 10 times and prints throughput for each run and a final average after all runs have been completed.
