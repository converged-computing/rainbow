import argparse
import os

from rainbow.protos import rainbow_pb2
from rainbow.client import RainbowClient

# Config file from a few directories up
here = os.path.abspath(os.path.dirname(__file__))
root = here

# rainbow root directory
for iter in range(4):
    root = os.path.dirname(root)


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
    parser.add_argument(
        "--cluster-nodes",
        help="Nodes to provide for registration",
        default=os.path.join(
            root, "docs", "examples", "scheduler", "cluster-nodes.json"
        ),
    )
    return parser


def main():
    parser = get_parser()
    args, _ = parser.parse_known_args()
    cli = RainbowClient(host=args.host)
    response = cli.register(
        args.cluster, secret=args.secret, cluster_nodes=args.cluster_nodes
    )
    print(response)
    if response.status == rainbow_pb2.RegisterResponse.ResultType.REGISTER_EXISTS:
        print(f"The cluster {args.cluster} already exists.")
    else:
        print(
            f"🤫️ The token you will need to submit jobs to this cluster is {response.token}"
        )
        print(f"🔐️ The secret you will need to accept jobs is {response.secret}")


if __name__ == "__main__":
    main()
