import logging

import grpc
import dataservice.dataservice_pb2
import dataservice.dataservice_pb2_grpc


def run():
    with grpc.insecure_channel("localhost:50051") as channel:
        stub = dataservice.dataservice_pb2_grpc.DataService(channel)
        response = stub.GiveMeData()

    print("Greeter client received: " + response.message)


if __name__ == "__main__":
    logging.basicConfig()
    run()
