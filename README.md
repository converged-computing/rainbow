# rainbow

> 🌈️ Where keebler elves and schedulers live, somewhere in the clouds, and with marshmallows 

[![PyPI version](https://badge.fury.io/py/rainbow-scheduler.svg)](https://badge.fury.io/py/rainbow-scheduler)
![docs/img/rainbow.png](docs/img/rainbow.png)

This is a prototype that will use a Go [gRPC](https://grpc.io/) server/client to demonstrate multi-cluster scheduling. 
For more information:

 - ⭐️ [Documentation](https://converged-computing.github.io/rainbow) ⭐️


## TODO


- (advanced) request jobs should accept way to filter or specify criteria for request

Next steps: Containers and example (kind?) to run in a flux instance and:

- Run in poll, at some increment
- "Do you have jobs for me?"
- "Yes I'll accept and run N" (delete from database)

The above should be a very basic prototype - we can then build this into containers and deploy in different places (and deploy a client separate from a Flux instance) and demonstrate submitting jobs across different places. For the Flux instance logic, we could write the grpc endpoints in Python, but it would be more fun to (finally) make Go bindings for flux core.

### Design Thinking for Scheduler

- Authentication: We need to authenticate the user for *multiple* clusters. We should likely create a token/auth file to do this, that has cluster names and tokens. To start (with testing) the tokens can be the same (shared).
  - In the server, we have to check all tokens to see if the user has permission. In the future there could be some concept of a cluster group (with one token).

The new flow will be as follows:

- At registration, the cluster also sends over metadata about itself (and the nodes it has). This is going to allow for selection for those nodes. But it needs to be some kind of summary information, maybe across a graph? TODO: start with a spec of nodes, maybe Kubernetes, and summarize counts?
- When submitting a job, the user no longer is giving an exact command, but a command + an image with compatibility metadata. The compatibility metadata (somehow) needs to be used to inform the cluster selection.
- At selection, the rainbow schdeuler needs to filter down cluster options, and choose a subset.
 - Level 1: Don't ask, juts choose the top choice and submit
 - Level 2: Ask the cluster for TBA time or cost, choose based on that.
 - Job is added to that queue.

TODO:

- how should we manage auth for multiple clusters (user accounts gets complicated)
- need to go back to compspec-go first and write extractor for nodes, and ability to summarize in directory.

## License

HPCIC DevTools is distributed under the terms of the MIT license.
All new contributions must be made under this license.

See [LICENSE](https://github.com/converged-computing/rainbow/blob/main/LICENSE),
[COPYRIGHT](https://github.com/converged-computing/rainbow/blob/main/COPYRIGHT), and
[NOTICE](https://github.com/converged-computing/rainbow/blob/main/NOTICE) for details.

SPDX-License-Identifier: (MIT)

LLNL-CODE- 842614
