from __future__ import print_function

import logging

import json
import argparse
from rainbow.client import RainbowClient
import rainbow.jobspec.converter as converter
import rainbow.jobspec as js

def get_parser():
    parser = argparse.ArgumentParser(description="🌈️ Rainbow scheduler submit")
    parser.add_argument("--config-path", help="config path with cluster metadata")
    parser.add_argument("jobspec", help="Jobspec path to submit")
    return parser


def main():

    parser = get_parser()
    args = parser.parse_args()

    # The config path (with clusters) will be required for submit
    cli = RainbowClient(config_file=args.config_path)

    # Generate the jobspec here so we can json dump it for the user
    # Note that this can be done with cli.submit_job(command, nodes, tasks)
    jobspec = js.Jobspec(args.jobspec)
    print(json.dumps(jobspec.jobspec, indent=4))
    response = cli.submit_jobspec(jobspec)
    print(response)


if __name__ == "__main__":
    logging.basicConfig()
    main()
