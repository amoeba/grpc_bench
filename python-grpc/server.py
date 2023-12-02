import asyncio
import logging
import argparse
import os

import numpy as np
import grpc

import dataservice.dataservice_pb2
import dataservice.dataservice_pb2_grpc


def read(path):
    with open(path, "rb") as f:
        return f.read()


CHUNK_SIZE = 4 * 1000 * 1000  # near 4MiB max but same as Go impl


def gen_random(length=int) -> np.array:
    return np.random.randint(0, length, length, dtype=np.dtype(np.int64))


class Greeter(dataservice.dataservice_pb2_grpc.DataServiceServicer):
    def __init__(self, size: int):
        self.payload_length = size
        self.set_payload()

        super()

    def set_payload(self):
        print(f"Creating payload of {self.payload_length} bytes")
        # Divide by eight to account for 64-bit ints
        self.payload = gen_random(int(self.payload_length / 8)).tobytes()
        print("Done")

    async def GiveMeData(
        self, request, context
    ) -> dataservice.dataservice_pb2.DataResponse:
        index = 0

        while index < len(self.payload):
            yield dataservice.dataservice_pb2.DataResponse(
                data=self.payload[index : index + CHUNK_SIZE]
            )
            index += CHUNK_SIZE


async def serve(port: str, size: int, tls: bool, mtls: bool) -> None:
    server = grpc.aio.server()
    dataservice.dataservice_pb2_grpc.add_DataServiceServicer_to_server(
        Greeter(size), server
    )
    listen_addr = f"[::]:{port}"

    if tls:
        creds = grpc.ssl_server_credentials(
            ((read("../tls/server_key.pem"), read("../tls/server_cert.pem")),),
        )

        server.add_secure_port(listen_addr, creds)
        print("Server started in TLS mode, listening on" + listen_addr)
    if mtls:
        creds = grpc.ssl_server_credentials(
            ((read("../tls/server_key.pem"), read("../tls/server_cert.pem")),),
            root_certificates=read("../tls/client_cert.pem"),
            require_client_auth=True,
        )

        server.add_secure_port(listen_addr, creds)
        print("Server started in mTLS mode, listening on" + listen_addr)
    else:
        server.add_insecure_port(listen_addr)
        print("Server started, listening on" + listen_addr)

    await server.start()
    await server.wait_for_termination()


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("-p", "--port", default=5000)
    parser.add_argument("-s", "--size", type=int, default=int(1 * 1024 * 1024 * 1024))
    parser.add_argument("--tls", action="store_true")
    parser.add_argument("--mtls", action="store_true")

    args = parser.parse_args()

    asyncio.run(serve(args.port, args.size, args.tls, args.mtls))
