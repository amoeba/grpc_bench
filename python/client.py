import logging

import grpc
import dataservice.dataservice_pb2
import dataservice.dataservice_pb2_grpc


def run():
    with grpc.insecure_channel("localhost:50051") as channel:
        stub = dataservice.dataservice_pb2_grpc.DataServiceStub(channel)
        response = stub.GiveMeData(dataservice.dataservice_pb2.DataRequest())

    print(response)


if __name__ == "__main__":
    logging.basicConfig()
    run()
