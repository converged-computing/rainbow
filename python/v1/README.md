# rainbow (python)

> 🌈️ Where keebler elves and schedulers live, somewhere in the clouds, and with marshmallows

![https://github.com/converged-computing/rainbow/raw/main/docs/img/rainbow.png](https://github.com/converged-computing/rainbow/raw/main/docs/img/rainbow.png)

This is the rainbow scheduler prototype, specifically Python bindings for a gRPC client. To learn more about rainbow, visit [https://github.com/converged-computing/rainbow](https://github.com/converged-computing/rainbow).

## Example

Assuming that you can run the server with Go, let's first do that (e.g., from the root of the repository linked above, and soon we will provide a container):

### Register

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

And then let's do a registration, but this time from the Python bindings (client) here! We will use the core bindings in [rainbow/client.py](rainbow/client.py) but run a custom command from [examples](examples). Assuming you've installed everything into a venv:

```bash
python -m venv env
source env/bin/activate
pip install -e .
```

The command below will register and save the secret to a new configuration file.
Note that if you provide an existing one, it will use or update it.

```bash
python ./examples/flux/register.py keebler --config-path ./rainbow-config.yaml
```
```console
Saving rainbow config to ./rainbow-config.yaml
🤫️ The token you will need to submit jobs to this cluster is rainbow
🔐️ The secret you will need to accept jobs is 649598a9-e77b-4aa3-ab46-bfbbc5e2d606
```
Try running it again - you can't register a cluster twice. But of course other cluster names you can register. A "cluster" can actually be a cluster, or a flux instance, or any entity that can accept jobs. The script also accepts arguments (see `register.py --help`)

```console
python ./examples/flux/register.py --help

🌈️ Rainbow scheduler register

options:
  -h, --help            show this help message and exit
  --cluster CLUSTER     cluster name to register
  --host HOST           host of rainbow cluster
  --secret SECRET       Rainbow cluster registration secret
  --config-path CONFIG_PATH
                        Path to rainbow configuration file to write or use
  --cluster-nodes CLUSTER_NODES
                        Nodes to provide for registration
```

### Register Subsystem

Let's now register the subsystem. Akin to register, this has the path to the subsystem nodes set as a default,
and the name `--subsystem` set to "io." This assumes you've registered your cluster and have the cluster secret
in your ./rainbow-config.yaml

```bash
python ./examples/flux/register-subsystem.py keebler --config-path ./rainbow-config.yaml
```
```console
status: REGISTER_SUCCESS
```

In the server window you'll see the subsystem added:

```console
...
2024/03/09 14:21:50 📝️ received subsystem register: keebler
2024/03/09 14:21:50 Preparing to load 6 nodes and 30 edges
2024/03/09 14:21:50 We have made an in memory graph (subsystem io) with 7 vertices, with 15 connections to the dominant!
{
 "keebler": {
  "Name": "keebler",
  "Counts": {
   "io": 1,
   "mtl1unit": 1,
   "mtl2unit": 1,
   "mtl3unit": 1,
   "nvme": 1,
   "shm": 1
  }
 }
}
```

### Submit Job (Simple)

Now let's submit a job to our faux cluster. We need to provide the token we received above. Remember that this is a two stage process:

1. Query the graph database for one or more cluster matches.
2. Send that request to rainbow.

The client handles both, so you (as the user) only are exposed to the single submit. We will be providing basic arguments for
the job, but note you can provide other arguments too:

```console
python ./examples/flux/submit-job.py --help

🌈️ Rainbow scheduler submit

positional arguments:
  command               Command to submit

options:
  -h, --help            show this help message and exit
  --config-path CONFIG_PATH
                        config path with cluster names
  --host HOST           host of rainbow cluster
  --token TOKEN         Cluster token for permission to submit jobs
  --nodes NODES         Nodes for job (defaults to 1)
```

And then submit! Remember that you need to have registered first. Note that we need to provide our cluster config path.

```console
$ python examples/flux/submit-job.py --config-path ./rainbow-config.yaml --nodes 1 echo hello world
```bash
```console
{
    "version": 1,
    "resources": [
        {
            "type": "node",
            "count": 1,
            "with": [
                {
                    "type": "slot",
                    "count": 1,
                    "label": "echo",
                    "with": [
                        {
                            "type": "core",
                            "count": 1
                        }
                    ]
                }
            ]
        }
    ],
    "tasks": [
        {
            "command": [
                "echo",
                "hello",
                "world"
            ],
            "slot": "echo",
            "count": {
                "per_slot": 1
            }
        }
    ],
    "attributes": {}
}
clusters: "keebler"
status: RESULT_TYPE_SUCCESS

status: SUBMIT_SUCCESS
```

### Submit Jobspec

We can also submit a jobspec directly, which is an advanced use case. It works predominantly the same, except we load in the Jobspec from
the yaml directly.

```console
python examples/flux/submit-jobspec.py --config-path ./rainbow-config.yaml ../../docs/examples/scheduler/jobspec-io.yaml

🌈️ Rainbow scheduler submit

positional arguments:
  jobspec               Jobspec path to submit

options:
  -h, --help            show this help message and exit
  --config-path CONFIG_PATH
                        config path with cluster metadata
```

It largely looks the same - I'll cut most of it out. It's just a different entry point for the job definition.

```console
clusters: "keebler"
status: RESULT_TYPE_SUCCESS

status: SUBMIT_SUCCESS
```

### Receive Jobs

After we submit jobs, rainbow assigns them to a cluster. For this dummy example we are assigning to the same cluster (keebler) so we can also use our host "keebler" to receive the job. Here is what that looks like.

```console
python ./examples/flux/receive-jobs.py --help

🌈️ Rainbow scheduler receive jobs

options:
  -h, --help            show this help message and exit
  --max-jobs MAX_JOBS   Maximum jobs to request (unset defaults to all)
  --config-path CONFIG_PATH
                        config path with cluster metadata
```

And then request and accept jobs:

```console
 python examples/flux/receive-jobs.py --config-path ./rainbow-config.yaml
Status: REQUEST_JOBS_SUCCESS
Received 1 jobs to accept...
```

If this were running in Flux, we would be able to run it, and the response above has told rainbow that you've accepted it (and rainbow deletes the record of it).


## License

HPCIC DevTools is distributed under the terms of the MIT license.
All new contributions must be made under this license.

See [LICENSE](https://github.com/converged-computing/cloud-select/blob/main/LICENSE),
[COPYRIGHT](https://github.com/converged-computing/cloud-select/blob/main/COPYRIGHT), and
[NOTICE](https://github.com/converged-computing/cloud-select/blob/main/NOTICE) for details.

SPDX-License-Identifier: (MIT)

LLNL-CODE- 842614
