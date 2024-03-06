import json

import yaml


def read_json(filename):
    """
    Read json from file
    """
    return json.loads(read_file(filename))


def read_file(filename):
    """
    Read in a file content
    """
    with open(filename, "r") as fd:
        content = fd.read()
    return content


def read_yaml(filename):
    """
    Read yaml from file
    """
    with open(filename, "r") as fd:
        content = yaml.safe_load(fd)
    return content


def write_yaml(obj, filename):
    """
    Read yaml to file
    """
    with open(filename, "w") as fd:
        yaml.dump(obj, fd)
