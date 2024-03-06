import argparse
from io import TextIOWrapper
import pyarrow as pa
import pyarrow.compute as pc
import pyarrow.flight
from multiprocessing import shared_memory
from contextlib import contextmanager
import time

import numpy


def retrieve_flight(client, name, compute=False) -> int:
    descriptor = pa.flight.FlightDescriptor.for_path(name)
    flight_info = client.get_flight_info(descriptor)

    reader = client.do_get(flight_info.endpoints[0].ticket)

    for chunk in reader:
        for col in chunk.data.columns:
            if compute:
                pc.sum(col)

    return flight_info.total_bytes


def run_one(client, compute=False) -> float:
    start_time = time.perf_counter()
    bytes_received = retrieve_flight(client, "table", compute)
    end_time = time.perf_counter()
    elapsed = end_time - start_time

    total_gb = bytes_received / 1024 / 1024 / 1024
    throughput_gbs = total_gb / elapsed

    return throughput_gbs


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("-a", "--address", default="localhost:5000")
    parser.add_argument("-n", "--niters", default=10)
    parser.add_argument("--tls", action="store_true")

    args = parser.parse_args()

    kwargs = {}

    with open("../../../certs/ca_cert.pem", "rb") as cert:
        kwargs["tls_root_certs"] = cert.read()

    if args.tls:
        location = f"grpc+tls://{args.address}"
    else:
        location = f"grpc+tcp://{args.address}"

    client = pa.flight.connect(location, **kwargs)

    runs = []

    for i in range(args.niters):
        res = run_one(client, False)
        print(res)
        runs.append(res)
    print(f"average of {args.niters} runs: {numpy.mean(runs)} GiB/s")
