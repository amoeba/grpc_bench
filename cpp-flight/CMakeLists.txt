cmake_minimum_required(VERSION 3.10)

project(throughput_test VERSION 1.0.0 LANGUAGES CXX)

find_package(Arrow REQUIRED)
find_package(ArrowFlight REQUIRED)
find_package(Parquet REQUIRED)

add_executable(client client.cc)
set_property(TARGET client PROPERTY CXX_STANDARD 17)
target_link_libraries(client PRIVATE arrow_shared arrow_flight_shared parquet_shared)

add_executable(server server.cc)
set_property(TARGET server PROPERTY CXX_STANDARD 17)
target_link_libraries(server PRIVATE arrow_shared arrow_flight_shared parquet_shared)
