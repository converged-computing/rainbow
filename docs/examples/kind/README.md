# Kubernetes in Docker (kind)

This example shows using the [rainbow docker images](https://github.com/orgs/converged-computing/packages?repo_name=rainbow) locally via kind, which is Kubernetes in docker. These same manifests (the YAML files) will likely work in production Kubernetes as well.

## Usage

## 1. Create Cluster

This cluster is going to allow us to create ingress.

```bash
kind create cluster --config ./kind-config.yaml
```

## 2. Load Images

This step is optional, but if you want to load your images in before creating Kubernetes objects, you can.
This mostly helps if you have a local image (otherwise it will pull from the remote registry directory)

```bash
kind load docker-image ghcr.io/converged-computing/rainbow-flux:latest
kind load docker-image ghcr.io/converged-computing/rainbow-scheduler:latest
```

## 3. Create Service and Ingress

Let's next create the service. While we don't technically need this (communication happens within the network of pods) we anticipate some case when we will want to interact from outside of that space and thus show you how to set it up.

```bash
kubectl create -f ./service.yaml
kubectl create -f ./ingress.yaml
```

## 4. Create Rainbow Scheduler

Let's create the rainbow scheduler deployment.

```bash
kubectl create -f ./scheduler.yaml
```

And ensure it is running OK:

```bash
kubectl logs scheduler-798ddccf-pxfxx 
```
```console
2024/02/14 20:53:37 creating 🌈️ server...
2024/02/14 20:53:37 ✨️ creating rainbow.db...
2024/02/14 20:53:37    rainbow.db file created
2024/02/14 20:53:37    create cluster table...
2024/02/14 20:53:37    cluster table created
2024/02/14 20:53:37    create jobs table...
2024/02/14 20:53:37    jobs table created
2024/02/14 20:53:37 ⚠️ WARNING: global-token is set, use with caution.
2024/02/14 20:53:37 starting scheduler server: rainbow v0.1.0-draft
2024/02/14 20:53:37 server listening: [::]:8080
```

Importantly, we give the scheduler a predictable hostname. 

```bash
kubectl exec -it scheduler-798ddccf-pxfxx -- cat /etc/hosts
```
```console
# Kubernetes-managed hosts file.
127.0.0.1       localhost
::1     localhost ip6-localhost ip6-loopback
fe00::0 ip6-localnet
fe00::0 ip6-mcastprefix
fe00::1 ip6-allnodes
fe00::2 ip6-allrouters
10.244.0.5      scheduler.rainbow.default.svc.cluster.local     scheduler
```

## 5. Start Clusters

We are going to cheat a bit and create multiple "clusters" (pods) via an indexed job.
The names of the pods (hostnames) will correspond with the names of the clusters. We also
are using a `--global-token` so a shared filesystem is not needed - the scheduler will assign
the same token to all newly registered clusters. This is of course not intended for a production
setup.

```bash
kubectl apply -f ./clusters.yaml
```

Four pods should be running now! You can watch the scheduler logs, and ultimately see logs of a cluster to see what is happening.

```console
👋️ Hello, I'm clusters-0!
📜️ Registering clusters-0...
token: "jellaytime"
secret: "e2342db8-1c0b-40dd-9b42-e49e56480916"
status: REGISTER_SUCCESS

🥳️ All 3 clusters are registered.
status: SUBMIT_SUCCESS

status: SUBMIT_SUCCESS

status: SUBMIT_SUCCESS

Status: REQUEST_JOBS_SUCCESS
Received 2 jobs for inspection!
Accepting 2 jobs...
[5, 1]
Submit job ['echo', 'hello', 'from', 'clusters-1,', 'a', 'new', 'word', 'is', 'perfume']: ƒCYGUnZd
Submit job ['echo', 'hello', 'from', 'clusters-2,', 'a', 'new', 'word', 'is', 'sn']: ƒCYraW3Z
Ran job ƒCYGUnZd hello from clusters-1, a new word is perfume

Ran job ƒCYraW3Z hello from clusters-2, a new word is sn

status: SUBMIT_SUCCESS

status: SUBMIT_SUCCESS

status: SUBMIT_SUCCESS

Status: REQUEST_JOBS_SUCCESS
Received 2 jobs for inspection!
Accepting 2 jobs...
[16, 14]
Submit job ['echo', 'hello', 'from', 'clusters-2,', 'a', 'new', 'word', 'is', 'world']: ƒH7VNJbZ
Submit job ['echo', 'hello', 'from', 'clusters-0,', 'a', 'new', 'word', 'is', 'essex']: ƒH86x1Mq
Ran job ƒH86x1Mq hello from clusters-0, a new word is essex

Ran job ƒH7VNJbZ hello from clusters-2, a new word is world

status: SUBMIT_SUCCESS

status: SUBMIT_SUCCESS

status: SUBMIT_SUCCESS

status: SUBMIT_SUCCESS

status: SUBMIT_SUCCESS

Status: REQUEST_JOBS_SUCCESS
Received 5 jobs for inspection!
Accepting 5 jobs...
[27, 23, 20, 29, 22]
Submit job ['echo', 'hello', 'from', 'clusters-0,', 'a', 'new', 'word', 'is', 'resorts']: ƒMetzhd1
Submit job ['echo', 'hello', 'from', 'clusters-1,', 'a', 'new', 'word', 'is', 'puzzles']: ƒMfY4Pfd
Submit job ['echo', 'hello', 'from', 'clusters-1,', 'a', 'new', 'word', 'is', 'mean']: ƒMg9e6Ru
Submit job ['echo', 'hello', 'from', 'clusters-0,', 'a', 'new', 'word', 'is', 'jamie']: ƒMgnhnUX
Submit job ['echo', 'hello', 'from', 'clusters-1,', 'a', 'new', 'word', 'is', 'final']: ƒMhNoVxT
Ran job ƒMhNoVxT hello from clusters-1, a new word is final

Ran job ƒMg9e6Ru hello from clusters-1, a new word is mean

Ran job ƒMgnhnUX hello from clusters-0, a new word is jamie

Ran job ƒMfY4Pfd hello from clusters-1, a new word is puzzles

Ran job ƒMetzhd1 hello from clusters-0, a new word is resorts

status: SUBMIT_SUCCESS

status: SUBMIT_SUCCESS

status: SUBMIT_SUCCESS

status: SUBMIT_SUCCESS

status: SUBMIT_SUCCESS

Status: REQUEST_JOBS_SUCCESS
Received 2 jobs for inspection!
Accepting 2 jobs...
[33, 35]
Submit job ['echo', 'hello', 'from', 'clusters-1,', 'a', 'new', 'word', 'is', 'goal']: ƒSDXnWB1
Submit job ['echo', 'hello', 'from', 'clusters-1,', 'a', 'new', 'word', 'is', 'during']: ƒSE9NCwH
Ran job ƒSE9NCwH hello from clusters-1, a new word is during

Ran job ƒSDXnWB1 hello from clusters-1, a new word is goal

status: SUBMIT_SUCCESS

status: SUBMIT_SUCCESS

Status: REQUEST_JOBS_SUCCESS
Received 1 jobs for inspection!
Accepting 1 jobs...
[41]
Submit job ['echo', 'hello', 'from', 'clusters-2,', 'a', 'new', 'word', 'is', 'manufacture']: ƒWjQTeHm
Ran job ƒWjQTeHm hello from clusters-2, a new word is manufacture

💤️ Cluster clusters-0 is finished! Shutting down.
```


And that's it! Clean up when you are done:

```bash
kind delete cluster
```