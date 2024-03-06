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
2024/03/05 01:18:59 🌈️ starting client (localhost:50051)...
2024/03/05 01:18:59 registering cluster: keebler
2024/03/05 01:18:59 status: REGISTER_SUCCESS
2024/03/05 01:18:59 secret: e6794098-b209-463d-b761-98d2d418b26f
2024/03/05 01:18:59  token: rainbow
2024/03/05 01:18:59 Saving cluster secret to ./docs/examples/scheduler/rainbow-config.yaml
```

If you ran this using the rainbow client you would do:

```bash
rainbow register --cluster-name keebler --cluster-nodes ./docs/examples/scheduler/cluster-nodes.json --config-path ./docs/examples/scheduler/rainbow-config.yaml --save
```

Note in the above we are providing a config file path and `--save` so our cluster secret gets saved there. Be careful always about overwriting any configuration file.
The new secret will be provided in the console as a more conservative approach. If you are watching the server, you'll see that the registration happens (token, secret, etc) and then the nodes are sent over to rainbow.

```console
go run cmd/server/server.go --global-token rainbow
2024/03/05 01:18:54 creating 🌈️ server...
2024/03/05 01:18:54 🧩️ selection algorithm: random
2024/03/05 01:18:54 🧩️ graph database: memory
2024/03/05 01:18:54 ✨️ creating rainbow.db...
2024/03/05 01:18:54    rainbow.db file created
2024/03/05 01:18:54    🏓️ creating tables...
2024/03/05 01:18:54    🏓️ tables created
2024/03/05 01:18:54 ⚠️ WARNING: global-token is set, use with caution.
2024/03/05 01:18:54 starting scheduler server: rainbow v0.1.1-draft
2024/03/05 01:18:54 🧠️ Registering memory graph database...
2024/03/05 01:18:54 Adding special vertex root at index 0
2024/03/05 01:18:54 server listening: [::]:50051
2024/03/05 01:18:59 📝️ received register: keebler
2024/03/05 01:18:59 Received cluster graph with 44 nodes and 86 edges
2024/03/05 01:18:59 SELECT count(*) from clusters WHERE name = 'keebler': (0)
2024/03/05 01:18:59 INSERT into clusters (name, token, secret) VALUES ("keebler", "rainbow", "e6794098-b209-463d-b761-98d2d418b26f"): (1)
2024/03/05 01:18:59 Preparing to load 44 nodes and 86 edges
2024/03/05 01:18:59 Adding special vertex keebler at index 12
2024/03/05 01:18:59 We have made an in memory graph (subsystem nodes) with 45 vertices!
{
 "keebler": {
  "Name": "keebler",
  "Counts": {
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

Submission has two steps that are discussed below.

### 1. Satisfy Request

The satisfy request interacts with the graph database and determines if any clusters can satisfy the jobspec.
To submit a job, we need the client `token` associated with a cluster. We are going to use the following strategy, and allow the following submission types:

- **simple**: for basic users, a command and the most basic of parameters will be provided and converted to a Jobspec.
- **jobspec**: for advanced users, a Jobspec can be provided directly.
- **Kubernetes job**: for Kubernetes users, a [batchv1/Job](https://kubernetes.io/docs/concepts/workloads/controllers/job/) can be provided that will be converted to a Jobspec.

We will likely start with the first (simple) and then work on implementations for the latter. The converters will be implemented alongside the [Jobspec](https://github.com/compspec/jobspec-go/issues/1)
library and used here.


```bash
# Look at help
go run ./cmd/rainbow/rainbow.go submit --help
```
```
usage: rainbow submit [-h|--help] [--token "<value>"] [-n|--nodes <integer>]
               [-t|--tasks <integer>] [-c|--command "<value>"] [--job-name
               "<value>"] [--config-path "<value>"] [--host "<value>"]
               [--cluster-name "<value>"] [--config "<value>"]
               [--graph-database "<value>"]

               Submit a job to a rainbow scheduler

Arguments:

  -h  --help            Print help information
      --token           Client token to submit jobs with.. Default:
                        chocolate-cookies
  -n  --nodes           Number of nodes to request. Default: 1
  -t  --tasks           Number of tasks to request (per node? total?)
  -c  --command         Command to submit. Default: chocolate-cookies
      --job-name        Name for the job (defaults to first command)
      --config-path     Rainbow config file. Default: rainbow-config.yaml
      --host            Scheduler server address (host:port). Default:
                        localhost:50051
      --cluster-name    Name of cluster to register
      --config          Configuration file for cluster credentials
      --graph-database  Graph database backend to use
```

Let's try doing that. Note that since we just created a cluster with a global token `rainbow`, and since we want to submit to rainbow and potentially
hit one of many clusters, a single command line request won't suffice anymore, e.g.,:

```bash
go run ./cmd/rainbow/rainbow.go submit --token "712747b7-b2a9-4bea-b630-056cd64856e6" --command hostname --cluster-name keebler
```

We are instead going to use a config file provided in the examples directory that can have more than one cluster defined. The idea is that you don't
know where the work will best run, and are querying rainbow. Note that for a more final design, we would want the interaction to go through another service
that connects to the same database (to check the clusters you have access to) and then to the graph database directly without touching rainbow.
However for development, we are going to still interact with the in-memory database grpc to keep things simple, since the authentication (token)
is known there (and we have not [sent it to a truly external graph database](https://dgraph.io/docs/v21.03/graphql/authorization/authorization-overview/)).
Note that the flow (for searching the cluster graph) is going to go directly from the client to the graph, e.g.,:

```bash
rainbow submit -> graph database GRPC or query -> response
```

And where the middle step is provided from will depend on the graph - the in-memory database will be GRPC from rainbow, for example.
Assuming that rainbow is running with the in-memory database and we've registered (and our config file has the correct token),
here is how we ask for a simple job:

```bash
go run ./cmd/rainbow/rainbow.go submit --config-path ./docs/examples/scheduler/rainbow-config.yaml --nodes 2 --tasks 24 --command "echo hello world"
```
```console
2024/02/29 21:04:11 🌈️ starting client (localhost:50051)...
2024/02/29 21:04:11 submit job: echo hello world
2024/02/29 21:04:11 🎯️ We found 1 matches! [keebler]
2024/02/29 21:04:11
```

On the server side, we see that it also registers a match! Note that this is coming from rainbow because the in-memory database GRPC hits there, but doesn't necessarily have to.
```console
🍇️ Satisfy request to Graph 🍇️
 jobspec: {"version":1,"resources":[{"type":"node","count":2,"with":[{"type":"slot","count":1,"with":[{"type":"core","count":24}],"label":"echo"}]}],"tasks":[{"command":["echo","hello","world"],"slot":"echo","count":{"per_slot":1}}],"attributes":{"system":{}}}
  match: 🎯️ cluster keebler has enough resources and is a match
```

Now let's ask for a request we know cannot be satisfied. Here is the client:

```bash
go run ./cmd/rainbow/rainbow.go submit --config-path ./docs/examples/scheduler/rainbow-config.yaml --nodes 100 --tasks 24 --command "echo hello world"
```
```console
2024/02/29 21:05:44 🌈️ starting client (localhost:50051)...
2024/02/29 21:05:44 submit job: echo hello world
2024/02/29 21:05:44 😥️ There were no matches for this job
2024/02/29 21:05:44
```
On the server side, we see it cannot be satisfied. We just don't have that many nodes!

```console
🍇️ Satisfy request to Graph 🍇️
 jobspec: {"version":1,"resources":[{"type":"node","count":100,"with":[{"type":"slot","count":1,"with":[{"type":"core","count":24}],"label":"echo"}]}],"tasks":[{"command":["echo","hello","world"],"slot":"echo","count":{"per_slot":1}}],"attributes":{"system":{}}}
cluster keebler does not have sufficient resource type node - actual 3 vs needed 100
  match: 😥️ no clusters could satisfy this request. We are sad
```

Note that the above has a two step process:

- A quick check against clusters in the graph database if total resources can be satisfied.
- For that set, a (Vanessa written and janky) "DFS" that likely has bugs that traverses the graph

This will be improved upon with Fluxion and actual graph databases, but this is OK for the prototype.

### 2. Assignment

When the initial satisfy request is done (the step above) and we have a list of clusters, we can then tell rainbow about them.
Rainbow then uses higher level metadata about each cluster (that reflects state) along with a selection algorithm
to assign to a specific cluster. The cluster assignment is added to the database to be picked up by the cluster
on it's next pass for jobs. Although this is a pull model (the assigned cluster is pulling work) the assignment and
decision is done - the cluster is going to accept the job. The selection algorithm can be provided on the command line,
or more likely is defined in the rainbow cluster configuration file. As an example:

```yaml
scheduler:
    secret: chocolate-cookies
    name: rainbow-cluster
    algorithm:
      name: randon
      options:
         key: value

graphdatabase:
    name: memory
    options:
      host: "127.0.0.1:50051"

clusters:
  - name: keebler
    token: rainbow
```

In the above, we see the default algorithm (if it were not provided) that is random, meaning that the list of clusters
is selected from randomly. We will likely have some representation of state provided in the graph or rainbow, and combined with
this ability to customize algorithms, a more intelligent assignment to clusters.

## Receive Jobs

> Receive: Request and Accept jobs

The next endpoint is to receive jobs, and although this is a pull design (assuming most clusters will not expose services but can pull from them) there has to be a contract between rainbow and the cluster to honor doing this at some frequency, and some number to accept. Assuming that we have registered and submit a job (that will be assigned to cluster keebler) we can then receive the job to run. The above assumes that you have used `--save` on the registration step to save the cluster secret into your configuration file. If you haven't, you can provide it with `--request-secret` on the command line.

```bash
$ go run ./cmd/rainbow/rainbow.go receive --config-path ./docs/examples/scheduler/rainbow-config.yaml --max-jobs 3
```
```console
2024/03/05 01:45:58 🌈️ starting client (localhost:50051)...
2024/03/05 01:45:58 receive jobs: 10
2024/03/05 01:45:58 🌀️ Received 3 jobs!
2024/03/05 01:45:58 2 : {"id":2,"cluster":"keebler","name":"echo","jobspec":"attributes:\n  system: {}\nresources:\n- count: 1\n  type: node\n  with:\n  - count: 1\n    label: echo\n    type: slot\n    with:\n    - count: 24\n      type: core\ntasks:\n- command:\n  - echo\n  - hello\n  - moon\n  count:\n    per_slot: 1\n  slot: echo\nversion: 1\n","command":""}
2024/03/05 01:45:58 3 : {"id":3,"cluster":"keebler","name":"echo","jobspec":"attributes:\n  system: {}\nresources:\n- count: 1\n  type: node\n  with:\n  - count: 1\n    label: echo\n    type: slot\n    with:\n    - count: 24\n      type: core\ntasks:\n- command:\n  - echo\n  - hello\n  - moon\n  count:\n    per_slot: 1\n  slot: echo\nversion: 1\n","command":""}
2024/03/05 01:45:58 1 : {"id":1,"cluster":"keebler","name":"echo","jobspec":"attributes:\n  system: {}\nresources:\n- count: 1\n  type: node\n  with:\n  - count: 1\n    label: echo\n    type: slot\n    with:\n    - count: 24\n      type: core\ntasks:\n- command:\n  - echo\n  - hello\n  - moon\n  count:\n    per_slot: 1\n  slot: echo\nversion: 1\n","command":""}
2024/03/05 01:45:58 ✅️ Accepting 3 jobs!
2024/03/05 01:45:58 status:RESULT_TYPE_SUCCESS
```

The above can be prettier printed, especially since the jobspec is sent back now! And on the server side:

```console
024/03/05 01:45:58 SELECT * from clusters WHERE name LIKE "keebler" LIMIT 1: keebler
2024/03/05 01:45:58 🌀️ accepting 3 for cluster keebler
2024/03/05 01:45:58 DELETE FROM jobs WHERE cluster = 'keebler' AND idJob in (2,3,1): (3)
```

Note that if you don't define the max jobs (so it is essentially 0) you will get all jobs.
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
