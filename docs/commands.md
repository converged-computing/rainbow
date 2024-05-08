# Commands

The following commands are currently supported. For Python, see the [README](https://github.com/converged-computing/rainbow/tree/main/python/v1) in the Python directory. For advanced examples, see the [advanced](#advanced.md) page.

## Run the Server

You can run the server (with defaults) as follows:

```bash
# Regular logging
make server

# Verbose logging
make server-verbose
```
```console
go run cmd/server/server.go --global-token rainbow
2024/03/30 14:56:26 creating 🌈️ server...
2024/03/30 14:56:26 🧩️ selection algorithm: random
2024/03/30 14:56:26 🧩️ graph database: memory
2024/03/30 14:56:26 ✨️ creating rainbow.db...
2024/03/30 14:56:26    rainbow.db file created
2024/03/30 14:56:26    🏓️ creating tables...
2024/03/30 14:56:26    🏓️ tables created
2024/03/30 14:56:26 ⚠️ WARNING: global-token is set, use with caution.
2024/03/30 14:56:26 starting scheduler server: rainbow v0.1.1-draft
2024/03/30 14:56:26 🧠️ Registering memory graph database...
2024/03/30 14:56:26 server listening: [::]:50051
```
It shows you the commands that are run above with go. You could also build the `rainbow` binary instead with `make build` and use that instead.
All subsequent commands require a server to be running.

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

That example is provided in [examples](examples/scheduler/cluster-nodes.json) if you want to look. The high level TLDR of this step is that you need your nodes in JGF format to register, which will
be shown after the config section, next.

## Config

If you want to generate a new configuration file:

```bash
$ ./bin/rainbow config init
2024/02/26 15:55:36 Writing rainbow config to rainbow-config.yaml
```
This generates the following file.

```yaml
scheduler:
    secret: chocolate-cookies
    name: rainbow-cluster
    algorithms:
        selection:
            name: random
        match:
            name: match
cluster: {}
graphdatabase:
    name: memory
    host: 127.0.0.1:50051
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
rainbow register --cluster-name keebler --nodes-json ./docs/examples/scheduler/cluster-nodes.json --config-path ./docs/examples/scheduler/rainbow-config.yaml --save
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

## Update State

A cluster state is intended to be a superficial view of the cluster status. It's not considered a subsystem because (for the time being) we are only considering a flat listing of key value pairs that describe a cluster. The data is also intended to be small so it can be provided via this update endpoint more frequently. As an example, an update payload may look like the following:

```json
{
  "cost-per-node": 12,
  "nodes-free": 100
}
```

While the above would not be suited for a real-world deployment (for example, there are many more costs than per node, and occupancy goes beyond nodes free)
but this will be appropriate for small tests and simulations. The metadata above will be provided, on a cluster level, for a final selection algorithm (to use or not). So after you've created your cluster, let's update the state.

```bash
make update-state
```
```console
Adding edge from socket -contains-> core
Adding edge from socket -contains-> core
2024/04/05 18:38:13 We have made an in memory graph (subsystem cluster) with 45 vertices!
Metrics for subsystem cluster{
 "cluster": 1,
 "core": 36,
 "node": 3,
 "rack": 1,
 "socket": 3
}
2024/04/05 18:38:16 📝️ received state update: keebler
Updating state cost-per-node to 12
Updating state max-jobs to 100
```
In debug logging mode (`make server-debug`) you will see the values updated, as shown above. They are also in blue, which you can't see! Note that this state metadata is provided to a selection algorithm, and we will be added more interesting ones soon for experiments!

## Register Subsystem

Adding a subsystem means adding another graph that has nodes with edges that connect (in some meaningful way) to the dominant subsystem.

### Nodes

While the dominant subsystem nodes have identifiers without a namespace (e.g., "0" through "3" for 4 nodes) a subsystem needs to own a namespace of nodes, so it should have `<subsystem-name>0` through `<subsystem-name>N`, where the subsystem name is also the root of the subsystem graph.

### Edges

The edges should reference a node in the dominant subsystem. For example, given that these nodes have these corresponding vertex identifiers (the label or unique id):

- node0 --> "2"
- node1 --> "16"
- node2 --> "30"

We would expect edges for I/O to reference them as follows - in the example below, the I/O node of type io0 (the global identifier) is attached or relevant for node0 above:

```json
{
    "edges": [
      {
        "source": "2",
        "target": "io1",
        "relation": "contains"
      },
      {
        "source": "io1",
        "target": "2",
        "relation": "in"
      }
    ]
}
```

Note that is only partial json, and validation when adding a subsystem will ensure that:

- All nodes in the subsystem are linked to the dominant subsystem graph or another subsystem node.
- All edges defined for the subsystem exist in the graph.

The root exists primarily as a handle to all of the children in the subsystem. You are not allowed to add edges to nodes that don't exist in the dominant subsystem, nor are you allowed to add subsystem nodes that are not being used (and are unlinked or have no edges). When you run the register command, you'll see the following output (e.g, I normally have two terminals and do):

```bash
# terminal 1
rm rainbow.db && make server

# terminal 2
make register && make subsystem
```

And then I'll see the following output in terminal 1:

```console
...
2024/03/08 18:34:44 📝️ received register: keebler
2024/03/08 18:34:44 Received cluster graph with 44 nodes and 86 edges
2024/03/08 18:34:44 SELECT count(*) from clusters WHERE name = 'keebler': (0)
2024/03/08 18:34:44 INSERT into clusters (name, token, secret) VALUES ("keebler", "rainbow", "d6aa12a2-cbff-4504-8a0b-1b36e8796ed8"): (1)
2024/03/08 18:34:44 Preparing to load 44 nodes and 86 edges
2024/03/08 18:34:44 We have made an in memory graph (subsystem cluster) with 45 vertices!
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
2024/03/08 18:34:45 SELECT * from clusters WHERE name LIKE "keebler" LIMIT 1: keebler
2024/03/08 18:34:45 📝️ received subsystem register: keebler
2024/03/08 18:34:45 Preparing to load 6 nodes and 30 edges
2024/03/08 18:34:45 We have made an in memory graph (subsystem io) with 7 vertices, with 15 connections to the dominant!
{
 "keebler": {
  "Name": "keebler",
  "Counts": {
   "io": 1,
   "mtl1unit": 1,
   "mtl2unit": 1,
   "mtl3unit": 1,
   "nvme": 1,
   "shm": 1
  }
 }
}
```
And in terminal 2:

```console
...
2024/03/08 18:34:44 Saving cluster secret to ./docs/examples/scheduler/rainbow-config.yaml
go run cmd/rainbow/rainbow.go register subsystem --subsystem io --nodes-json ./docs/examples/scheduler/cluster-io-subsystem.json --config-path ./docs/examples/scheduler/rainbow-config.yaml
2024/03/08 18:34:45 🌈️ starting client (localhost:50051)...
2024/03/08 18:34:45 registering subsystem to cluster: keebler
2024/03/08 18:34:45 status:REGISTER_SUCCESS
```

Next we are going to submit jobs - first without anything special, and then taking into account our subsystem. This design is based on the [thinking here](https://github.com/flux-framework/flux-sched/discussions/1153#discussioncomment-8726678).

1. Tasks have resources, because technically speaking, a subsystem is another kind of resource. It's just the needs specific to a task.
2. Each resource entry is scoped to the equivalently named subsystem. The subsystem can control the algorithms and metadata provided in this section.
3. A user can request a kind of resource defined in a subsystem that is present on the cluster (e.g., the storage type, power, GPU, etc.) without knowing about additional plugins / data files that are needed.

## Submit Job

Submission has two steps that are discussed below. We will talk about:

1. The satisfy request
2. Assignment
3. Satisfy and Assignment in the context of asking for subsystem resources

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
    algorithms:
      selection:
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

### 3. Satisfy and Assignment for a Subsystem

Now we will take the same command, but submit with a jobspec directly. This is considered an advanced use-case, because it's unlikely that someone would be writing jobspecs directly (but not impossible). We have a simple design here for the jobspec that is detailed in [algorithms](algorithms.md). You can run the example, again with two terminals, as follows:

```bash
# terminal 1 for server
rm -f rainbow.db && make server
```

#### Match Algorithm (default)

```bash
# terminal 2 to register cluster, subsystem, and submit job
make register && make subsystem && go run ./cmd/rainbow/rainbow.go submit --config-path ./docs/examples/scheduler/rainbow-config.yaml --jobspec ./docs/examples/scheduler/jobspec-io.yaml
```

The new portion from the above is seeing that the subsystem "io" is satisfied at some level of resource.

```console
...
🍇️ Satisfy request to Graph 🍇️
 jobspec: {"version":1,"resources":{"ior":{"type":"node","replicas":1,"with":[{"type":"core","count":2,"attributes":{}}],"requires":[{"field":"type","match":"shm","name":"io"}],"attributes":{}}}}
  🎰️ Resources that that need to be satisfied with matcher match
     node:  (slot)  1
       requires
         field: type
         match: shm
         name: io

  🔍️ Exploring cluster keebler deeper with depth first search
      => Searching for resource type core from parent contains->rack
      => Searching for resource type core from parent contains->node
           Found subsystem edge for io with type shm
           Minimum slot needs are satisfied at node for io at shm, returning early.
         slotNeeds are satisfied, returning 1 slots matched
Slots found 1/1 for vertex cluster
  match: ✅️ there are 1 matches with sufficient resources
2024/05/07 19:40:56 📝️ received job app for 1 contender clusters
2024/05/07 19:40:56 📝️ job app is assigned to cluster [keebler]
```

And the work is still assigned to the cluster.

#### Range Algorithm (default)

This algorithm is intended to match a range of versions, either a min, max, or both.
We have an example subsystem JGF intended for spack, complete with packages, compilers, externals, licenses, and anguish.  In one
window, start the server:

```bash
make server
```

In another terminal register the nodes, the subsystem, and then submit the job with the range algorithm;

```bash
# Create your rainbow config
go run cmd/rainbow/rainbow.go config init --cluster-name spack-builder --config-path ./docs/examples/match-algorithms/range/rainbow-config.yaml --match-algorithm range

# Register your nodes
go run cmd/rainbow/rainbow.go register cluster --cluster-name spack-builder --nodes-json ./docs/examples/match-algorithms/range/cluster-nodes.json --config-path ./docs/examples/match-algorithms/range/rainbow-config.yaml --save

# Register the subsystem
go run cmd/rainbow/rainbow.go register subsystem  --subsystem spack --nodes-json ./docs/examples/match-algorithms/range/spack-subsystem.json --config-path ./docs/examples/match-algorithms/range/rainbow-config.yaml

# Submit a job that asked for a valid range
go run ./cmd/rainbow/rainbow.go submit --config-path ./docs/examples/match-algorithms/range/rainbow-config.yaml --jobspec ./docs/examples/match-algorithms/range/jobspec-valid-range.yaml --match-algorithm range
```
For the above job, you'll see it's satisfied:

```console
  match: ✅️ there are 1 matches with sufficient resources
2024/03/30 17:03:35 📝️ received job ior for 1 contender clusters
2024/03/30 17:03:35 📝️ job ior is assigned to cluster spack-builder
```

Try submitting a job that can't be satisfied for the range.

```bash
# Submit a job that asked for a valid range
go run ./cmd/rainbow/rainbow.go submit --config-path ./docs/examples/match-algorithms/range/rainbow-config.yaml --jobspec ./docs/examples/match-algorithms/range/jobspec-invalid-range.yaml --match-algorithm range
```
```console
Slots found 0/1 for vertex cluster
  match: 🎯️ cluster spack-builder does not have sufficient resources and is NOT a match
  match: 😥️ no clusters could satisfy this request. We are sad
```

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




[home](/README.md#rainbow-scheduler)
