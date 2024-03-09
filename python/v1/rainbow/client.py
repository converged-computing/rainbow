import json
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
        self.host = host or self.cfg.get("scheduler", {}).get("host")

        # load the graph database backend
        self.set_database(database)
        self.load_backend()

    def receive_jobs(self, max_jobs=None):
        """
        Receive jobs (query and send accept response back to rainbow)
        """
        jobs = []

        # This requires a cluster to be defined
        cfg = self.cfg
        cluster = cfg._cfg.get("cluster")
        if "name" not in cluster or "secret" not in cluster:
            raise ValueError("'cluster' defined in the rainbow config needs a name and secret")

        # These are submit variables. A more substantial submit script would have argparse, etc.
        request = rainbow_pb2.ReceiveJobsRequest(secret=cluster["secret"], cluster=cluster["name"])
        # This defaults to 0, so only set of non 0 or not None
        if max_jobs:
            request.maxJobs = max_jobs

        with grpc.insecure_channel(self.host) as channel:
            stub = rainbow_pb2_grpc.RainbowSchedulerStub(channel)
            response = stub.ReceiveJobs(request)

            # Case 1: no jobs to receive
            if response.status == 0:
                print("There are no jobs to receive.")
                return jobs

            if response.status != 1:
                print("Issue with requesting jobs:")
                return jobs

        print("Status: REQUEST_JOBS_SUCCESS")
        print(f"Received {len(response.jobs)} jobs to accept...")
        jobs = [json.loads(job) for job in list(response.jobs.values())]

        # Tell rainbow we accepted them
        jobids = list(response.jobs.keys())
        request = rainbow_pb2.AcceptJobsRequest(
            cluster=cluster["name"],
            secret=cluster["secret"],
            jobids=jobids,
        )
        with grpc.insecure_channel(self.host) as channel:
            stub = rainbow_pb2_grpc.RainbowSchedulerStub(channel)
            response = stub.AcceptJobs(request)
        return jobs

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

    def register_subsystem(self, cluster, subsystem, secret, nodes):
        """
        Register subsystem nodes to rainbow
        """
        if not cluster:
            raise ValueError("A cluster name is required to register")
        if not secret:
            raise ValueError("A secret is required to register")
        if not os.path.exists(nodes):
            raise ValueError(f"Subsystem nodes file {nodes} for registration does not exist.")

        # Read in the nodes to string
        nodes = utils.read_file(nodes)

        # These are the variables for our cluster - name for now
        registerRequest = rainbow_pb2.RegisterRequest(
            name=cluster,
            secret=secret,
            nodes=nodes,
            subsystem=subsystem,
        )

        with grpc.insecure_channel(self.host) as channel:
            stub = rainbow_pb2_grpc.RainbowSchedulerStub(channel)
            response = stub.RegisterSubsystem(registerRequest)
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
        # Ask the database backend if our jobspec can be satisfied
        response = self.backend.satisfies(jobspec)
        matches = response.clusters
        print(response)

        # No matches?
        if not matches:
            print("No clusters match the request")
            return response

        # TODO these need to have (again) the token and name checked
        # This is backwards because we check the token AFTER getting it, and it needs
        # to go to the graph (request above). I haven't implemented this yet.
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
