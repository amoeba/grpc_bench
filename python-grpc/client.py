import asyncio
import logging
import time
import argparse

import grpc
import dataservice.dataservice_pb2
import dataservice.dataservice_pb2_grpc


def read(path):
    with open(path, "rb") as f:
        return f.read()


async def run_tls(host: str, ntimes: int) -> None:
    total_bytes_received = 0
    total_time_taken = 0

    credential = grpc.ssl_channel_credentials(read("../tls/ca_cert.pem"))

    async with grpc.aio.secure_channel(host, credential) as channel:
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


async def run_mtls(host: str, ntimes: int) -> None:
    print("MTLS")
    total_bytes_received = 0
    total_time_taken = 0

    credential = grpc.ssl_channel_credentials(
        root_certificates=read("../tls/ca_cert.pem"),
        private_key=read("../tls/client_key.pem"),
        certificate_chain=read("../tls/client_cert.pem"),
    )

    async with grpc.aio.secure_channel(host, credential) as channel:
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
    parser.add_argument("--tls", action="store_true")
    parser.add_argument("--mtls", action="store_true")

    args = parser.parse_args()

    logging.basicConfig()

    if args.tls:
        asyncio.run(run_tls(args.address, args.ntimes))
    if args.mtls:
        asyncio.run(run_mtls(args.address, args.ntimes))

    else:
        asyncio.run(run(args.address, args.ntimes))
