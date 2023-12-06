// Adapted from https://github.com/grpc/grpc/tree/master/examples/cpp/helloworld

#include <iostream>
#include <memory>
#include <string>

#include "absl/flags/flag.h"
#include "absl/flags/parse.h"
#include "absl/strings/str_format.h"

#include <grpcpp/ext/proto_server_reflection_plugin.h>
#include <grpcpp/grpcpp.h>
#include <grpcpp/health_check_service_interface.h>

#include "build/dataservice.grpc.pb.h"
#include "build/dataservice.pb.h"

ABSL_FLAG(uint16_t, port, 5000, "Server port for the service");

class DataServiceImpl final : public grpc_bench::DataService::Service {
  grpc::Status
  GiveMeData(grpc::ServerContext *context,
             const grpc_bench::DataRequest *request,
             grpc::ServerWriter<grpc_bench::DataResponse> *writer) override {
    std::cout << "GiveMeData() called" << std::endl;

    grpc_bench::DataResponse response;
    // TODO: Give this real data
    response.set_data("data");

    if (!writer->Write(response)) {
      std::cout << "In GiveMeData, writer->Write failed..." << std::endl;
    } else {
      std::cout << "In GiveMeData, writer->Write succeeded!" << std::endl;
    }

    return grpc::Status::OK;
  }
};

void RunServer(uint16_t port) {
  std::string server_address = absl::StrFormat("0.0.0.0:%d", port);
  DataServiceImpl service;

  grpc::EnableDefaultHealthCheckService(true);
  grpc::reflection::InitProtoReflectionServerBuilderPlugin();

  grpc::ServerBuilder builder;
  builder.AddListeningPort(server_address, grpc::InsecureServerCredentials());
  builder.RegisterService(&service);

  std::unique_ptr<grpc::Server> server(builder.BuildAndStart());

  std::cout << "Server listening on " << server_address << std::endl;

  server->Wait();
}

int main(int argc, char **argv) {
  absl::ParseCommandLine(argc, argv);
  RunServer(absl::GetFlag(FLAGS_port));
  return 0;
}
