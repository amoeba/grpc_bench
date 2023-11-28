from concurrent import futures
import logging

import grpc
import dataservice.dataservice_pb2
import dataservice.dataservice_pb2_grpc

# todo: figure out the best way to send raw random bytes, this is not it
payload = bytes(b"123")


class Greeter(dataservice.dataservice_pb2_grpc.DataServiceServicer):
    def GiveMeData(self, request, context):
        print("GiveMeData is running")
        return dataservice.dataservice_pb2.HelloReply(data=payload)


def serve():
    port = "50051"
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    dataservice.dataservice_pb2_grpc.add_DataServiceServicer_to_server(
        Greeter(), server
    )
    server.add_insecure_port("[::]:" + port)
    server.start()
    print("Server started, listening on " + port)
    server.wait_for_termination()


if __name__ == "__main__":
    logging.basicConfig()
    serve()
