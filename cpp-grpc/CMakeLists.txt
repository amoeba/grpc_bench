cmake_minimum_required(VERSION 3.15)
project(DataService)

find_package(gRPC CONFIG REQUIRED)
message(STATUS "Using gRPC ${gRPC_VERSION}")
find_package(absl CONFIG REQUIRED)
message(STATUS "Using abseil ${absl_VERSION}")
find_program(PROTOC_BIN protoc)
message(STATUS "Found protoc bin at: ${PROTOC_BIN}")
find_program(_GRPC_CPP_PLUGIN_EXECUTABLE grpc_cpp_plugin)
message(STATUS "Found protoc-gen-grpc bin at: ${_GRPC_CPP_PLUGIN_EXECUTABLE}")

get_filename_component(PROTO_PATH "../protos/dataservice/dataservice.proto" ABSOLUTE)
get_filename_component(PROTO_DIR "${PROTO_PATH}" PATH)

message(STATUS ${PROTO_PATH})
message(STATUS ${PROTO_DIR})

# Generate protobuf sources from our service definition
set(PROTO_SRCS "${CMAKE_CURRENT_BINARY_DIR}/dataservice.pb.cc")
set(PROTO_HDRS "${CMAKE_CURRENT_BINARY_DIR}/dataservice.pb.h")
set(GRPC_SRCS "${CMAKE_CURRENT_BINARY_DIR}/dataservice.grpc.pb.cc")
set(GRPC_HDRS "${CMAKE_CURRENT_BINARY_DIR}/dataservice.grpc.pb.h")
add_custom_command(
      OUTPUT "${PROTO_SRCS}" "${PROTO_HDRS}" "${GRPC_SRCS}" "${GRPC_HDRS}"
      COMMAND ${PROTOC_BIN}
      ARGS --grpc_out "${CMAKE_CURRENT_BINARY_DIR}"
        --cpp_out "${CMAKE_CURRENT_BINARY_DIR}"
        -I "${PROTO_DIR}"
        --plugin=protoc-gen-grpc="${_GRPC_CPP_PLUGIN_EXECUTABLE}"
        "${PROTO_PATH}"
      DEPENDS "${PROTO_PATH}")

include_directories("${CMAKE_CURRENT_BINARY_DIR}")

add_library(grpc_proto
  ${GRPC_SRCS}
  ${GRPC_HDRS}
  ${PROTO_SRCS}
  ${PROTO_HDRS})
target_link_libraries(grpc_proto
  gRPC::grpc++_reflection
  gRPC::grpc++
  protobuf::libprotobuf
)

add_executable(server server.cc)
target_link_libraries(server
  grpc_proto
  absl::flags
  absl::flags_parse
  gRPC::grpc++_reflection
  gRPC::grpc++
  protobuf::libprotobuf
)

add_executable(client client.cc)
target_link_libraries(client
  grpc_proto
  absl::flags
  absl::flags_parse
  gRPC::grpc++_reflection
  gRPC::grpc++
  protobuf::libprotobuf
)
