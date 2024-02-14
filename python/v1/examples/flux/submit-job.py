from __future__ import print_function

import logging

import argparse
import grpc
import sys
from rainbow.protos import rainbow_pb2
from rainbow.protos import rainbow_pb2_grpc


def get_parser():
    parser = argparse.ArgumentParser(description="🌈️ Rainbow scheduler submit")
    parser.add_argument("--cluster", help="cluster name to register", default="keebler")
    parser.add_argument(
        "--host", help="host of rainbow cluster", default="localhost:50051"
    )
    parser.add_argument(
        "--token", help="Cluster token for permission to submit jobs", required=True
    )
    parser.add_argument(
        "--nodes", help="Nodes for job (defaults to 1)", default=1, type=int
    )
    parser.add_argument("command", help="Command to submit", nargs="+")
    return parser


def main():

    parser = get_parser()
    args, _ = parser.parse_known_args()

    if not args.command:
        sys.exit("A command (positional arguments) is required")
    command = " ".join(args.command)
    print(f"⭐️ Submitting job: {command}")

    # These are submit variables. A more substantial submit script would have argparse, etc.
    submitRequest = rainbow_pb2.SubmitJobRequest(
        token=args.token, nodes=args.nodes, cluster=args.cluster, command=command
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
    with grpc.insecure_channel(args.host) as channel:
        stub = rainbow_pb2_grpc.RainbowSchedulerStub(channel)
        response = stub.SubmitJob(submitRequest)
        print(response)


if __name__ == "__main__":
    logging.basicConfig()
    main()
