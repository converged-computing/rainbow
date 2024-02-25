# Rainbow Scheduler

The rainbow scheduler is a combined scheduler and client to allow for multi-cluster scheduling, meaning submission and management of jobs across environments. It is currently in a prototype state.

## Designs

For designs, see the [design](design.md) pages

## Components

 - The main server (and optionally, a client) are implemented in Go, here
 - Under [python](https://github.com/converged-computing/rainbow/tree/main/python/v1) we also have a client that is intended to run from a flux instance, another scheduler, or anywhere really. We haven't implemented the same server in entirety because it's assumed if you plan to run a server, Go is the better choice (and from a container we will provide). That said, the skeleton is there, but unimplemented for the most part.
 - See [examples](https://github.com/converged-computing/rainbow/tree/main/docs/examples) for basic documentation and ways to deploy (containers and Kubernetes with kind, for example).

## Setup

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

Note that we also provide [containers](https://github.com/orgs/converged-computing/packages?repo_name=rainbow) for running the scheduler, or a client with Flux. For more advanced examples, continue reading commands below or check out our [examples](https://github.com/converged-computing/rainbow/tree/main/docs/examples). 

## Commands

Read more about the commands shown above [here](commands.md#commands).

## Development

Read our [developer guide](#developer.md)