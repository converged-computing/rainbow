# Advanced Commands

This guide assumes that you have some familarity with rainbow, and want to do advanced things like change algorithms. Instead of walking through by commands, we will walk through by way of specific goals or examples.
These examples won't abstract away commands with the Makefile, but will show them in entirety for you to better understand. If you aren't ready for this guide, you should read about [commands](commands.md) first.

## Constraint Selection Algorithm

Let's walk through an example of submitting a job that uses a constraint selection algorithm.  More specifically, instead of a random selection of the contender clusters at the end (the default) we have a set of rules that are set based on options for the algorithm, the JobSpec, and cluster state. You can read about [algorithms](algorithms.md) to understand more here. Let's first start rainbow, and give a custom configuration file that uses the selection->constraint algorithm.

```bash
# Start the server (terminal 1)
rm rainbow.db || true
go run cmd/server/server.go --loglevel 6 --global-token rainbow --config ./docs/examples/scheduler/rainbow-selection-config.yaml

# Register a cluster (terminal 2)
go run cmd/rainbow/rainbow.go register cluster --cluster-name keebler --nodes-json ./docs/examples/scheduler/cluster-nodes.json --config-path ./docs/examples/scheduler/rainbow-config.yaml --save

# Update cluster state
go run cmd/rainbow/rainbow.go update state --state-file ./docs/examples/scheduler/cluster-state.json --config-path ./docs/examples/scheduler/rainbow-config.yaml

# Try submitting a job using this algorithm!
go run ./cmd/rainbow/rainbow.go submit --config-path ./docs/examples/scheduler/rainbow-selection-config.yaml --jobspec ./docs/examples/scheduler/jobspec-constraint.yaml
```

The above demonstrates using a more advanced selection algorithm.
Note that this cluster state requires further discussion and thinking about where and how to accommodate it - it currently uses the old design with attributes on the level of the Jobspec, and while this works, we likely want to be using the attributes on the level of schedule-able unit.

[home](/README.md#rainbow-scheduler)
