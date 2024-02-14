# rainbow (python)

> 🌈️ Where keebler elves and schedulers live, somewhere in the clouds, and with marshmallows 

![https://github.com/converged-computing/rainbow/raw/main/img/rainbow.png](https://github.com/converged-computing/rainbow/raw/main/img/rainbow.png)

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

```bash
python ./examples/flux/register.py keebler
```
```console
$ python ./examples/flux/register.py keebler
token: "e2d48b92-7801-4ad3-8b4d-87e30c1441e0"
secret: "39ef6c99-ac20-40de-92df-d16e9fe651e1"
status: REGISTER_SUCCESS

🤫️ The token you will need to submit jobs to this cluster is e2d48b92-7801-4ad3-8b4d-87e30c1441e0
🔐️ The secret you will need to accept jobs is 39ef6c99-ac20-40de-92df-d16e9fe651e1
```

Try running it again - you can't register a cluster twice.

```console
python ./examples/flux/register.py keebler
status: REGISTER_EXISTS

The cluster keebler alreadey exists.
```

But of course other cluster names you can register. A "cluster" can actually be a cluster, or a flux instance, or any entity that can accept jobs. The script also accepts arguments (see `register.py --help`)

```console
python ./examples/flux/register.py --help

🌈️ Rainbow scheduler register

options:
  -h, --help         show this help message and exit
  --cluster CLUSTER  cluster name to register
  --host HOST        host of rainbow cluster
  --secret SECRET    Rainbow cluster registration secret
```

### Submit Job

Now let's submit a job to our faux cluster. We need to provide the token we received above. Note you can provide other arguments too:

```console
python ./examples/flux/submit-job.py --help

🌈️ Rainbow scheduler register

positional arguments:
  command            Command to submit

options:
  -h, --help         show this help message and exit
  --cluster CLUSTER  cluster name to register
  --host HOST        host of rainbow cluster
  --token TOKEN      Cluster token for permission to submit jobs
  --nodes NODES      Nodes for job (defaults to 1)
```

And then submit!

```console
$ python examples/flux/submit-job.py --token $token --cluster keebler echo hello world
⭐️ Submitting job: echo hello world
status: SUBMIT_SUCCESS
```

### Poll and Accept

Given the above (we have submit jobs to the keebler cluster, not necessarily from it) we would then (from the keebler cluster)
want to receive them, and accept some number to run. Let's mock that next. This is going to include two steps:

 - request jobs: a request from the keebler cluster to list jobs (requires the secret)
 - accept jobs: given the list of jobs requested, tell the rainbow scheduler you accept some subset to run.


```console
python ./examples/flux/poll-jobs.py --help

🌈️ Rainbow scheduler poll (request jobs) and accept

options:
  -h, --help           show this help message and exit
  --cluster CLUSTER    cluster name to register
  --host HOST          host of rainbow cluster
  --max-jobs MAX_JOBS  Maximum jobs to request (unset defaults to all)
  --secret SECRET      Cluster secret to access job queue
  --nodes NODES        Nodes for job (defaults to 1)
  --accept ACCEPT      Number of jobs to accept
```

And then request (poll) for jobs. This does not accept any, it's akin to just asking for a listing, and up to a maximum
number. We will eventually want to add logic to better filter or query for what we _can_ accept.

```console
$ python examples/flux/poll-jobs.py --secret $secret --cluster keebler echo hello world
Status: REQUEST_JOBS_SUCCESS
Received 3 jobs for inspection!
{"id":6,"cluster":"blueberry","name":"","nodes":1,"tasks":0,"command":"echo hello world"}
{"id":4,"cluster":"blueberry","name":"","nodes":1,"tasks":0,"command":"echo hello world"}
{"id":5,"cluster":"blueberry","name":"","nodes":1,"tasks":0,"command":"echo hello world"}
```

Now let's try accepting! 

```console
$ python examples/flux/poll-jobs.py --secret $secret --cluster keebler echo hello world
Status: REQUEST_JOBS_SUCCESS
Received 3 jobs for inspection!
{"id":4,"cluster":"blueberry","name":"","nodes":1,"tasks":0,"command":"echo hello world"}
{"id":6,"cluster":"blueberry","name":"","nodes":1,"tasks":0,"command":"echo hello world"}
{"id":5,"cluster":"blueberry","name":"","nodes":1,"tasks":0,"command":"echo hello world"}
Accepting 1 jobs...
[4]
```

If you were to ask again (for all jobs) you'd only see two because you took one off of the dispatcher,
and it's now owned by your keebler cluster. 
And that's it! Next we can build the above into a container with an actual flux instance and actually run
the jobs we accept.

## License

HPCIC DevTools is distributed under the terms of the MIT license.
All new contributions must be made under this license.

See [LICENSE](https://github.com/converged-computing/cloud-select/blob/main/LICENSE),
[COPYRIGHT](https://github.com/converged-computing/cloud-select/blob/main/COPYRIGHT), and
[NOTICE](https://github.com/converged-computing/cloud-select/blob/main/NOTICE) for details.

SPDX-License-Identifier: (MIT)

LLNL-CODE- 842614