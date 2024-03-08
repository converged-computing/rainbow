# rainbow

> 🌈️ Where keebler elves and schedulers live, somewhere in the clouds, and with marshmallows

[![PyPI version](https://badge.fury.io/py/rainbow-scheduler.svg)](https://badge.fury.io/py/rainbow-scheduler)
![docs/img/rainbow.png](docs/img/rainbow.png)

This is a prototype that will use a Go [gRPC](https://grpc.io/) server/client to demonstrate multi-cluster scheduling.
For more information:

 - ⭐️ [Documentation](https://converged-computing.github.io/rainbow) ⭐️


## TODO

- submit
 - need strategy to add summary state data to top of graph (to provide to selection algorithms)
 - need more selection algorithms!
- receive
 - implement python examples next, then prototype alongside flux to submit
 - after that, think about adding subsystem register (and rest of points below)
- clusters
 - implement function to add a subsystem to an existing cluster (e.g., add I/O)
- subsystems
  - do we need bidirectional links for the memory graph?
  - a satisfies request will need to have a representation of subsystems. E.g., what are we asking of each?
    - right now we assume a node resources request going to the dominant subsystem
  - we will want a function to add a new subsystem, right now we have one dominant for nodes
  - make also a function to delete subsystems
  - subsystems should allow for multiple (with keys) and references across to dominant subsystem
- is there a way to unify into one graph?
- ephemeral case - actual nodes don't exist, but instead rules for requests and limits. Need to develop this and means to authenticate to use it.


## License

HPCIC DevTools is distributed under the terms of the MIT license.
All new contributions must be made under this license.

See [LICENSE](https://github.com/converged-computing/rainbow/blob/main/LICENSE),
[COPYRIGHT](https://github.com/converged-computing/rainbow/blob/main/COPYRIGHT), and
[NOTICE](https://github.com/converged-computing/rainbow/blob/main/NOTICE) for details.

SPDX-License-Identifier: (MIT)

LLNL-CODE- 842614
