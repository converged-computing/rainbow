from __future__ import print_function

import logging

import argparse
import grpc
import random
from rainbow.protos import rainbow_pb2
from rainbow.protos import rainbow_pb2_grpc


def get_parser():
    parser = argparse.ArgumentParser(
        description="🌈️ Rainbow scheduler poll (request jobs) and accept"
    )
    parser.add_argument("--cluster", help="cluster name to register", default="keebler")
    parser.add_argument(
        "--host", help="host of rainbow cluster", default="localhost:50051"
    )
    parser.add_argument(
        "--max-jobs", help="Maximum jobs to request (unset defaults to all)", type=int
    )
    parser.add_argument(
        "--secret", help="Cluster secret to access job queue", required=True
    )
    parser.add_argument(
        "--nodes", help="Nodes for job (defaults to 1)", default=1, type=int
    )
    parser.add_argument("--accept", help="Number of jobs to accept", type=int)
    return parser


def main():

    parser = get_parser()
    args, _ = parser.parse_known_args()

    # These are submit variables. A more substantial submit script would have argparse, etc.
    pollRequest = rainbow_pb2.RequestJobsRequest(
        secret=args.secret, maxJobs=args.max_jobs, cluster=args.cluster
    )

    # NOTE(gRPC Python Team): .close() is possible on a channel and should be
    # used in circumstances in which the with statement does not fit the needs
    # of the code.
    with grpc.insecure_channel(args.host) as channel:
        stub = rainbow_pb2_grpc.RainbowSchedulerStub(channel)
        response = stub.RequestJobs(pollRequest)
        if response.status != 1:
            print("Issue with requesting jobs:")
            print(response)
            return

        # Unwrap ourselves to prettier print
        print("Status: REQUEST_JOBS_SUCCESS")
        print(f"Received {len(response.jobs)} jobs for inspection!")
        for _, job in response.jobs.items():
            # Note this can be json loaded, it's a json string
            print(job)

        # Cut out early if not accepting
        if not args.accept or args.accept < 0:
            return

        # We would normally save metadata to submit to flux, but just faux accept
        # for now (meaning we just need the job ids)
        jobs = list(response.jobs)
        random.shuffle(jobs)

        # We can only accept up to the max that we have
        if args.accept > len(jobs):
            args.accept = len(jobs)

        accepted = jobs[: args.accept]
        print(f"Accepting {args.accept} jobs...")
        print(accepted)
        acceptRequest = rainbow_pb2.AcceptJobsRequest(
            secret=args.secret, jobids=accepted, cluster=args.cluster
        )
        response = stub.AcceptJobs(acceptRequest)


if __name__ == "__main__":
    logging.basicConfig()
    main()
