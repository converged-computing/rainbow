from __future__ import print_function

import logging
import os

import argparse
from rainbow.client import RainbowClient

from jobspec.plugin import get_transformer_registry
     

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

    # Get the registry
    registry = get_transformer_registry()

    # The cool thing about transformers is that you can have
    # one tiny server that acts an an interface to several cloud (or other)
    # APIs. A transformer doesn't have to be for cluster batch, it could
    # be for an API to an emphemeral resource
    plugin = registry.get_plugin("flux")()

    for job in jobs:
       js = job["jobspec"]
       print(js)

       # Run the plugin with the jobspec (submit the job)
       plugin.run(js)


if __name__ == "__main__":
    logging.basicConfig()
    main()
