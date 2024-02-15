from __future__ import print_function

# This demo is modified from the docker-compose to yaml
# to expect a shared token (and thus not need a shared filesystem)!

import logging

import time
import json
import argparse
import grpc
import random
import platform
import requests
import sys
import shlex

from rainbow.protos import rainbow_pb2
from rainbow.protos import rainbow_pb2_grpc

# This needs to be run in a flux instance!
import flux
import flux.job

handle = flux.Flux()

# Get some words!
url = "https://www.mit.edu/~ecprice/wordlist.10000"
response = requests.get(url)
WORDS = response.text.splitlines()


def get_parser():
    parser = argparse.ArgumentParser(
        description="🌈️ Rainbow scheduler poll (request jobs) and accept"
    )
    parser.add_argument("--cluster", help="cluster name to register", default="keebler")
    parser.add_argument(
        "--host",
        help="host of rainbow cluster",
        default="scheduler.rainbow.default.svc.cluster.local:8080",
    )
    parser.add_argument(
        "--max-jobs", help="Maximum jobs to request (unset defaults to all)", type=int
    )
    parser.add_argument(
        "--secret",
        help="Cluster secret to access job queue",
        default="peanutbutta",
    )
    parser.add_argument(
        "--sleep",
        help="Range max to select sleep time between iters (defaults to 10)",
        default=10,
        type=int,
    )
    parser.add_argument(
        "--iters",
        help="Iterations of submit and accept to run (defaults to 10)",
        default=10,
        type=int,
    )
    parser.add_argument(
        "--nodes", help="Nodes for job (defaults to 1)", default=1, type=int
    )
    parser.add_argument(
        "--accept", help="Number of jobs to accept", type=int, default=1
    )
    parser.add_argument(
        "--data", help="Data directory to expect other hostname files", default="/code"
    )
    parser.add_argument("--peer", help="peers to look for", action="append")
    return parser


def wait_hosts_registered(data_dir, peers):
    """
    Wait for all host peers (clusters) to be registered

    Instead of actually waiting, just sleep a little longer here
    anticipating the indexed job pods coming up.
    """
    time.sleep(20)
    print(f"🥳️ All {len(peers)} clusters are registered.")


def write_file(content, filepath):
    """
    Write content to file
    """
    with open(filepath, "w") as fd:
        fd.write(content)


def read_file(filepath):
    """
    Read content from file
    """
    with open(filepath, "r") as fd:
        content = fd.read().strip()
    return content


