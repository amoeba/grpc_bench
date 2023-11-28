import asyncio
import logging
import io

import numpy as np
import grpc

import dataservice.dataservice_pb2
import dataservice.dataservice_pb2_grpc


def gen_random(length=int) -> np.array:
    return np.random.randint(0, length, length, dtype=np.dtype(np.int64))


payload = gen_random(100)


class Greeter(dataservice.dataservice_pb2_grpc.DataServiceServicer):
    async def GiveMeData(
        self, request, context
    ) -> dataservice.dataservice_pb2.DataResponse:
        logging.info("Serving GiveMeData request %s", request)
        yield dataservice.dataservice_pb2.DataResponse(data=payload.tobytes())


async def serve() -> None:
    server = grpc.aio.server()
    dataservice.dataservice_pb2_grpc.add_DataServiceServicer_to_server(
        Greeter(), server
    )
    listen_addr = "[::]:50051"
    server.add_insecure_port(listen_addr)
    print("Server started, listening on %s" + listen_addr)

    await server.start()
    await server.wait_for_termination()


if __name__ == "__main__":
    logging.basicConfig()
    asyncio.run(serve())
