# Note that jsonschema used to be a part of flux core but was dropped
# https://github.com/flux-framework/flux-core/pull/5678/files

import json
import os

import jsonschema
import yaml

import rainbow.schema as schema
import rainbow.utils as utils


class JobspecV1:
    def __init__(self, filename, validate=True, schema=schema.jobspec_v1):
        """
        Load in and validate a Flux Jobspec
        """
        self.schema = schema
        self.filename = filename
        self.jobspec = None
        self.load(filename)
        if validate:
            self.validate()

    def to_str(self):
        """
        Convert to string, which is needed for the submit.
        """
        return json.dumps(self.jobspec)

    def to_yaml(self):
        """
        Dump the jobspec to yaml string
        """
        return yaml.dump(self.jobspec)

    @property
    def name(self):
        try:
            return self.jobspec.get("tasks", {})[0]["command"][0]
        except Exception:
            return "app"

    def load(self, filename):
        """
        Load the jobspec
        """
        # Case 1: given a raw filename
        if isinstance(filename, str) and os.path.exists(filename):
            self.filename = os.path.abspath(filename)

            try:
                self.jobspec = utils.read_json(self.filename)
            except:
                self.jobspec = utils.read_yaml(self.filename)

        # Case 2: jobspec as dict (that we just want to validate)
        elif isinstance(filename, dict):
            self.jobspec = filename
        # Case 3: jobspec as string
        else:
            self.jobspec = json.loads(filename)

        # Case 4: wtf are you giving me? :X
        if not self.jobspec:
            raise ValueError("Unrecognized input format for jobspec.")

    def validate(self):
        """
        Validate the jsonschema
        """
        jsonschema.validate(self.jobspec, self.schema)
        # Require at least one of command, batch, or script
        for task in self.jobspec.get("tasks", []):
            if "command" not in task and "batch" not in task and "script" not in task:
                raise ValueError("Jobspec is not valid, task is missing a command|script|batch")
