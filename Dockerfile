FROM debian@sha256:133a1f2aa9e55d1c93d0ae1aaa7b94fb141265d0ee3ea677175cdb96f5f990e5

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    cmake \
    libgrpc++-dev \
    libprotobuf-dev \
    openssl\
    protobuf-compiler-grpc

ADD ./cpp-grpc /work/cpp-grpc
ADD ./tls /work/tls
ADD ./protos /work/protos
ADD ./scripts /work/scripts

WORKDIR /work

RUN scripts/gen_certs.sh

RUN mkdir -p ./cpp-grpc/build
WORKDIR /work/cpp-grpc/build
RUN cmake .. -DCMAKE_BUILD_TYPE=Release
RUN make
