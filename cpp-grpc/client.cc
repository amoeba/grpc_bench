// Adapted from https://github.com/grpc/grpc/tree/master/examples/cpp/helloworld

#include <iostream>
#include <memory>
#include <string>

#include "absl/flags/flag.h"
#include "absl/flags/parse.h"

#include <grpcpp/grpcpp.h>

#include "build/dataservice.grpc.pb.h"
#include "build/dataservice.pb.h"

ABSL_FLAG(std::string, target, "localhost:5000", "Server address");

class DataServiceClient {
public:
  DataServiceClient(std::shared_ptr<grpc::Channel> channel)
      : stub_(grpc_bench::DataService::NewStub(channel)) {}

  void GiveMeData() {
    grpc_bench::DataRequest request;
    grpc_bench::DataResponse reply;
    grpc::ClientContext context;

    std::unique_ptr<grpc::ClientReader<grpc_bench::DataResponse>> reader(
        stub_->GiveMeData(&context, request));

    while (reader->Read(&reply)) {
      // TODO: We can get the read size and the data out here?
      std::cout << "Read..." << std::endl;
      std::cout << "reply.data() is : `" << reply.data() << "`" << std::endl;
    }

    grpc::Status status = reader->Finish();

    if (!status.ok()) {
      std::cout << "Status was NOT okay:" << status.error_message()
                << std::endl;

    } else {
      std::cout << "Status WAS okay. Done." << std::endl;
    }
  }

private:
  std::unique_ptr<grpc_bench::DataService::Stub> stub_;
};

int main(int argc, char **argv) {
  absl::ParseCommandLine(argc, argv);
  std::string target_str = absl::GetFlag(FLAGS_target);
  DataServiceClient client(
      grpc::CreateChannel(target_str, grpc::InsecureChannelCredentials()));

  client.GiveMeData();

  std::cout << "Done received: " << std::endl;

  return 0;
}
