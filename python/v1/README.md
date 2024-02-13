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
token: "956580b8-7339-40aa-84c2-489539bbdc16"
status: REGISTER_SUCCESS

The token you will need to submit jobs to this cluster is 956580b8-7339-40aa-84c2-489539bbdc16
```

Try running it again - you can't register a cluster twice.

```console
python ./examples/flux/register.py keebler
status: REGISTER_EXISTS

The cluster keebler alreadey exists.
```

But of course other cluster names you can register. A "cluster" can actually be a cluster, or a flux instance, or any entity that can accept jobs.

### Submit Job

Now let's submit a job to our faux cluster. We need to provide the token we received above.

```console
$ python examples/flux/submit-job.py 956580b8-7339-40aa-84c2-489539bbdc16
status: SUBMIT_SUCCESS
```

Nice! We will next be writing a receiving endpoint that can poll the server at some increment to ask for jobs, and then accept some number. TBA!


## License

HPCIC DevTools is distributed under the terms of the MIT license.
All new contributions must be made under this license.

See [LICENSE](https://github.com/converged-computing/cloud-select/blob/main/LICENSE),
[COPYRIGHT](https://github.com/converged-computing/cloud-select/blob/main/COPYRIGHT), and
[NOTICE](https://github.com/converged-computing/cloud-select/blob/main/NOTICE) for details.

SPDX-License-Identifier: (MIT)

LLNL-CODE- 842614