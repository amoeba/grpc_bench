FROM debian@sha256:133a1f2aa9e55d1c93d0ae1aaa7b94fb141265d0ee3ea677175cdb96f5f990e5

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    cmake \
    libgrpc++-dev \
    libprotobuf-dev \
    openssl\
    protobuf-compiler-grpc

ADD . /work
WORKDIR /work

# Generate certs
RUN ./scripts/gen_certs.sh

# Create build dir
WORKDIR /work/src/cpp/grpc
RUN mkdir -p build
WORKDIR /work/src/cpp/grpc/build
RUN cmake -DCMAKE_BUILD_TYPE=Release ..
RUN make
