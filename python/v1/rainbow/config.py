import os

import jsonschema

import rainbow.schema as schemas
import rainbow.utils as utils


def new_rainbow_config(host, cluster, secret, scheduler_name="rainbow"):
    template = {
        "scheduler": {
            "name": scheduler_name,
            "secret": secret,
            "algorithm": {"name": "random"},
        },
        "cluster": {
            "name": cluster,
        },
        "graphdatabase": {
            "name": "memory",
            "options": {"host": host},
        },
    }
    cfg = RainbowConfig()
    cfg._cfg = template
    cfg.validate()
    return cfg


class RainbowConfig:
    """
    A RainbowClient is able to interact with a Rainbow cluster from Python.
    """

    def __init__(self, config_file=None, validate=True):
        """
        Create a new rainbow client to interact with a rainbow cluster.
        """
        self.config_file = config_file
        self._cfg = None
        if self.config_file and os.path.exists(self.config_file):
            self.load()
            if validate:
                self.validate()

    @property
    def cfg(self):
        """
        Helper function to get/load the config if not done yet.
        """
        if self._cfg is None:
            self.load()
        return self._cfg

    def load(self, config_file=None):
        """
        Load a rainbow config
        """
        config_file = config_file or self.config_file

        # The config is required
        if not config_file or not os.path.exists(config_file):
            raise ValueError("This functionality requires a <instance>.config_file")
        self._cfg = utils.read_yaml(config_file)
        self.config_file = config_file
        return self._cfg

    def validate(self):
        """
        Validate the config against its schema
        """
        if not self._cfg:
            return
        jsonschema.validate(self._cfg, schema=schemas.rainbow_config_v1)

    def save_yaml(self, path):
        """
        Write yaml to file
        """
        utils.write_yaml(self._cfg, path)

    def remove_cluster(self, name):
        """
        Remove a cluster from the listing.
        """
        updated = []
        for cluster in self._cfg["clusters"]:
            if cluster["name"] == name:
                continue
            updated.append(cluster)
        self._cfg["clusters"] = updated

    def add_cluster(self, name, token):
        """
        Add a cluster to the listing.
        """
        if not self._cfg:
            return

        if "clusters" not in self._cfg:
            self._cfg["clusters"] = []

        # Ensure we don't have the cluster already
        for cluster in self._cfg["clusters"]:
            if cluster["name"] == name:
                raise ValueError(f"cluster {name} already exists - remove it first")
        self._cfg["clusters"].append({"name": name, "token": token})

    def get_clusters(self, names):
        """
        Get clusters, optionally filtering to a set of names
        """
        clusters = []
        if not self._cfg:
            return clusters

        for cluster in self._cfg.get("clusters", []):
            name = cluster.get("name")
            token = cluster.get("token")
            if not name or not token:
                continue

            # Are we filtering down to a set?
            if names and name not in names:
                continue
            clusters.append(cluster)
        return clusters

    def get_database(self):
        """
        Get the database defined in the config, if defined
        """
        # Cut out early if we don't have a config
        if self._cfg is None:
            return
        return self.cfg.get("graphdatabase")
