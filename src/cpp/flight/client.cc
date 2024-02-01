#include <cstddef>
#include <iostream>
#include <memory>
#include <thread>

#include <arrow/api.h>
#include <arrow/filesystem/api.h>
#include <arrow/flight/api.h>
#include <arrow/result.h>
#include <arrow/status.h>

using arrow::Status;

Status RunMain(int argc, char **argv) {
  arrow::flight::Location location;
  ARROW_ASSIGN_OR_RAISE(location,
                        arrow::flight::Location::ForGrpcTcp("localhost", 3000));

  std::unique_ptr<arrow::flight::FlightClient> client;
  ARROW_ASSIGN_OR_RAISE(client, arrow::flight::FlightClient::Connect(location));
  std::cout << "Connected to " << location.ToString() << std::endl;

  // We don't need to craft a real ticket because our server always sends the
  // same payload
  //   arrow::flight::Ticket fake_ticket;

  //   std::unique_ptr<arrow::flight::FlightStreamReader> stream;
  //   ARROW_ASSIGN_OR_RAISE(stream, client->DoGet(fake_ticket));
  //   std::shared_ptr<arrow::Table> table;
  //   ARROW_ASSIGN_OR_RAISE(table, stream->ToTable());

  ARROW_ASSIGN_OR_RAISE(auto listing, client->ListFlights())

  while (true) {
    std::unique_ptr<arrow::flight::FlightInfo> flight_info;
    ARROW_ASSIGN_OR_RAISE(flight_info, listing->Next());
    if (!flight_info)
      break;
    std::cout << flight_info->descriptor().ToString() << std::endl;
  }
  return Status::OK();
}

int main(int argc, char **argv) {
  Status status = RunMain(argc, argv);

  if (!status.ok()) {
    std::cerr << status << std::endl;
  }

  return 0;
}
