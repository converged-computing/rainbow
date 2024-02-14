from __future__ import print_function

import logging

import argparse
import grpc
from rainbow.protos import rainbow_pb2
from rainbow.protos import rainbow_pb2_grpc


def get_parser():
    parser = argparse.ArgumentParser(description="🌈️ Rainbow scheduler register")
    parser.add_argument("--cluster", help="cluster name to register", default="keebler")
    parser.add_argument(
        "--host", help="host of rainbow cluster", default="localhost:50051"
    )
    parser.add_argument(
        "--secret",
        help="Rainbow cluster registration secret",
        default="chocolate-cookies",
    )
    return parser


def main():

    parser = get_parser()
    args, _ = parser.parse_known_args()

    # These are the variables for our cluster - name for now
    registerRequest = rainbow_pb2.RegisterRequest(name=args.cluster, secret=args.secret)

    # NOTE(gRPC Python Team): .close() is possible on a channel and should be
    # used in circumstances in which the with statement does not fit the needs
    # of the code.
    with grpc.insecure_channel(args.host) as channel:
        stub = rainbow_pb2_grpc.RainbowSchedulerStub(channel)
        response = stub.Register(registerRequest)
        print(response)
        if response.status == rainbow_pb2.RegisterResponse.ResultType.REGISTER_EXISTS:
            print(f"The cluster {args.cluster} already exists.")
        else:
            print(
                f"🤫️ The token you will need to submit jobs to this cluster is {response.token}",
            )
            print(
                f"🔐️ The secret you will need to accept jobs is {response.secret}",
            )


if __name__ == "__main__":
    logging.basicConfig()
    main()
