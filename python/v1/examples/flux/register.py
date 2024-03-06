import argparse
import os

from rainbow.protos import rainbow_pb2
from rainbow.client import RainbowClient
import rainbow.config as config

# Config file from a few directories up
here = os.path.abspath(os.path.dirname(__file__))
root = here

# rainbow root directory
for iter in range(4):
    root = os.path.dirname(root)


def get_parser():
    parser = argparse.ArgumentParser(description="🌈️ Rainbow scheduler register")
    parser.add_argument("--cluster", help="cluster name to register", default="keebler")
    parser.add_argument("--host", help="host of rainbow cluster", default="localhost:50051")
    parser.add_argument(
        "--secret",
        help="Rainbow cluster registration secret",
        default="chocolate-cookies",
    )
    parser.add_argument(
        "--config-path",
        help="Path to rainbow configuration file to write or use",
    )
    parser.add_argument(
        "--cluster-nodes",
        help="Nodes to provide for registration",
        default=os.path.join(root, "docs", "examples", "scheduler", "cluster-nodes.json"),
    )
    return parser


def main():
    parser = get_parser()
    args, _ = parser.parse_known_args()
    cli = RainbowClient(host=args.host)

    # Do we want to write or update a config file?
    if not args.config_path or not os.path.exists(args.config_path):
        cfg = config.new_rainbow_config(args.host, args.cluster, args.secret)
    else:
        cfg = config.RainbowConfig(args.config_path)
    response = cli.register(args.cluster, secret=args.secret, cluster_nodes=args.cluster_nodes)

    # In the response:
    # secret is for the cluster to receive jobs
    # token is to submit to it
    cfg._cfg["cluster"]["secret"] = response.secret
    cfg.add_cluster(args.cluster, response.token)

    # Save to path if provided
    if args.config_path:
        print(f"Saving rainbow config to {args.config_path}")
        cfg.save_yaml(args.config_path)

    if response.status == rainbow_pb2.RegisterResponse.ResultType.REGISTER_EXISTS:
        print(f"The cluster {args.cluster} already exists.")
    else:
        print(f"🤫️ The token you will need to submit jobs to this cluster is {response.token}")
        print(f"🔐️ The secret you will need to accept jobs is {response.secret}")


if __name__ == "__main__":
    main()
