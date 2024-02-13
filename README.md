# rainbow

> 🌈️ Where keebler elves and schedulers live, somewhere in the clouds, and with marshmallows 

[![PyPI version](https://badge.fury.io/py/rainbow-scheduler.svg)](https://badge.fury.io/py/rainbow-scheduler)
![img/rainbow.png](img/rainbow.png)

This is a prototype that will use a Go [gRPC](https://grpc.io/) server/client to demonstrate multi-cluster scheduling. This won't be doing anything intelligent with respect to scheduling (but could) but instead:

- Will expose an API that can take job requests, where a request is a simple command and resources.
- Clusters can register to it, meaning they are allowed to ask for work.
- Users will submit jobs (from anywhere) to the API, targeting a specific cluster (again, no scheduling here)
- The cluster will run a client that periodically checks for new jobs to run.

This will just be a prototype that demonstrates we can do a basic interaction from multiple places, and obviously will have a lot of room for improvement.
We can run the client alongside any flux instance that has access to this service (and is given some shared secret).

## Components

 - The main server (and optionally, a client) are implemented in Go, here
 - Under [python](python) we also have a client that is intended to run from a flux instance, another scheduler, or anywhere really. We haven't implemented the same server in entirety because it's assumed if you plan to run a server, Go is the better choice (and from a container we will provide). That said, the skeleton is there, but unimplemented for the most part.

## Development

### proto

We are using [Protocol Buffers](https://developers.google.com/protocol-buffers/)  "Protobuf" to define the API (how the payloads are shared and the methods for communication between client and server). These are defined in [api/v1/sample.proto](api/v1/sample.proto). 
You can read more about Protobuf [here](https://github.com/golang/protobuf), I first saw / used them with fluence and am still pretty new.

```shell
make proto
```

That will download protoc and needed tools into a local "bin" and then generate the bindings.

## Getting Started

### Setup

Ensure you have your dependencies:

```bash
make tidy
```

In two terminals, start the server in one:

```bash
make server
```
```console
go run cmd/server/server.go
2024/02/12 19:38:58 creating 🌈️ server...
2024/02/12 19:38:58 ✨️ creating rainbow.db...
2024/02/12 19:38:58    rainbow.db file created
2024/02/12 19:38:58    create cluster table...
2024/02/12 19:38:58    cluster table created
2024/02/12 19:38:58    create jobs table...
2024/02/12 19:38:58    jobs table created
2024/02/12 19:38:58 starting scheduler server: rainbow v0.1.0-draft
2024/02/12 19:38:58 server listening: [::]:50051
```

### Register

And then mock a registration:

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

### Submit Job

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

### Request Jobs

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


### Accept Jobs

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


## Container Images

**Coming soon**

## TODO

- endpoint to summarize jobs? Or update the request jobs to return a summary?
- request jobs should accept way to filter or specify criteria for request
- receiving endpoint to accept (meaning just deleting from the database)

Next steps: Python bindings so we can run the client in a flux instance and:

- Run in poll, at some increment
- "Do you have jobs for me?"
- "Yes I'll accept and run N" (delete from database)

The above should be a very basic prototype - we can then build this into containers and deploy in different places (and deploy a client separate from a Flux instance) and demonstrate submitting jobs across different places. For the Flux instance logic, we could write the grpc endpoints in Python, but it would be more fun to (finally) make Go bindings for flux core.

## License

HPCIC DevTools is distributed under the terms of the MIT license.
All new contributions must be made under this license.

See [LICENSE](https://github.com/converged-computing/cloud-select/blob/main/LICENSE),
[COPYRIGHT](https://github.com/converged-computing/cloud-select/blob/main/COPYRIGHT), and
[NOTICE](https://github.com/converged-computing/cloud-select/blob/main/NOTICE) for details.

SPDX-License-Identifier: (MIT)

LLNL-CODE- 842614
