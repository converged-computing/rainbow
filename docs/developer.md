# Developer Guide

This is a short guide to help with development. It's recommended to develop rainbow locally with Go (I use version 1.20.14) but a `.devcontainer` environment is provided if you don't have that.

### Documentation

The main documentation for the repository is in the [docs](https://github.com/converged-computing/rainbow/tree/main/docs) directory, and the interface itself is static and generated from the markdown with
javascript. You can edit the markdown files there to update the documentation.

### Protobuf

We are using [Protocol Buffers](https://developers.google.com/protocol-buffers/)  "Protobuf" to define the API (how the payloads are shared and the methods for communication between client and server). These are defined in [api/v1/sample.proto](api/v1/sample.proto).
You can read more about Protobuf [here](https://github.com/golang/protobuf), I first saw / used them with fluence and am still pretty new.

```shell
make proto
```

That will download protoc and needed tools into a local "bin" and then generate the bindings.

### Build Local

You can build the binaries:

```console
$ make build
mkdir -p /home/vanessa/Desktop/Code/rainbow/bin
GO111MODULE="on" go build -o /home/vanessa/Desktop/Code/rainbow/bin/rainbow cmd/rainbow/rainbow.go
GO111MODULE="on" go build -o /home/vanessa/Desktop/Code/rainbow/bin/rainbow-scheduler cmd/server/server.go
```

Note that the `rainbow-scheduler` starts the server, and `rainbow` is the set of client commands.

```console
$ ls bin/
protoc-gen-go  protoc-gen-go-grpc  rainbow  rainbow-scheduler
```

They are placed in the local bin, as shown above.

### Python

To build Python GRPC, ensure you have the grpc-tools installed:

```bash
pip install grpcio-tools
```

Then:

```bash
make python
```

and cd into [python/v1](python/v1) and follow the README instructions there.


## Container Images

We provide make commands to build:

- **ghcr.io/converged-computing/rainbow-scheduler**: the scheduler (the `rainbow` client and `rainbow-scheduler` binaries in an ubuntu base, intended to be run as the scheduler image)
- **ghcr.io/converged-computing/rainbow-flux**: the client (includes flux) for interacting with a scheduler.

Both images above have both binaries, it's just that the second has flux added. We can add more schedulers or other entities that can
accept jobs as needed. You can build in any of the following ways:

```bash
# both images, default registry
make docker

# scheduler
make docker-ubuntu

# client with flux
make docker-flux

# customize the registry for any command above
REGISTRY=vanessa make docker
```

Further instructions will be added for running these containers in the next round of work - likely we will have a basic kind setup that demonstrates the orchestration.


### VSCode Developer Environment

We have a [Dev Container](https://code.visualstudio.com/docs/remote/containers)
provided via the assets in [.devcontainer](https://code.visualstudio.com/docs/remote/containers#_create-a-devcontainerjson-file).

You can follow the [tutorial](https://code.visualstudio.com/docs/remote/containers-tutorial) where you'll basically
need to:

1. Install Docker, or compatible engine
2. Install the [Development Containers](vscode:extension/ms-vscode-remote.remote-containers) extension

Then you can go to the command palette (View -> Command Palette) and select `Dev Containers: Open Workspace in Container.`
This will build a development environment with Go that is ready to go and use rainbow.

You are free to change the base image and rebuild if you need to test on another operating system!
When your container is built, when you open `Terminal -> New Terminal`, and you'll be in the container.

**Important** the development container assumes you are on a system with uid 1000 and gid 1000. If this isn't the case,
edit the `.devcontainer/Dockerfile` to be your user and group id. This will ensure changes written inside the container
are owned by your user. It's recommended that you commit on your system (not inside the container) because if you
need to sign your commits, the container doesn't have access and won't be able to. If you find that you accidentally
muck up permissions and need to fix, you can run this from your terminal outside of VSCode:

```bash
$ sudo chown -R $USER .git/
# and then commit
```

[home](/README.md#rainbow-scheduler)
