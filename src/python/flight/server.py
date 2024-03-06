import argparse
import ast
import threading
import time

import pyarrow as pa
import pyarrow.flight
import numpy as np

MAX_CHUNK_SIZE = 64_000


# Basic Flight server, copied from https://github.com/apache/arrow
class FlightServer(pyarrow.flight.FlightServerBase):
    def __init__(
        self,
        host="localhost",
        location=None,
        tls_certificates=None,
        verify_client=False,
        root_certificates=None,
        auth_handler=None,
    ):
        super(FlightServer, self).__init__(
            location, auth_handler, tls_certificates, verify_client, root_certificates
        )
        self.flights = {}
        self.host = host
        self.tls_certificates = tls_certificates
        self.location = location

    @classmethod
    def descriptor_to_key(self, descriptor):
        return (
            descriptor.descriptor_type.value,
            descriptor.command,
            tuple(descriptor.path or tuple()),
        )

    def _make_flight_info(self, key, descriptor, table):
        location = pyarrow.flight.Location.for_grpc_unix(self.location)
        endpoints = [
            pyarrow.flight.FlightEndpoint(repr(key), [location]),
        ]

        mock_sink = pyarrow.MockOutputStream()
        stream_writer = pyarrow.RecordBatchStreamWriter(mock_sink, table.schema)
        stream_writer.write_table(table)
        stream_writer.close()
        data_size = mock_sink.size()

        return pyarrow.flight.FlightInfo(
            table.schema, descriptor, endpoints, table.num_rows, data_size
        )

    def list_flights(self, context, criteria):
        for key, table in self.flights.items():
            if key[1] is not None:
                descriptor = pyarrow.flight.FlightDescriptor.for_command(key[1])
            else:
                descriptor = pyarrow.flight.FlightDescriptor.for_path(*key[2])

            yield self._make_flight_info(key, descriptor, table)

    def get_flight_info(self, context, descriptor):
        key = FlightServer.descriptor_to_key(descriptor)

        if key in self.flights:
            table = self.flights[key]
            return self._make_flight_info(key, descriptor, table)

        raise KeyError("Flight not found.")

    def do_put(self, context, descriptor, reader, writer):
        key = FlightServer.descriptor_to_key(descriptor)
        self.flights[key] = reader.read_all()

    def do_get(self, context, ticket):
        key = ast.literal_eval(ticket.ticket.decode())

        if key not in self.flights:
            return None

        return pyarrow.flight.RecordBatchStream(
            self.flights[key].to_reader(MAX_CHUNK_SIZE)
        )


# From https://github.com/wjones127/arrow-ipc-bench
def calculate_ipc_size(table: pa.Table) -> int:
    sink = pa.MockOutputStream()

    with pa.ipc.new_stream(sink, table.schema) as writer:
        writer.write_table(table)

    return sink.size()


def create_table(nrows: int, ncols: int):
    tbl = {}

    for i in range(ncols):
        tbl[f"col{i}"] = np.random.random(nrows)

    return pa.table(tbl)


def run_server(location, tls_certificates=[]):
    if len(tls_certificates) > 0:
        server = FlightServer("localhost", location, tls_certificates=tls_certificates)
    else:
        server = FlightServer("localhost", location)

    server.serve()


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("-p", "--port", default=5000)
    parser.add_argument("-n", "--nrows", type=int, default=68_000_000)
    parser.add_argument("-c", "--ncols", type=int, default=2)
    parser.add_argument("--tls", action="store_true")

    args = parser.parse_args()
    client_kwargs = {}
    tls_certificates = []

    if args.tls:
        # Set up server side of TLS
        with open("../../../certs/server_cert.pem", "rb") as cert_file:
            tls_cert_chain = cert_file.read()
        with open("../../../certs/server_key.pem", "rb") as key_file:
            tls_private_key = key_file.read()

        tls_certificates.append((tls_cert_chain, tls_private_key))

        scheme = "grpc+tls"

        # Set up client side of TLS
        with open("../../../certs/ca_cert.pem", "rb") as cert:
            client_kwargs["tls_root_certs"] = cert.read()
    else:
        scheme = "grpc+tcp"

    location = f"{scheme}://localhost:{args.port}"

    # Start server in another thread
    serve_thread = threading.Thread(
        target=run_server,
        kwargs={"location": location, "tls_certificates": tls_certificates},
    )
    serve_thread.start()
    print("Sleeping for 1s to let the sever become available...")
    time.sleep(1)

    # Generate a Table to insert
    print(f"Creating {args.nrows}x{args.ncols} table...")
    tbl = create_table(args.nrows, args.ncols)
    buffer_size = calculate_ipc_size(tbl)
    buffer_size_gb = buffer_size / 1024 / 1024 / 1024
    print(
        f"Created {args.nrows}x{args.ncols} which came out to be {buffer_size} bytes ({buffer_size_gb} GiB)"
    )

    descriptor = pa.flight.FlightDescriptor.for_path("table")
    client = pa.flight.connect(location, **client_kwargs)
    print("Client calling DoPut...")
    writer, _ = client.do_put(descriptor, tbl.schema)
    writer.write_table(tbl, max_chunksize=MAX_CHUNK_SIZE)
    writer.close()
    print("Table inserted, ready to go.")
    print("Serving on", location)


if __name__ == "__main__":
    main()
