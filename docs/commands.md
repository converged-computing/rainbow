# Commands

The following commands are currently supported. For Python, see the [README](https://github.com/converged-computing/rainbow/tree/main/python/v1) in the Python directory.

## Prepare to Register

The registration step happens when a cluster joins the rainbow scheduler. The registering cluster submits a [JGF format](https://github.com/converged-computing/jsongraph-go) resource graph.
This allows the rainbow scheduler to "intelligently"  (subjective right there I know, especially if I wrote it 😜️) filter down clusters to choose your cluster (or not) based on the resources you provide. If you
cannot provide a cluster graph (a likely case if your cluster is ephemeral, created on demand, or does not have consistent resources) then we will likely (eventually) have a specification for limits of requests still.
This is not developed yet.

In the example below, we will extract node level metadata with `compspec extract` ([see here](https://github.com/compspec/compspec-go/tree/main/docs/rainbow)) and then generate the cluster JGF to send for registration with compspec create-nodes.
That two step process looks like this, and note this is faux cluster node metadata since I'm running it three times on my local machine :). The faux use case is that my cluster has three identical nodes.

```bash
mkdir -p ./docs/rainbow/cluster
compspec extract --name library --name nfd[cpu,memory,network,storage,system] --name system[cpu,processor,arch,memory] --out ./docs/rainbow/cluster/node-1.json
compspec extract --name library --name nfd[cpu,memory,network,storage,system] --name system[cpu,processor,arch,memory] --out ./docs/rainbow/cluster/node-2.json
compspec extract --name library --name nfd[cpu,memory,network,storage,system] --name system[cpu,processor,arch,memory] --out ./docs/rainbow/cluster/node-3.json
```

Then we give that directory to compspec, and used the cluster creation plugin to output the JGF of the cluster:

```bash
compspec create nodes --cluster-name cluster-red --node-dir ./docs/rainbow/cluster/ --nodes-output ./cluster-nodes.json
```

That example is provided in [examples](examples/scheduler/cluster-nodes.json). This is the cluster metadata that we need to send over to the rainbow scheduler on the register step,
discussed next.

## Config

If you want to generate a new configuration file:

```bash
$ ./bin/rainbow config init
2024/02/26 15:55:36 Writing rainbow config to rainbow-config.yaml
```
This generates the following file.

```yaml
scheduler:
    secret: chocolate-cookied
    name: rainbow-cluster
graphdatabase:
    name: memory
clusters: []
```

Note that the name of the database corresponds to your choice of graph database. For each, you should read about [databases](databases.md) to 
run a corresponding databaset that your application can interact with.

## Register

The registration step happens when a cluster joins. Using the make command it is expected that you have the cluster-nodes.json in the path shown above.
You should also be running a server with a database selected (e.g., `make server` to use the default in memory model):

```bash
make register
```
```console
go run cmd/rainbow/rainbow.go register --cluster-name keebler --cluster-nodes ./docs/examples/scheduler/cluster-nodes.json
2024/02/12 22:17:43 🌈️ starting client (localhost:50051)...
2024/02/12 22:17:43 registering cluster: keebler
2024/02/12 22:17:43 status: REGISTER_SUCCESS
2024/02/12 22:17:43 secret: 54c4568a-14f2-465f-aa1e-5e6e0e3efd33
2024/02/12 22:17:43  token: 67e0f258-96c3-4d88-8253-287a95653138
```

If you ran this using the rainbow client you would do:

```bash
rainbow register --cluster-name keebler --cluster-nodes ./docs/examples/scheduler/cluster-nodes.json
```

If you are watching the server, you'll see that the registration happens (token, secret, etc) and then the nodes are sent over
to rainbow. 

```console
2024/02/28 23:26:17 creating 🌈️ server...
2024/02/28 23:26:17 ✨️ creating rainbow.db...
2024/02/28 23:26:17    rainbow.db file created
2024/02/28 23:26:17    create cluster table...
2024/02/28 23:26:17    cluster table created
2024/02/28 23:26:17    create jobs table...
2024/02/28 23:26:17    jobs table created
2024/02/28 23:26:17 ⚠️ WARNING: global-token is set, use with caution.
2024/02/28 23:26:17 starting scheduler server: rainbow v0.1.1-draft
2024/02/28 23:26:17 🧠️ Registering memory graph database...
2024/02/28 23:26:17 Adding special vertex root at index 0
2024/02/28 23:26:17 server listening: [::]:50051
2024/02/28 23:26:19 📝️ received register: keebler
2024/02/28 23:26:19 Received cluster graph with 44 nodes and 86 edges
2024/02/28 23:26:19 SELECT count(*) from clusters WHERE name = 'keebler': (0)
2024/02/28 23:26:19 INSERT into clusters (name, token, secret) VALUES ("keebler", "rainbow", "3f78d433-f24b-4664-858c-0577971e259e"): (1)
2024/02/28 23:26:19 Preparing to load 44 nodes and 86 edges
2024/02/28 23:26:19 Adding special vertex keebler at index 1
2024/02/28 23:26:19 We have made an in memory graph (subsystem nodes) with 46 vertices!
2024/02/29 01:17:37 Preparing to load 44 nodes and 86 edges
2024/02/29 01:17:37 Adding special vertex keebler at index 1
2024/02/29 01:17:37 We have made an in memory graph (subsystem nodes) with 46 vertices!
{
 "keebler": {
  "Name": "keebler",
  "Counts": {
   "cluster": 1,
   "core": 36,
   "node": 3,
   "rack": 1,
   "socket": 3
  }
 }
}
```

This is actually a modular process that works as follows:

1. When we create the server, we select a database backend. The default is a "memory" (in memory graph database)
2. The client is interacting with rainbow via GRPC, and doesn't need to know about the database.
3. The rainbow client hits the rainbow server via GRPC, and sends over the cluster nodes, from JSON into a [json graph version 2](https://github.com/converged-computing/jsongraph-go)
4. Once the registration is validated, the graph database service is sent the nodes to add to the graph.

For the last step, the default in memory database still serves GRPC (anticipating a client will interact with it in a read only fashion in the future to assess jobspecs), but
since this default database plugin is part of rainbow, we interact with the in memory database directly to write, and we do this because it's faster than GRPC.
At the end, you see that the nodes are sent over, and added to the graph, and that's the most that you should care about! In the client window, the registration
is successful:

```console
2024/02/27 01:26:11 🌈️ starting client (localhost:50051)...
2024/02/27 01:26:11 registering cluster: keebler
2024/02/27 01:26:11 status: REGISTER_SUCCESS
2024/02/27 01:26:11 secret: 4a5d5f6d-c510-45f2-9cca-cd53f4a40e79
2024/02/27 01:26:11  token: rainbow
```

In case you don't remember, here is what the response metadata mean. Both of these parameters you can save to a `rainbow-config.yaml` for future, programmatic use.

- `token` is what is given to clients to submit jobs
- `secret` is a secret just for your cluster / instance / place you can receive jobs to receive them!

While this isn't the final design, for an early first crack (that is likely making graph experts spin in their graves) I am creating a single, dominant subsystem (node resources)
off of which we can add as many clusters as we like. For salient vertices that need to be found again, we have a small lookup. This is primarily the root and named clusters off of that.
For persistence of data, if the config you provide has a backup file for the graph database, it will be saved and loaded as a [gob](https://pkg.go.dev/encoding/gob).
I'm hoping to come up with a more elegant "multi-cluster" graph design, and also add support for multiple subsystems, which should be possible by linking vertices between subsystem graphs.
I actually don't know because I haven't thought about it yet. Finally, the metadata for nodes obviously needs to be added. As a final distinguishing note, I decided to use vertex instead of node for
the following reasons:

- A **node** is typically referring to the physical node of an HPC system or Kubernetes
- A **vertex** is "that thing" but represented in the graph.

In Computer Science I think they are used interchangeably. For next steps we will be updating the memory graph database to be a little more meaty (adding proper metadata and likely a summary of resources at the top as a quick "does it satisfy" heuristic)
and then working on the next interaction, the client submit command, which is going to hit the `Satisfies` endpoint. I will write up more about the database and submit design after that.


## Submit Job

To submit a job, we need the client `token` associated with a cluster.

```bash
# Look at help
go run ./cmd/rainbow/rainbow.go submit --help
```
```
usage: rainbow submit [-h|--help] [-s|--secret "<value>"] [-n|--nodes
               <integer>] [-t|--tasks <integer>] [-c|--command "<value>"]
               [--job-name "<value>"] [--host "<value>"] [--cluster-name
               "<value>"]

               Submit a job to a rainbow scheduler

Arguments:

  -h  --help          Print help information
      --token         Client token to submit jobs with.. Default:
                      chocolate-cookies
  -n  --nodes         Number of nodes to request. Default: 1
  -t  --tasks         Number of tasks to request (per node? total?)
  -c  --command       Command to submit. Default: chocolate-cookies
      --job-name      Name for the job (defaults to first command)
      --host          Scheduler server address (host:port). Default:
                      localhost:50051
      --cluster-name  Name of cluster to register. Default: keebler
```

Let's try doing that.

```bash
go run ./cmd/rainbow/rainbow.go submit --token "712747b7-b2a9-4bea-b630-056cd64856e6" --command hostname
```
```console
2024/02/11 21:43:17 🌈️ starting client (localhost:50051)...
2024/02/11 21:43:17 submit job: hostname
2024/02/11 21:43:17 status:SUBMIT_SUCCESS
```

Hooray! On the server log side we see...

```console
SELECT * from clusters WHERE name LIKE "keebler" LIMIT 1: keebler
2024/02/11 21:43:17 📝️ received job hostname for cluster keebler
```

Now we have a job in the database, and it's oriented for a specific cluster.
We can next (as the cluster) request to receive some number of max jobs. Let's
emulate that.

## Request Jobs

> Also List Jobs

We now are pretending to be the cluster that originally registered, and we want to request some number of max jobs
to look at. This doesn't mean we have to run them, but we want to ask for some small set to consider for running.
Right now this just does a query for the count, but in the future we can have actual filters / query parameters
for the jobs (nodes, time, etc.) that we want to ask for. Have some fun and submit a few jobs above, and then request
to see them:

```console
$ go run ./cmd/rainbow/rainbow.go request --request-secret 3cc06871-0990-4dc2-94d5-eec653c5d7a0 --cluster-name keebler --max-jobs 3
2024/02/12 23:29:59 🌈️ starting client (localhost:50051)...
2024/02/12 23:29:59 request jobs: 3
2024/02/12 23:29:59 🌀️ Found 3 jobs!
2024/02/12 23:29:59 1 : {"id":1,"cluster":"keebler","name":"hostname","nodes":1,"tasks":0,"command":"hostname"}
2024/02/12 23:29:59 2 : {"id":2,"cluster":"keebler","name":"sleep","nodes":1,"tasks":0,"command":"sleep 10"}
2024/02/12 23:29:59 3 : {"id":3,"cluster":"keebler","name":"dinosaur","nodes":1,"tasks":0,"command":"dinosaur things"}
```

And on the server side:

```console
2024/02/12 23:27:29 SELECT * from clusters WHERE name LIKE "keebler" LIMIT 1: keebler
2024/02/12 23:27:29 🌀️ requesting 3 max jobs for cluster keebler
```

Note that if you don't define the max jobs (so it is essentially 0) you will get all jobs. This is akin to listing jobs.
Awesome! Next we can put that logic in a flux instance (from the Python grpc to start) and then have Flux
accept some number of them. The response back to the rainbow scheduler will be those to accept, which will then be removed from the database. For another day.

## Accept Jobs

A derivative of the above is to request and accept jobs. This can be done with the example client above, and adding `--accept N`.

```console
$ go run ./cmd/rainbow/rainbow.go request --request-secret 3cc06871-0990-4dc2-94d5-eec653c5d7a0 --cluster-name keebler --max-jobs 3 --accept 1
```
```console
2024/02/13 12:29:29 🌀️ Found 3 jobs!
2024/02/13 12:29:29 1 : {"id":1,"cluster":"keebler","name":"hostname","nodes":1,"tasks":0,"command":"hostname"}
2024/02/13 12:29:29 2 : {"id":2,"cluster":"keebler","name":"sleep","nodes":1,"tasks":0,"command":"sleep 10"}
2024/02/13 12:29:29 3 : {"id":3,"cluster":"keebler","name":"dinosaur","nodes":1,"tasks":0,"command":"dinosaur things"}
2024/02/13 12:29:29 ✅️ Accepting 1 jobs!
2024/02/13 12:29:29    1
2024/02/13 12:29:29 status:RESULT_TYPE_SUCCESS
```

What this does is randomly select from the set you receive, and send back a response to the server to accept it, meaning the identifier is removed from the database. The server shows the following:

```console
2024/02/13 12:29:29 🌀️ accepting 1 for cluster keebler
2024/02/13 12:29:29 DELETE FROM jobs WHERE cluster = 'keebler' AND idJob in (1): (1)
```

The logic you would expect is there - that you can't accept greater than the number available.
You could try asking for a high level of max jobs again, and see that there is one fewer than before. It was deleted from the database.

[home](/README.md#rainbow-scheduler)
