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

## License

HPCIC DevTools is distributed under the terms of the MIT license.
All new contributions must be made under this license.

See [LICENSE](https://github.com/converged-computing/cloud-select/blob/main/LICENSE),
[COPYRIGHT](https://github.com/converged-computing/cloud-select/blob/main/COPYRIGHT), and
[NOTICE](https://github.com/converged-computing/cloud-select/blob/main/NOTICE) for details.

SPDX-License-Identifier: (MIT)

LLNL-CODE- 842614
