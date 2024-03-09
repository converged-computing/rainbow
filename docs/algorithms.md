# Algorithms

This is a brief summary of notes about current algorithms.

## Memory Graph

The "memory" graph backend is an in-memory graph database that is a custom implementation (by @vsoch). Although it is primarily intended for learning, it serves as a good base for development and prototyping too, and warrants a discussion of algorithms involved. For design, see the [design](design.md) document. This will detail basics about the search.

### Depth First Search

While Fluxion uses depth first search and up (to support an adjacency list), since we are just using this graph for prototyping, we instead use recursion, which means we can traverse (depth) and not need to find our way back up, because we can return from a recursive call.

#### 1. Quick Check

We start with a hieuristic that says "if I know the totals that are needed for this Jobspec are not available across the cluster, bail out before doing any search." That works as follows.

1. We start with a Jobspec and an empty list of matches (that will be clusters)
2. We generate a lookup of totals by the resource type, where the key is the type, and value is the count needed.
3. We defnie a recursive function that takes a "resource" section in the Jobspec and is able to do a check for a single entry. E.g., if we are parsing through a jobspec and find node with a count of 2, we add that to our lookup.
4. We run this recursively across the Jobspec resources.

At the end, we have a summary of the total resources requested by the jobspec, and do a quick check to see if any clusters have less than that amount (the totals we already have cached from registration) OR if the clusters are missing a resource entirely. Note that this is only for the dominant subsystem. If a cluster passes these checks, it proceeds into depth first search.

#### 2. Depth First Search

Depth first search is going to do checks from the perspective of a slot, because this (as I understand it) is the level where we are "pinning" the request. Thus, we start our search by creating a lookup of slots, which we do from the "tasks" section of the jobspec. We do this because as we are traversing we are going to be randomly hitting slots defined by the user, and we need to be able to look up details about it.

Note that this search is still rooted in the dominant subsystem, and for other subsystem resources (e.g., IO) these are going to linked off of vertices here. For each cluster in our matches, we then start at the root, which is generally just a node named by the cluster. We get that vertex, because since this memory database has an object oriented design, all children vertices are going to be edges off of that.

##### findSlots

We then define a recursive function `findSlots` that is going to recurse into a slot resource and recurse into child resources under that to count what it finds. For example, if the Jobspec is saying that it wants some number of cores per slot, the `findSlots` function will start at a vertex where the slot is, and then figure out if we have that number. It returns a number that represents that count. Specifically, the function works as follows:

1. We start with input the vertex to start the search and the resource root from the Jobspec
2. We assume starting with 0 slots found.
  - If the vertex type we are at is the slot type we are looking for, our search is done for this section, and we can return the size of the vertex (that represents the number of the resource type it has)
  - If the vertex type isn't the resource type, we likely need to keep searching its edges looking for the slot we want. We traverse into the child edges that have the "contains" relationship and call the function recursively.
  - In early implementations, we did not consister subsystem resources. Now, given that a slot has a subsystem defined, and given that we find an edge that points to a subsystem, we check the needs defined in the Jobspec against what the subsystem has. This check is based on the subsystem matching algorithm, which is currently just looking for matched key value pairs. When an entire subsystem is satisfied, we set a single boolean so we do not check again in the future, and when we determine if the slot is satisfied, we can do one check to this boolean.

The function `findSlots` will (should) return with the number of matches for a specific resource type below a vertex in the graph, allowing us to determine if a subtree can match a request that is specific to a slot.

##### satisfies

Satisfies is a recursive function that determines if a vertex can satisfy a resource need.
Given a resource and a vertex root, it returns the count of vertices under the root that satisfy the request. This function uses `findSlots` because as it is traversing, when it finds a `resource.Type`
of type "slot" it will call that function. Akin to `findSlots`, it works as follows:

1. We first check if the resource type matches the vertex type we are at.
  - If not, we iterate through edges and look for "contains" relations. For each contains relation, we recursively call satisfies for the same resource, and add the result (a count) to our total count of the number found. When found is >= the resource count requested, we break (and return from the recursive function) because we've found the first match.
  - If the type matches, then we can return the current value of found plus the new resource count we find at this vertex.

The result of satisfies is returning the count for some resource that is satisfied starting at some root, accounting for slots too.

##### traverseResource

The traverse resource is the main (also recursive function) to handle traversing the graph. It starts at the top level resource from the Jobspec, and instead of returning a count, returns a boolean to indicate if the match is a yes or no. It has two cases:

1. If it finds a slot, it starts with the number of slots that are requested or needed, and sets a counter of "slots found" to zero. It then starts to traverse the "resource.With" that represents in the Jobspec the resources under the slot. It calls the recursive function `findSlots` that will return the count of resources needed at that vertex. If we have enough, we return early because we have a match.
2. If we don't have a slot, we instead call "satisfies," which (internally) can also handle hitting a slot and calling "findSlots" as discussed above. On this first call we set our "found" counter to 0 since we are just starting a recursive search.

The final check after satisfies is to see if we found enough matches, and then add the cluster to be a contender for assignment. Outside of the recursive function, the "get it all going" logic is very simple, and looks like:

```console
# pseudocode
for resource in jobspec.resources:
  isMatch = traverseResource(resource)
  if !isMatch -> break early

  for each subresource (resource.With):
    isMatch = traverseResource(subresource)
    if !isMatch -> break early

if isMatch is true here, add the cluster to matches
```

At this point, the basic list of clusters is returned to the calling function (the interface in rainbow) and passed on to a selection algorithm, which can take some logic about the clusters (likely state) and make a final decision. We currently just randomly select from the set (random is the only selection algorithm available, mainly for development).

## Jobspec Resources

While we need to have more [discussion](https://github.com/flux-framework/flux-sched/discussions/1153#discussioncomment-8726678) on what constitutes a request for subsystem resources, I am taking a simple approach that will satisfy an initial need to run experiments with compatibility metadata (relevant to subsystems) that use a scheduler. The approach I am taking is the following. You can read about the [design](design.md) and I'll repeat the high level points here. When we register a subsystem, it is a separate graph that (at the highest level) is still organized by cluster name. However, each node in the graph needs to be attached to another node known to itself, or to a vertex in the dominant subsystem graph. When asking for a subsystem resource, we are asking for a check at a specific vertex (defined by the slot) that is relevant for a specific subsystem and resource type. We do this by way of defining "resources" under a task, as shown below:

```yaml
version: 1
resources:
- count: 2
  type: node
  with:
  - count: 1
    label: default
    type: slot
    with:
    - count: 2
      type: core
tasks:
- command:
  - ior
  slot: default
  count:
    per_slot: 1
  resources:
    io:
      match:
      - type: shm
```

In the above, we are saying that when we find a slot, we need to see if the vertex has an edge to the "ior" subsystem with this particular kind of storage (shared memory). If it does, it's a match. Note that this basic structure is currently enforced - the top level under resources are keys for subsystems, under those keys are algorithm types, of which the current only available option is to "match," which means an exact match of resource types. This is the "does the slot have features, yes or no" approach, and is done intentionally to satisfy simple experiments that describe subsystem resources as present or not. Different algorithm types can be defined here that implement different logic (taking into account counts, or actually traversing the subsystem graph at that vertex, which currently is not done).

I understand this is likely not perfect for what everyone wants, but I believe it to be a reasonable first shot, and within the ability of what I can prototype without having fluxion ready yet.

[home](/README.md#rainbow-scheduler)
