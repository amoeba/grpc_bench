// Adapted from https://github.com/grpc/grpc/tree/master/examples/cpp/helloworld

#include "absl/flags/flag.h"
#include "absl/flags/parse.h"
#include <chrono>
#include <fstream>
#include <iostream>
#include <memory>
#include <string>

#include "build/dataservice.grpc.pb.h"
#include "build/dataservice.pb.h"
#include <grpcpp/grpcpp.h>
#include <grpcpp/security/credentials.h>

#include "common.h"

ABSL_FLAG(std::string, target, "localhost:5000", "Server address");
ABSL_FLAG(int, ntimes, 10, "Number of times to run the test");
ABSL_FLAG(bool, tls, false, "Whether to enable TLS");

class DataServiceClient {
public:
  DataServiceClient(std::shared_ptr<grpc::Channel> channel)
      : stub_(grpc_bench::DataService::NewStub(channel)) {}
  uint64_t GiveMeData() {
    uint64_t total_bytes_read = 0;

    grpc_bench::DataRequest request;
    grpc_bench::DataResponse reply;
    grpc::ClientContext context;

    std::unique_ptr<grpc::ClientReader<grpc_bench::DataResponse>> reader(
        stub_->GiveMeData(&context, request));

    while (reader->Read(&reply)) {
      total_bytes_read += reply.data().size();
    }

    grpc::Status status = reader->Finish();

    if (!status.ok()) {
      std::cout << "Status was NOT okay:" << status.error_message() << " "
                << std::endl;

      return -1;
    }

    return total_bytes_read;
  }

private:
  std::unique_ptr<grpc_bench::DataService::Stub> stub_;
  std::shared_ptr<grpc::ChannelCredentials> creds;
};

double RunMain(DataServiceClient &client) {
  auto start = std::chrono::high_resolution_clock::now();
  auto bytes_read = client.GiveMeData();

  if (bytes_read < 0) {
    std::cout << "Encountered an unexpected error when calling RPC. Exiting."
              << std::endl;
    return (double)-1;
  }

  auto end = std::chrono::high_resolution_clock::now();
  std::chrono::duration<double> duration = end - start;
  double throughput =
      (((double)bytes_read / 1024 / 1024 / 1024) / duration.count());
  std::cout << throughput << " GB/s" << std::endl;

  return throughput;
}

DataServiceClient getClientTLS() {
  std::string target_str = absl::GetFlag(FLAGS_target);

  auto ssl_opts = grpc::SslCredentialsOptions();
  ssl_opts.pem_root_certs = read_file("../../../../certs/ca_cert.pem");

  auto channel_creds = grpc::SslCredentials(ssl_opts);

  auto channel = grpc::CreateChannel(target_str, channel_creds);
  DataServiceClient client(channel);

  return client;
}

DataServiceClient getClient() {
  std::string target_str = absl::GetFlag(FLAGS_target);

  DataServiceClient client(
      grpc::CreateChannel(target_str, grpc::InsecureChannelCredentials()));

  return client;
}

void Benchmark(DataServiceClient client, int ntimes) {
  double values[ntimes];

  for (int i = 0; i < ntimes; i++) {
    values[i] = RunMain(client);
  }

  std::cout << "Average throughput: " << mean(values, ntimes) << " GB/s"
            << std::endl;
}

int main(int argc, char **argv) {
  absl::ParseCommandLine(argc, argv);
  int ntimes = absl::GetFlag(FLAGS_ntimes);

  if (absl::GetFlag(FLAGS_tls)) {
    Benchmark(getClientTLS(), ntimes);
  } else {
    Benchmark(getClient(), ntimes);
  }

  return 0;
}
