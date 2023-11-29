import asyncio
import logging
import time
import argparse

import grpc
import dataservice.dataservice_pb2
import dataservice.dataservice_pb2_grpc


async def run(host: str, ntimes: int) -> None:
    total_bytes_received = 0
    total_time_taken = 0

    async with grpc.aio.insecure_channel(host) as channel:
        stub = dataservice.dataservice_pb2_grpc.DataServiceStub(channel)
        for i in range(ntimes):
            responses_gen = stub.GiveMeData(dataservice.dataservice_pb2.DataRequest())

            bytes_received = 0

            start_time = time.perf_counter()

            async for response in responses_gen:
                bytes_received += len(response.data)
            end_time = time.perf_counter()
            elapsed = end_time - start_time

            print(f"streamed total bytes {bytes_received} bytes in {elapsed} seconds")
            total_bytes_received += bytes_received
            total_time_taken += elapsed

    total_in_GB = total_bytes_received / 1024 / 1024 / 1024
    throughput_GB_s = total_in_GB / total_time_taken
    print(f"average throughput {throughput_GB_s} GB/s")


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("-a", "--address", default="localhost:5000")
    parser.add_argument("-n", "--ntimes", type=int, default=10)
    args = parser.parse_args()

    logging.basicConfig()
    asyncio.run(run(args.address, args.ntimes))
