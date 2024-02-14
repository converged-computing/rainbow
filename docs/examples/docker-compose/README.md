# Docker Compose

This example shows using the [rainbow docker images](https://github.com/orgs/converged-computing/packages?repo_name=rainbow) locally via docker-compose.

## Usage

### 0. TLDR

Don't like reading instructions? Just do this and watch.

```bash
# clean up previous runs
rm -rf ./data/*.txt

# Start the scheduler and three clusters
docker compose -f docker-compose-demo.yaml up -d

# In another terminal, to watch the scheduler
docker compose logs scheduler -f

# and then start the last peer to ensure everything is kicked off
docker exec -it cluster-red flux start python3 /data/run-demo.py --peer cluster-red --peer cluster-blue --peer cluster-yellow
```

That's it! Read more below to see what is actually happening.


### 1. Start Containers

Bring up the cluster.

```bash
docker compose up -d
```
```console
[+] Running 4/4
 ✔ Container scheduler       Started                                                0.5s 
 ✔ Container cluster-blue    Started                                                1.0s 
 ✔ Container cluster-yellow  Started                                                0.8s 
 ✔ Container cluster-red     Started                                                1.0s
```

You should have four containers running - the central scheduler, and three "clusters."

```bash
docker compose ps
```
```console
NAME             IMAGE                                                  COMMAND                  SERVICE          CREATED          STATUS          PORTS
cluster-blue     ghcr.io/converged-computing/rainbow-flux:latest        "rainbow-scheduler t…"   cluster-blue     53 seconds ago   Up 51 seconds   80/tcp, 443/tcp, 8080/tcp
cluster-red      ghcr.io/converged-computing/rainbow-flux:latest        "rainbow-scheduler t…"   cluster-red      53 seconds ago   Up 51 seconds   80/tcp, 443/tcp, 8080/tcp
cluster-yellow   ghcr.io/converged-computing/rainbow-flux:latest        "rainbow-scheduler t…"   cluster-yellow   53 seconds ago   Up 51 seconds   80/tcp, 443/tcp, 8080/tcp
scheduler        ghcr.io/converged-computing/rainbow-scheduler:latest   "rainbow-scheduler r…"   scheduler        53 seconds ago   Up 52 seconds   443/tcp, 0.0.0.0:80->80/tcp, :::80->80/tcp, 8080/tcp
```

Next, take a look at the logs for the scheduler. It should be running on the container port 80 without issue.

```bash
$ docker compose logs scheduler
```
```console
scheduler  | 2024/02/14 05:46:34 creating 🌈️ server...
scheduler  | 2024/02/14 05:46:34 ✨️ creating rainbow.db...
scheduler  | 2024/02/14 05:46:34    rainbow.db file created
scheduler  | 2024/02/14 05:46:34    create cluster table...
scheduler  | 2024/02/14 05:46:34    cluster table created
scheduler  | 2024/02/14 05:46:34    create jobs table...
scheduler  | 2024/02/14 05:46:34    jobs table created
scheduler  | 2024/02/14 05:46:34 starting scheduler server: rainbow v0.1.0-draft
scheduler  | 2024/02/14 05:46:34 server listening: [::]:80
```
The database is currently being written inside the container to a local file (sqlite3)
and will be deleted with the container. You could either bind to the host, or we could add a
database service (another container). Likely we will eventually want the latter.

### 2. Register Clusters

Let's first interactively show you how to register a cluster, with the command line tool (Go) and Python.
Shell in:

```bash
docker exec -it cluster-red bash

# These two commands are the same (but you can only run / register once)
rainbow register --host scheduler:8080 --secret peanutbuttajellay --cluster-name cluster-red
python3 /code/python/v1/examples/flux/register.py --host scheduler:8080 --secret peanutbuttajellay --cluster cluster-red
```
```console
2024/02/14 06:03:33 🌈️ starting client (scheduler:8080)...
2024/02/14 06:03:33 registering cluster: red
2024/02/14 06:03:33 status: REGISTER_SUCCESS
2024/02/14 06:03:33 secret: be6aa08a-c86d-4732-9fee-71c3027c2b18
2024/02/14 06:03:33  token: 94010eeb-a19a-4836-a119-50e89b63dca1
```

I am going to exit, destroy it, and be lazy and register each cluster, and from the outside.

```bash
exit
docker compose stop
docker compose rm
```

Bring it up, and watch the scheduler logs from another terminal:

```bash
docker compose up -d
docker compose logs scheduler -f
```

Now let's register all our clusters at once!

```bash
for color in red blue yellow
  do
    docker exec -it cluster-${color} rainbow register --host scheduler:8080 --secret peanutbuttajellay --cluster-name cluster-${color}
done
```
```console
2024/02/14 06:09:39 🌈️ starting client (scheduler:8080)...
2024/02/14 06:09:39 registering cluster: cluster-red
2024/02/14 06:09:39 status: REGISTER_SUCCESS
2024/02/14 06:09:39 secret: d36f6a5b-5354-4a24-a381-2ef4088b446a
2024/02/14 06:09:39  token: 35d37025-ddde-428e-9a7e-566423e698b6
2024/02/14 06:09:39 🌈️ starting client (scheduler:8080)...
2024/02/14 06:09:39 registering cluster: cluster-blue
2024/02/14 06:09:39 status: REGISTER_SUCCESS
2024/02/14 06:09:39 secret: 78bd0eb8-a9e9-48f1-8762-a27d5625dffb
2024/02/14 06:09:39  token: d3baa9e4-9676-442a-9012-ec5c3462711a
2024/02/14 06:09:39 🌈️ starting client (scheduler:8080)...
2024/02/14 06:09:39 registering cluster: cluster-yellow
2024/02/14 06:09:39 status: REGISTER_SUCCESS
2024/02/14 06:09:39 secret: af790474-d49c-4dd3-ba85-d87ce79b45bc
2024/02/14 06:09:39  token: ec072044-03be-4454-a5d7-2634ad4813e9
```

We also have a script [scripts/register.sh](scripts/register.sh) that you can run to "automate" that command.
We can use that later. Instead of the above, next time you can just do:

```bash
./scripts/register.sh
```

### 3. Poll and Submit Jobs

Now let's just let all hell-o break loose. We are going to use a simple script that registers each cluster, and then submits
jobs randomly to the other clusters. We will do this (very stupidly) by way of a shared `data` directory where we will wait
for all registrations to happen (via a filesystem indicator) and then run a script to register, save the token, and
submit jobs. Let's clean up and start fresh.

```bash
docker compose stop
docker compose rm

# Clean up old secrets and tokens
rm -rf ./data/*.txt

# Bring up a different compose file
docker compose -f docker-compose-demo.yaml up -d
```
This is actually going to start two workers (that will wait for the third) and we are going to launch the third so we can easily watch it.

```bash
docker exec -it cluster-red flux start python3 /data/run-demo.py --peer cluster-red --peer cluster-blue --peer cluster-yellow
```

You'll see running and accepting jobs from all clusters (including ourselves)!

```console
Ran job ƒMXpL8ZV hello from cluster-blue, a new word is rd

status: SUBMIT_SUCCESS

status: SUBMIT_SUCCESS

Status: REQUEST_JOBS_SUCCESS
Received 1 jobs for inspection!
Status: REQUEST_JOBS_SUCCESS
Received 1 jobs for inspection!
Accepting 1 jobs...
[29]
Submit job ['echo', 'hello', 'from', 'cluster-yellow,', 'a', 'new', 'word', 'is', 'ampland']: ƒWS6n2WK
Ran job ƒWS6n2WK hello from cluster-yellow, a new word is ampland

status: SUBMIT_SUCCESS

status: SUBMIT_SUCCESS

status: SUBMIT_SUCCESS

status: SUBMIT_SUCCESS

Status: REQUEST_JOBS_SUCCESS
Received 2 jobs for inspection!
Accepting 2 jobs...
[39, 37]
Submit job ['echo', 'hello', 'from', 'cluster-blue,', 'a', 'new', 'word', 'is', 'specifics']: ƒawCyYAj
Submit job ['echo', 'hello', 'from', 'cluster-blue,', 'a', 'new', 'word', 'is', 'rebecca']: ƒawo5Fef
Ran job ƒawo5Fef hello from cluster-blue, a new word is rebecca

Ran job ƒawCyYAj hello from cluster-blue, a new word is specifics

status: SUBMIT_SUCCESS

status: SUBMIT_SUCCESS

status: SUBMIT_SUCCESS

status: SUBMIT_SUCCESS

Status: REQUEST_JOBS_SUCCESS
Received 1 jobs for inspection!
Accepting 1 jobs...
[46]
Submit job ['echo', 'hello', 'from', 'cluster-red,', 'a', 'new', 'word', 'is', 'captain']: ƒfSU4yXD
Ran job ƒfSU4yXD hello from cluster-red, a new word is captain

status: SUBMIT_SUCCESS

Status: REQUEST_JOBS_SUCCESS
Received 1 jobs for inspection!
Accepting 1 jobs...
[50]
Submit job ['echo', 'hello', 'from', 'cluster-blue,', 'a', 'new', 'word', 'is', 'conservative']: ƒjv7mC7y
Ran job ƒjv7mC7y hello from cluster-blue, a new word is conservative

💤️ Cluster cluster-red is finished! Shutting down.
```

And then you will watch the demo! Be sure to look at the scheduler logs too in another terminal. E.g.,:

```console
scheduler  | 2024/02/14 08:16:05 DELETE FROM jobs WHERE cluster = 'cluster-yellow' AND idJob in (54,62,34): (3)
scheduler  | 2024/02/14 08:16:05 📝️ received register: cluster-blue
scheduler  | 2024/02/14 08:16:05 SELECT count(*) from clusters WHERE name = 'cluster-blue': (1)
scheduler  | 2024/02/14 08:16:05 SELECT * from clusters WHERE name LIKE "cluster-red" LIMIT 1: cluster-red
scheduler  | 2024/02/14 08:16:05 📝️ received job  for cluster cluster-red
scheduler  | 2024/02/14 08:16:05 SELECT * from clusters WHERE name LIKE "cluster-red" LIMIT 1: cluster-red
scheduler  | 2024/02/14 08:16:05 📝️ received job  for cluster cluster-red
scheduler  | 2024/02/14 08:16:05 SELECT * from clusters WHERE name LIKE "cluster-blue" LIMIT 1: cluster-blue
scheduler  | 2024/02/14 08:16:05 📝️ received job  for cluster cluster-blue
scheduler  | 2024/02/14 08:16:05 SELECT * from clusters WHERE name LIKE "cluster-blue" LIMIT 1: cluster-blue
scheduler  | 2024/02/14 08:16:05 🌀️ requesting 0 max jobs for cluster cluster-blue
scheduler  | 2024/02/14 08:16:05 SELECT * from clusters WHERE name LIKE "cluster-blue" LIMIT 1: cluster-blue
scheduler  | 2024/02/14 08:16:05 🌀️ accepting 4 for cluster cluster-blue
scheduler  | 2024/02/14 08:16:05 DELETE FROM jobs WHERE cluster = 'cluster-blue' AND idJob in (51,60,48,47): (4)
scheduler  | 2024/02/14 08:16:15 SELECT * from clusters WHERE name LIKE "cluster-red" LIMIT 1: cluster-red
scheduler  | 2024/02/14 08:16:15 📝️ received job  for cluster cluster-red
```