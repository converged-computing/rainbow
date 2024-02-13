from __future__ import print_function

import logging

import sys
import grpc
from rainbow.protos import api_pb2
from rainbow.protos import api_pb2_grpc


def main(cluster):

    # These are the variables for our cluster - name for now
    registerRequest = api_pb2.RegisterRequest(name=cluster, secret="chocolate-cookies")

    # NOTE(gRPC Python Team): .close() is possible on a channel and should be
    # used in circumstances in which the with statement does not fit the needs
    # of the code.
    with grpc.insecure_channel("localhost:50051") as channel:
        stub = api_pb2_grpc.RainbowSchedulerStub(channel)
        response = stub.Register(registerRequest)
        print(response)
        if response.status == api_pb2.RegisterResponse.ResultType.REGISTER_EXISTS:
            print(f"The cluster {cluster} alreadey exists.")
        else:
            print(
                f"The token you will need to submit jobs to this cluster is {response.token}",
            )


if __name__ == "__main__":
    logging.basicConfig()
    cluster = "keebler"
    if len(sys.argv) > 1:
        cluster = sys.argv[1]
    main(cluster)
