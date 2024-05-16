# Graph Backends

The following graph backends currently supported (in some capacity).

## Memory

The memory backend is provided natively by Rainbow, meaning that rainbow runs the graph service and it hits the same server. This is different than, for example, a distributed graph database that might
have grpc that are interacted with by a backend but does not live within rainbow itself.  The memory backend implements a custom depth first search algorithm that is described in [algorithms](algorithms.md).
It is the default backend if you don't specify a custom one.

## Memgraph

This backend uses [Memgraph](https://github.com/memgraph/memgraph) and we are going to deploy with [docker](https://memgraph.com/docs/getting-started/install-memgraph/docker) and then use the [Go SDK](https://memgraph.com/docs/client-libraries/go) in Rainbow. We are testing this because:

- The GitHub stats are reasonably good
- It has support for multi-tenancy and authentication / authorization
- A Go SDK
- Can support machine learning models
- Can also support custom query definitions in different languages (e.g., Python and C++)

Here is how to run the docker image that has algorithms "MAGE" and ensure we set a secret and custom username.

```bash
docker run -d --rm -p 7444:7444 -p 7687:7687 --name memgraph memgraph/memgraph-mage --memory-limit=500 --log-level=TRACE MGCONSOLE="--username rainbow --password chocolate-cookies"
```

Remove the `-d` to see logs live (can be easier for debugging, but takes up a terminal).
If you need to get to the console:

```bash
docker exec -it memgraph mgconsole
```

We also have a docker compose setup that will add the web interface.

```bash
cd ./docs/examples/memgraph
docker compose up -d
docker compose ps
```

Otherwise we communicate with the Go client.


### Register

Let's try a register. Start rainbow, targeting the memgraph config:

```bash
go run cmd/server/server.go --loglevel 6 --global-token rainbow --config ./docs/examples/memgraph/rainbow-config.yaml
```

Then register.

```bash
go run cmd/rainbow/rainbow.go register cluster --cluster-name keebler --nodes-json ./docs/examples/scheduler/cluster-nodes.json --config-path ./docs/examples/memgraph/rainbow-config.yaml --save
```

Then register a subsystem:

```bash
go run cmd/rainbow/rainbow.go register subsystem --subsystem io --nodes-json ./docs/examples/scheduler/cluster-io-subsystem.json --config-path ./docs/examples/memgraph/rainbow-config.yaml
```

Now let's try submitting a jobspec to match.

```bash
go run ./cmd/rainbow/rainbow.go submit --config-path ./docs/examples/memgraph/rainbow-config.yaml --nodes 2 --tasks 2 --command "echo hello world"
```

Note that you can open the interface to [localhost:3000](localhost:3000) for an interactive experience!
To get help with queries, see [here](https://memgraph.com/docs/querying/clauses).
When you are done:

```bash
# To stop containers (and bring up later)
docker compose stop

# And if you want to remove them too
docker compose rm
```

Note that this backend currently supports match algorithms for range and equality, and these are early in development and need further testing.


[home](/README.md#rainbow-scheduler)
