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

```bash
make register
```
```console
go run cmd/rainbow/rainbow.go register
2024/02/12 22:17:43 🌈️ starting client (localhost:50051)...
2024/02/12 22:17:43 registering cluster: keebler
2024/02/12 22:17:43 status: REGISTER_SUCCESS
2024/02/12 22:17:43 secret: 54c4568a-14f2-465f-aa1e-5e6e0e3efd33
2024/02/12 22:17:43  token: 67e0f258-96c3-4d88-8253-287a95653138
```

If you ran this using the rainbow client you would do:

```bash
```

In the above:

- `token` is what is given to clients to submit jobs
- `secret` is a secret just for your cluster / instance / place you can receive jobs to receive them!

You'll see this from the server:

```console
2024/02/12 22:17:43 📝️ received register: keebler
2024/02/12 22:17:43 SELECT count(*) from clusters WHERE name = 'keebler': (0)
2024/02/12 22:17:43 INSERT into clusters (name, token, secret) VALUES ("keebler", "67e0f258-96c3-4d88-8253-287a95653138", "54c4568a-14f2-465f-aa1e-5e6e0e3efd33"): (1)
```

In the above, we are providing a cluster name (keebler) and it is being registered to the database, and a token, secret and status returned. Note that if we want to submit a job to the "keebler" cluster, from anywhere, we need this token! Let's try that next.

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
