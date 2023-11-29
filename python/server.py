import asyncio
import logging
import argparse

import numpy as np
import grpc

import dataservice.dataservice_pb2
import dataservice.dataservice_pb2_grpc

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


async def serve(port: str, size: int) -> None:
    server = grpc.aio.server()
    dataservice.dataservice_pb2_grpc.add_DataServiceServicer_to_server(
        Greeter(size), server
    )
    listen_addr = f"[::]:{port}"
    server.add_insecure_port(listen_addr)
    print("Server started, listening on %s" + listen_addr)

    await server.start()
    await server.wait_for_termination()


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("-p", "--port", default=5000)
    parser.add_argument("-s", "--size", type=int, default=int(1 * 1024 * 1024 * 1024))

    args = parser.parse_args()

    logging.basicConfig()
    asyncio.run(serve(args.port, args.size))
