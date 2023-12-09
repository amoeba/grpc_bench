// Adapted from https://github.com/grpc/grpc/tree/master/examples/cpp/helloworld

#include "absl/flags/flag.h"
#include "absl/flags/parse.h"
#include "absl/strings/str_format.h"
#include <_types/_uint16_t.h>
#include <filesystem>
#include <fstream>
#include <iostream>

#include <memory>
#include <string>
#include <string_view>

#include <grpcpp/ext/proto_server_reflection_plugin.h>
#include <grpcpp/grpcpp.h>
#include <grpcpp/health_check_service_interface.h>
#include <vector>

#include "build/dataservice.grpc.pb.h"
#include "build/dataservice.pb.h"

#include "common.h"

ABSL_FLAG(uint16_t, port, 5000, "Server port for the service");
ABSL_FLAG(double, size, 1024, "Number of bytes to test with");
ABSL_FLAG(bool, tls, false, "Whether to enable TLS");

#define CHUNKS_SIZE 4 * 1000 * 1000

class DataServiceImpl final : public grpc_bench::DataService::Service {
public:
  DataServiceImpl() {
    auto s = absl::GetFlag(FLAGS_size);

    std::cout << "Initializing service with data size of " << s << std::endl;
    GenerateData(s);
  }

  void GenerateData(double size) {
    std::cout << "Generating data..." << std::endl;

    std::vector<char> payload = std::vector<char>();
    payload.reserve(size);

    // Temporary code: Just initialize the array with the alphabet
    // TODO: Make this random
    for (double i = 0; i < payload.capacity() - 2; i++) {
      payload.push_back('a');
    }
    payload.push_back('\0');

    std::cout << "done generating data..." << std::endl;

    this->data = std::string(payload.cbegin(), payload.cend());

    std::cout << "...Done." << std::endl;
  }

  grpc::Status
  GiveMeData(grpc::ServerContext *context,
             const grpc_bench::DataRequest *request,
             grpc::ServerWriter<grpc_bench::DataResponse> *writer) override {
    std::cout << "GiveMeData()" << std::endl;

    grpc_bench::DataResponse response;

    // Chunk the response into chunks of CHUNK_SIZE
    uint64_t position = 0;
    auto data_view = std::string_view(data);

    while (position < data_view.size()) {
      uint64_t read_to;

      if (position + CHUNKS_SIZE > data_view.size() - 1) {
        read_to = data_view.size();
      } else {
        read_to = position + CHUNKS_SIZE;
      }

      response.set_data(data_view.substr(position, read_to - position));
      position += CHUNKS_SIZE;

      if (!writer->Write(response)) {
        break;
      }
    }

    return grpc::Status::OK;
  }

private:
  std::string data;
};

void RunServerTLS(uint16_t port) {
  std::cout << "RunServerTLS" << std::endl;
  std::string server_address = absl::StrFormat("0.0.0.0:%d", port);
  DataServiceImpl service;

  // TLS stuff
  grpc::SslServerCredentialsOptions ssl_opts;
  ssl_opts.pem_root_certs = "";
  grpc::SslServerCredentialsOptions::PemKeyCertPair keypair = {
      read_file("../../tls/server_key.pem"),
      read_file("../../tls/server_cert.pem")};
  ssl_opts.pem_key_cert_pairs.push_back(keypair);
  auto server_creds = SslServerCredentials(ssl_opts);

  // Regular Server Stuff

  grpc::EnableDefaultHealthCheckService(true);
  grpc::reflection::InitProtoReflectionServerBuilderPlugin();

  grpc::ServerBuilder builder;
  builder.AddListeningPort(server_address, server_creds);
  builder.RegisterService(&service);

  std::unique_ptr<grpc::Server> server(builder.BuildAndStart());

  std::cout << "Server listening on " << server_address << std::endl;

  server->Wait();
}

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

  if (absl::GetFlag(FLAGS_tls)) {
    RunServerTLS(absl::GetFlag(FLAGS_port));
  } else {
    RunServer(absl::GetFlag(FLAGS_port));
  }

  return 0;
}
