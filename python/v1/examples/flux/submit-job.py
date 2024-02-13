from __future__ import print_function

import logging

import sys
import grpc
from rainbow.protos import api_pb2
from rainbow.protos import api_pb2_grpc


def main(token):

    # These are submit variables. A more substantial submit script would have argparse, etc.
    submitRequest = api_pb2.SubmitJobRequest(
        token=token, nodes=1, cluster="keebler", command="hostname"
    )

    # These are the variables currently allowed:
    #  string name = 1;
    #  string cluster = 2;
    #  string token = 3;
    #  int32 nodes = 4;
    #  int32 tasks = 5;
    #  string command = 6;
    #  google.protobuf.Timestamp sent = 7;

    # NOTE(gRPC Python Team): .close() is possible on a channel and should be
    # used in circumstances in which the with statement does not fit the needs
    # of the code.
    with grpc.insecure_channel("localhost:50051") as channel:
        stub = api_pb2_grpc.RainbowSchedulerStub(channel)
        response = stub.SubmitJob(submitRequest)
        print(response)


if __name__ == "__main__":
    logging.basicConfig()
    if len(sys.argv) < 2:
        sys.exit(
            "Please include the secret token you received when registering your cluster"
        )
    main(sys.argv[1])
