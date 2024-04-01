#!/bin/bash

go run cmd/rainbow/rainbow.go config init --cluster-name spack-builder --config-path ./docs/examples/match-algorithms/range/rainbow-config.yaml --match-algorithm range

# Register your nodes
go run cmd/rainbow/rainbow.go register cluster --cluster-name spack-builder --nodes-json ./docs/examples/match-algorithms/range/cluster-nodes.json --config-path ./docs/examples/match-algorithms/range/rainbow-config.yaml --save

# Register the subsystem
go run cmd/rainbow/rainbow.go register subsystem  --subsystem spack --nodes-json ./docs/examples/match-algorithms/range/spack-subsystem.json --config-path ./docs/examples/match-algorithms/range/rainbow-config.yaml

# Submit a job that asked for a valid range
go run ./cmd/rainbow/rainbow.go submit --config-path ./docs/examples/match-algorithms/range/rainbow-config.yaml --jobspec ./docs/examples/match-algorithms/range/jobspec-valid-range.yaml --match-algorithm range
