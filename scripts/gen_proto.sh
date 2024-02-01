#!/bin/sh
#
# Generate GRPC code for all benchmarks except C++.
#
# Dependency resolution isn't yet automated so you'll need the dependencies for
# each benchmark available to run this.
#
# NOTE: This may not be working correctly at the moment

cd "src/go/grpc" || exit
echo "$PWD"
protoc -I../protos --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    ../protos/dataservice/dataservice.proto
cd "../python" || exit

echo "$PWD"
python -m grpc_tools.protoc -I../protos --python_out=. --pyi_out=. --grpc_python_out=. ../protos/dataservice/dataservice.proto
cd ..
