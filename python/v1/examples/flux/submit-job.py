from __future__ import print_function

import logging

import json
import argparse
from rainbow.client import RainbowClient
from jobspec.core import Jobspec, converter

# Your host should be in the rainbow-config.yaml
def get_parser():
    parser = argparse.ArgumentParser(description="🌈️ Rainbow scheduler submit")
    parser.add_argument("--config-path", help="config path with cluster metadata")
    parser.add_argument(
        "--nodes", help="Nodes for job (defaults to 1)", default=1, type=int
    )
    parser.add_argument("command", help="Command to submit", nargs="+")
    return parser


# Note that if you are running in a flux instance, you can use flux to provide
# this parsing of the jobspec. Here we just manully generate it.

def main():

    parser = get_parser()
    args = parser.parse_args()

    # The config path (with clusters) will be required for submit
    cli = RainbowClient(config_file=args.config_path)

    # Generate the jobspec here so we can json dump it for the user
    # Note that this can be done with cli.submit_job(command, nodes, tasks)
    raw = converter.new_simple_jobspec(nodes=args.nodes, command=args.command)
    print(json.dumps(raw, indent=4))

    # This loads and validates
    jobspec = Jobspec(raw)
    response = cli.submit_jobspec(jobspec)
    print(response)


if __name__ == "__main__":
    logging.basicConfig()
    main()
