# rainbow

> 🌈️ Where keebler elves and schedulers live, somewhere in the clouds, and with marshmallows 

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

And then mock a registration:

```bash
make register
```
```console
go run cmd/register/register.go
2024/02/11 18:06:59 creating client (v0.0.1-default)...
2024/02/11 18:06:59 🌈️ starting client (localhost:50051)...
2024/02/11 18:06:59 registering cluster: keebler
2024/02/11 18:06:59 status: REGISTER_SUCCESS
2024/02/11 18:06:59  token: 712747b7-b2a9-4bea-b630-056cd64856e6
```

You'll see this from the server:

```console
2024/02/11 18:06:59 📝️ received register: keebler
SELECT count(*) from clusters WHERE name LIKE "keebler": (0)
INSERT into clusters VALUES ("keebler", "712747b7-b2a9-4bea-b630-056cd64856e6"): (1)
```

In the above, we are providing a cluster name (keebler) and it is being registered to the database, and a token and status returned. Note that if we want to submit a job to the "keebler" cluster, from anywhere, we need this token! Let's try that next.

```bash
# Look at help
go run ./cmd/rainbow/rainbow.go submit --help
```
```
usage: rainbow submit [-h|--help] [-s|--secret "<value>"] [-n|--nodes
               <integer>] [-t|--tasks <integer>] [-c|--command "<value>"]
               [--job-name "<value>"] [--host "<value>"] [--cluster-name
               "<value>"]

               Submit a job to a rainbow cluster

Arguments:

  -h  --help          Print help information
  -s  --secret        Registration 'secret'. Default: chocolate-cookies
  -n  --nodes         Number of nodes to request. Default: 1
  -t  --tasks         Number of tasks to request (per node? total?)
  -c  --command       Command to submit. Default: chocolate-cookies
      --job-name      Name for the job (defaults to first command).
      --host          Scheduler server address (host:port). Default:
                      localhost:50051
      --cluster-name  Name of cluster to register. Default: keebler
```
Let's try doing that.

```bash
go run ./cmd/rainbow/rainbow.go submit --secret "712747b7-b2a9-4bea-b630-056cd64856e6" --command hostname
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
Now we need to write the logic for the cluster to poll asking for jobs assigned to it to receive it! I'll work on that next.

## Container Images

**Coming soon**

## TODO

At this point we have a dumb little database with jobs assigned to clusters. We can then modify the client to add a polling command (intended to be run on a flux instance) that will use the cluster-specific token to say "Do you have any jobs for me?" at some interval. This can run anywhere there is a Flux instance. It can receive the job, and run it. When it receives the job, the job will be deleted from the database, because we don't care anymore.

And that should be a very basic prototype - we can then build this into containers and deploy in different places (and deploy a client separate from a Flux instance) and demonstrate submitting jobs across different places. For the Flux instance logic, we could write the grpc endpoints in Python, but it would be more fun to (finally) make Go bindings for flux core.

## License

HPCIC DevTools is distributed under the terms of the MIT license.
All new contributions must be made under this license.

See [LICENSE](https://github.com/converged-computing/cloud-select/blob/main/LICENSE),
[COPYRIGHT](https://github.com/converged-computing/cloud-select/blob/main/COPYRIGHT), and
[NOTICE](https://github.com/converged-computing/cloud-select/blob/main/NOTICE) for details.

SPDX-License-Identifier: (MIT)

LLNL-CODE- 842614