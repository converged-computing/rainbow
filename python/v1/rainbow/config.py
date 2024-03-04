import os

import rainbow.utils as utils


class RainbowConfig:
    """
    A RainbowClient is able to interact with a Rainbow cluster from Python.
    """

    def __init__(self, config_file=None):
        """
        Create a new rainbow client to interact with a rainbow cluster.
        """
        self.config_file = config_file
        self._cfg = None
        if self.config_file:
            self.load()

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
