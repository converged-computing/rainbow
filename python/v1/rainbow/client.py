import os

import grpc

import rainbow.backends as backends
import rainbow.config as config
import rainbow.defaults as defaults
import rainbow.jobspec as js
import rainbow.jobspec.converter as converter
import rainbow.utils as utils
from rainbow.protos import rainbow_pb2, rainbow_pb2_grpc

# TODO need to register the databases here...


class RainbowClient:
    """
    A RainbowClient is able to interact with a Rainbow cluster from Python.
    """

    def __init__(self, host="localhost:50051", config_file=None, database=None):
        """
        Create a new rainbow client to interact with a rainbow cluster.
        """
        self.cfg = config.RainbowConfig(config_file)

        # Default database is in memory
        self.host = host

        # load the graph database backend
        self.set_database(database)
        self.load_backend()

    def register(self, cluster, secret, cluster_nodes):
        """
        Register a cluster to the Rainbow Scheduler.
        """
        if not cluster:
            raise ValueError("A cluster name is required to register")
        if not secret:
            raise ValueError("A secret is required to register")
        if not os.path.exists(cluster_nodes):
            raise ValueError(f"Cluster nodes file {cluster_nodes} for registration does not exist.")

        # Read in the nodes to string
        nodes = utils.read_file(cluster_nodes)

        # These are the variables for our cluster - name for now
        registerRequest = rainbow_pb2.RegisterRequest(
            name=cluster,
            secret=secret,
            nodes=nodes,
        )

        with grpc.insecure_channel(self.host) as channel:
            stub = rainbow_pb2_grpc.RainbowSchedulerStub(channel)
            response = stub.Register(registerRequest)
        return response

    def load_backend(self):
        """
        Load the backend
        """
        # This will error if you do something wonky
        self.backend = backends.get_backend(self.database, self.database_options)

    def set_database(self, database=None):
        """
        Load the graph database backend, and options
        """
        # It's not technically required, so touch lightly
        # Only go through this if a database is not provided
        options = {"host": defaults.database_host}
        db = self.cfg.get_database()
        if db is not None:
            database = db.get("name")
            options = db.get("options") or options

        # If we get here and no database, use the default
        if not database:
            database = defaults.database_backend
        self.database = database
        self.database_options = options

    def submit_jobspec(self, jobspec):
        """
        Submit a jobspec directly. This is useful if you want to generate
        it custom with your own special logic.
        """
        # TODO check if len response clusters is 0...
        # TODO need a way to validate clusters we have permission to access here,
        # ideally via the backend graph...

        # Ask the database backend if our jobspec can be satisfied
        response = self.backend.satisfies(jobspec)
        matches = response.clusters
        print(response)

        # No matches?
        if not matches:
            print("No clusters match the request")
            return response

        # These need to have (again) the token and name
        matches = self.cfg.get_clusters(matches)
        clusters = []
        for match in matches:
            clusters.append(
                rainbow_pb2.SubmitJobRequest.Cluster(name=match["name"], token=match["token"])
            )

        # THEN contact rainbwo with clusters
        # These are submit variables. A more substantial submit script would have argparse, etc.
        submitRequest = rainbow_pb2.SubmitJobRequest(
            name=jobspec.name,
            clusters=clusters,
            jobspec=jobspec.to_yaml(),
        )

        with grpc.insecure_channel(self.host) as channel:
            stub = rainbow_pb2_grpc.RainbowSchedulerStub(channel)
            response = stub.SubmitJob(submitRequest)
        return response

    def submit_job(self, command, nodes=1, tasks=1):
        """
        Submit a simple job to rainbow. This includes:

        1. Converting the basic attributes into a jobspec (and validating)
        2. Reading and validating the cluster config
        3. Asking the graph for which clusters satisfy
        4. Then submitting to rainbow
        """
        if not command:
            raise ValueError("A command is required")

        # Pretty print for the user (arguably we don't need this, oh well)
        cmd = command
        if isinstance(cmd, list):
            command = " ".join(cmd)
            print(f"⭐️ Submitting job: {cmd}")

        # Generate the jobspec dictionary
        raw = converter.new_simple_jobspec(nodes=nodes, command=command, tasks=tasks)
        jobspec = js.Jobspec(raw)
        return self.submit_jobspec(jobspec)
