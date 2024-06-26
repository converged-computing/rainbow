import json
import os

import grpc
import jobspec.core as js
import jobspec.core.converter as converter

import rainbow.auth as auth
import rainbow.backends as backends
import rainbow.config as config
import rainbow.defaults as defaults
import rainbow.types as types
import rainbow.utils as utils
from rainbow.protos import rainbow_pb2, rainbow_pb2_grpc


class RainbowClient:
    """
    A RainbowClient is able to interact with a Rainbow cluster from Python.
    """

    def __init__(self, host="localhost:50051", config_file=None, database=None, use_ssl=False):
        """
        Create a new rainbow client to interact with a rainbow cluster.
        """
        self.cfg = config.RainbowConfig(config_file)
        self.host = self.cfg.get("graphdatabase", {}).get("options", {}).get("host") or host

        # load the graph database backend
        self.set_database(database)
        self.load_backend()
        self.use_ssl = use_ssl

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

        with auth.grpc_channel(self.host, self.use_ssl) as channel:
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

        with auth.grpc_channel(self.host, self.use_ssl) as channel:
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

        with auth.grpc_channel(self.host, self.use_ssl) as channel:
            stub = rainbow_pb2_grpc.RainbowSchedulerStub(channel)
            response = stub.Register(registerRequest)
        return response

    def update_state(self, cluster, state_data, secret):
        """
        Update a cluster state
        """
        if not cluster:
            raise ValueError("A cluster name is required to register")
        if not secret:
            raise ValueError("A secret is required to register")

        # State file can be a file path or loaded state metadata
        if not isinstance(state_data, dict) and not os.path.exists(state_data):
            raise ValueError(f"State metadata file {state_data} does not exist.")

        if isinstance(state_data, dict):
            payload = state_data
        else:
            payload = utils.read_file(state_data)

        # These are the variables for our cluster - name for now
        request = rainbow_pb2.UpdateStateRequest(
            cluster=cluster,
            secret=secret,
            payload=json.dumps(payload),
        )
        with auth.grpc_channel(self.host, self.use_ssl) as channel:
            stub = rainbow_pb2_grpc.RainbowSchedulerStub(channel)
            response = stub.UpdateState(request)
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

        with auth.grpc_channel(self.host, self.use_ssl) as channel:
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

    def submit_jobspec(
        self,
        jobspec,
        name=None,
        match_algo=None,
        select_algo=None,
        select_options=None,
        satisfy_only=False,
    ):
        """
        Submit a jobspec directly. This is useful if you want to generate
        it custom with your own special logic.
        """
        # Generate a random name if does not exist
        name = name or jobspec.name

        # Ask the database backend if our jobspec can be satisfied
        match_algo = match_algo or self.cfg.match_algorithm
        select_algo = select_algo or self.cfg.selection_algorithm
        select_options = select_options or self.cfg.selection_algorithm_options
        satisfy_response = self.backend.satisfies(jobspec, match_algo)
        matches = satisfy_response.clusters

        # No matches?
        if not matches:
            print("No clusters match the request")
            return satisfy_response

        # TODO these need to have (again) the token and name checked
        # This is backwards because we check the token AFTER getting it, and it needs
        # to go to the graph (request above). I haven't implemented this yet.
        matches = self.cfg.get_clusters(matches)
        clusters = []
        for match in matches:
            clusters.append(
                rainbow_pb2.SubmitJobRequest.Cluster(name=match["name"], token=match["token"])
            )

        # THEN contact rainbow with clusters
        # These are submit variables. A more substantial submit script would have argparse, etc.

        submitRequest = rainbow_pb2.SubmitJobRequest(
            name=name,
            satisfy_only=satisfy_only,
            clusters=clusters,
            select_algorithm=select_algo,
            select_options=select_options,
            jobspec=jobspec.to_yaml(),
        )

        with auth.grpc_channel(self.host, self.use_ssl) as channel:
            stub = rainbow_pb2_grpc.RainbowSchedulerStub(channel)
            response = stub.SubmitJob(submitRequest)

        res = types.SatisfyResponse(
            cluster=response.cluster,
            total_matches=satisfy_response.total_matches,
            total_mismatches=satisfy_response.total_mismatches,
            total_clusters=satisfy_response.total_clusters,
            status=response.status,
            clusters=response.clusters,
        )
        return res

    def submit_job(
        self, command, nodes=1, tasks=1, match_algo=None, select_algo=None, satisfy_only=False
    ):
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
        return self.submit_jobspec(
            jobspec, match_algo=match_algo, select_algo=select_algo, satisfy_only=satisfy_only
        )