class RainbowDemo:
    """
    Simple (dumb) class to wrap a channel
    """

    def __init__(self, args):
        self.args = args
        self.tokens = {}
        self.hostname = platform.node()
        self.jobids = set()
        self.start()

    def start(self):
        """
        Start (or open) the channel
        """
        # The cluster name willl be the hostname. This can be anything, really
        self.hostname = platform.node()
        print(f"👋️ Hello, I'm {self.hostname}!")

    def load_peers(self):
        """
        Wait for and load peers

        This is also faux because we are just setting it to be our
        secret since it is shared.
        """
        self.wait()
        for peer in self.args.peer:
            self.tokens[peer] = self.token

    def wait(self):
        """
        Wait for all other known peers to be registered
        """
        # Wait for all hosts to be registered before continuing
        wait_hosts_registered(self.args.data, self.args.peer)

    def stream_output(self, jobid):
        """
        Given a jobid, stream the output
        """
        try:
            for line in flux.job.event_watch(handle, jobid, "guest.output"):
                if "data" in line.context:
                    print("Ran job %s %s" % (jobid, line.context["data"]))
        except Exception:
            pass

    def poll_jobs(self, n_jobs):
        """
        Poll for jobs, accept some number
        """
        # These are submit variables. A more substantial submit script would have argparse, etc.
        pollRequest = rainbow_pb2.RequestJobsRequest(
            secret=self.secret, cluster=self.hostname
        )
        # Open the grpc connection
        channel = grpc.insecure_channel(self.args.host)
        stub = rainbow_pb2_grpc.RainbowSchedulerStub(channel)
        response = stub.RequestJobs(pollRequest)
        channel.close()
        # Success or no jobs
        if response.status not in [0, 1]:
            print("Issue with requesting jobs:")
            sys.exit(str(response))

        # Unwrap ourselves to prettier print
        print("Status: REQUEST_JOBS_SUCCESS")
        print(f"Received {len(response.jobs)} jobs for inspection!")
        if not n_jobs:
            return

        # We would normally save metadata to submit to flux, but just faux accept
        # for now (meaning we just need the job ids)
        jobs = list(response.jobs)
        random.shuffle(jobs)
        joblist = response.jobs

        # We can only accept up to the max that we have
        if n_jobs > len(jobs):
            n_jobs = len(jobs)

        accepted = jobs[:n_jobs]
        print(f"Accepting {n_jobs} jobs...")

        # We often don't have jobs yet
        if len(accepted) == 0:
            return

        print(accepted)
        channel = grpc.insecure_channel(self.args.host)
        stub = rainbow_pb2_grpc.RainbowSchedulerStub(channel)
        acceptRequest = rainbow_pb2.AcceptJobsRequest(
            secret=self.secret, jobids=accepted, cluster=self.hostname
        )
        response = stub.AcceptJobs(acceptRequest)

        # Now actually submit with flux!
        for jobid in accepted:
            job = json.loads(joblist[jobid])
            command = shlex.split(job["command"])
            jobspec = flux.job.JobspecV1.from_command(
                command=command, num_nodes=job["nodes"]
            )
            jobid = flux.job.submit(handle, jobspec)
            print(f"Submit job {command}: {jobid}")

            # We could get the log here, but can do/get later
            self.jobids.add(jobid)
        channel.close()

    def show_info(self):
        """
        Show job info for submit (but not yet seen) jobs
        """
        while self.jobids:
            jobid = self.jobids.pop()
            self.stream_output(jobid)

    def submit_jobs(self, n_submit):
        """
        Submit jobs to random peers (including ourselves)
        """
        channel = grpc.insecure_channel(self.args.host)
        for _ in range(n_submit):
            word = random.choice(WORDS)
            command = f"echo hello from {self.hostname}, a new word is {word}"

            # Randomly select a cluster (ourselves included)
            submit_to = random.choice(self.args.peer)
            submitRequest = rainbow_pb2.SubmitJobRequest(
                token=self.tokens[submit_to],
                nodes=1,
                cluster=submit_to,
                command=command,
            )

            # And submit!
            stub = rainbow_pb2_grpc.RainbowSchedulerStub(channel)
            response = stub.SubmitJob(submitRequest)
            print(response)
        channel.close()

    def register(self):
        """
        Register the cluster to the rainbow scheduler
        """
        registerRequest = rainbow_pb2.RegisterRequest(
            name=self.hostname, secret=self.args.secret
        )
        channel = grpc.insecure_channel(self.args.host)
        stub = rainbow_pb2_grpc.RainbowSchedulerStub(channel)
        print(f"📜️ Registering {self.hostname}...")
        response = stub.Register(registerRequest)
        print(response)
        channel.close()

        # Success
        self.secret = response.secret
        self.token = response.token


def main():

    parser = get_parser()
    args, _ = parser.parse_known_args()

    # Give a tiny bit of time for the server to boot
    time.sleep(5)

    demo = RainbowDemo(args)

    # Step 1 is to register ourselves! We will write our token (for the other clusters to access)
    # in the shared data directory
    demo.register()

    # Wait for other peers to be registered and read in the tokens
    demo.load_peers()

    # Now let the insanity begin!
    # We are going to, in a long loop, submit and accept jobs
    for _ in range(args.iters):

        # Submit some N jobs
        n_submit = random.choice(range(1, 6))
        demo.submit_jobs(n_submit)

        # Run some N jobs!
        n_accept = random.choice(range(1, 6))
        demo.poll_jobs(n_accept)
        time.sleep(args.sleep)
        demo.show_info()

    print(f"💤️ Cluster {demo.hostname} is finished! Shutting down.")


if __name__ == "__main__":
    logging.basicConfig()
    main()
