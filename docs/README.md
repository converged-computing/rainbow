# Rainbow Scheduler

The rainbow scheduler is a combined scheduler and client to allow for multi-cluster scheduling, meaning submission and management of jobs across environments. It is currently in a prototype state.

## Prototype Design

For earlier designs, see the [design](design.md) pages

### Current Design

> Late February 2024

We next want to add a simple scheduler, meaning that the new user interaction works as follows:

1. The user submits a job or application specification (e.g., run a container with compatibility information, or an application with the same) to the rainbow scheduler.
2. The rainbow scheduler then authenticates the user, and can select a best match from a subset of clusters for which the user has access
  - This requires the user tokens, and eventually something more robust like accounts in a database). - This also requires (finally) a graph in rainbow, making it more of a scheduler 
3. The rainbow scheduler then filters down clusters to those that might match.
  - This requires sending over cluster metadata on the register step
4. The clusters respond with Yes/No and ETA or cost to choose from.
5. The job is assigned a cluster (or rejected). If assigned, the cluster queries for it when ready.
 - Akin to before, the cluster can have its own means to select jobs to run from the set assigned to it.

### Authentication

In more detail, this is what the above means for work:

- Authentication: We need to authenticate the user for *multiple* clusters. We should likely create a token/auth file to do this, that has cluster names and tokens. To start (with testing) the tokens can be the same (shared). Eventually this should be more robust.
- In the server, we have to check all tokens to see if the user has permission. In the future there could be some concept of a cluster group (with one token).


### Registration

The new flow will be as follows:

- At registration, the cluster also sends over metadata about itself (and the nodes it has). This is going to allow for selection for those nodes. But it needs to be some kind of summary information, maybe across a graph? TODO: start with a spec of nodes, maybe Kubernetes, and summarize counts?
- When submitting a job, the user no longer is giving an exact command, but a command + an image with compatibility metadata. The compatibility metadata (somehow) needs to be used to inform the cluster selection.
- At selection, the rainbow schdeuler needs to filter down cluster options, and choose a subset.
 - Level 1: Don't ask, juts choose the top choice and submit
 - Level 2: Ask the cluster for TBA time or cost, choose based on that.
 - Job is added to that queue.

TODO:

- first make config file with multiple secrets for multiple clusters
- then allow specifying to give the config file instead
- then write client function that can read in a graph of nodes (how to generate)?


This first design was a proof of concept that we could submit jobs from a single point to multiple different flux clusters. In that sense, it was mostly a dispatcher (no scheduler) that:

- Exposes an API that can take job requests, where a request is a simple command and resources.
- Clusters can register to it, meaning they are allowed to ask for work.
- Users will submit jobs (from anywhere) to the API, targeting a specific cluster (again, no scheduling here)
- The cluster will run a client that periodically checks for new jobs to run.

This is currently a prototype that demonstrates we can do a basic interaction from multiple places, and obviously will have a lot of room for improvement.
We can run the client alongside any flux instance that has access to this service (and is given some shared secret).


For more details on the design, see [design.md](design.md)

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