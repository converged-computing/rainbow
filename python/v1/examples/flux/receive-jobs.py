from __future__ import print_function

import logging

import argparse
from rainbow.client import RainbowClient


def get_parser():
    parser = argparse.ArgumentParser(
        description="🌈️ Rainbow scheduler receive jobs"
    )
    parser.add_argument(
        "--max-jobs", help="Maximum jobs to request", type=int, default=1
    )
    parser.add_argument("--config-path", help="config path with cluster metadata")
    return parser

def main():

    parser = get_parser()
    args, _ = parser.parse_known_args()


    # The config path (with clusters) will be required for submit
    cli = RainbowClient(config_file=args.config_path)
    jobs = cli.receive_jobs(args.max_jobs)
    print(jobs)

if __name__ == "__main__":
    logging.basicConfig()
    main()
