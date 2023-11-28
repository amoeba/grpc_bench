import asyncio
import logging

import grpc
import dataservice.dataservice_pb2
import dataservice.dataservice_pb2_grpc


async def run() -> None:
    async with grpc.aio.insecure_channel("localhost:50051") as channel:
        stub = dataservice.dataservice_pb2_grpc.DataServiceStub(channel)
        responses_gen = stub.GiveMeData(dataservice.dataservice_pb2.DataRequest())

        async for response in responses_gen:
            print(
                f"Received {len(response.data)} bytes",
            )


if __name__ == "__main__":
    logging.basicConfig()
    asyncio.run(run())
