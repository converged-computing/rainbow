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
2024/02/11 15:36:19 creating server...
2024/02/11 15:36:19 starting server: server (development) vv0.0.1-default
2024/02/11 15:36:19 server listening: [::]:50051
2024/02/11 15:39:08 received register: type_url:"type.googleapis.com/google.protobuf.StringValue" value:"\n\x07keebler"
```

And then mock a registration:

```bash
make register
```
```console
go run cmd/register/register.go
2024/02/11 15:46:53 creating client (v0.0.1-default)...
2024/02/11 15:46:53 starting client (localhost:50051)...
2024/02/11 15:46:53 registering cluster: keebler
request_id:"0f7a0e7d-c2ed-4a46-9eaa-3d554349244e"
2024/02/11 15:46:53 received response: register success
```

Nothing meaningful is happening yet - I'm just creating a skeleton (and learning about servers / services in Go with grpc more) and am going to add meat to this. My plan is below in [TODO](#TODO).

## Container Images

**Coming soon**

## TODO

- Add an actual database (sqlite) to the server, which should init, create tables for clusters, jobs (ids and cluster assignment)
- When a registration is done, it should check against this database (and add a new cluster or determine already registered)
- Add a secret to validate that, and generation of a cluster-specific token to validate further responses.
- Write the job submission endpoint, which should take a cluster name and command, and return status (success, denied, etc.)

At this point we will have a dumb little database with jobs assigned to clusters. We can then modify the client to add a polling command (intended to be run on a flux instance) that will use the cluster-specific token to say "Do you have any jobs for me?" at some interval. This can run anywhere there is a Flux instance. It can receive the job, and run it. When it receives the job, the job will be deleted from the database, because we don't care anymore.

And that should be a very basic prototype - we can then build this into containers and deploy in different places (and deploy a client separate from a Flux instance) and demonstrate submitting jobs across different places. For the Flux instance logic, we could write the grpc endpoints in Python, but it would be more fun to (finally) make Go bindings for flux core.

## License

HPCIC DevTools is distributed under the terms of the MIT license.
All new contributions must be made under this license.

See [LICENSE](https://github.com/converged-computing/cloud-select/blob/main/LICENSE),
[COPYRIGHT](https://github.com/converged-computing/cloud-select/blob/main/COPYRIGHT), and
[NOTICE](https://github.com/converged-computing/cloud-select/blob/main/NOTICE) for details.

SPDX-License-Identifier: (MIT)

LLNL-CODE- 842614