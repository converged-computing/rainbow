import shlex


def from_compatibility_spec(
    compspec, command, nodes, name=None, tasks=1, jobspec_version=1, attributes=None
):
    """
    Generate a jobspec from a compatibility spec.
    """
    # Start with a basic one we will modify
    js = new_simple_jobspec(command, nodes, name, tasks, jobspec_version)

    # This is an optional lookup that can filter down to attributes of interest
    attributes = attributes or {}

    # The first task is lammps
    task = js["tasks"][0]

    # Generate task resources based on compatibility metadata
    # Note that each of these has associated graphs we aren't using
    resources = {}
    for compat in compspec.get("compatibilities", []):
        # Treat these flat for now. E.g., io.archspec instead of io -> archspec
        resource_set = {}

        # Here we are creating a set of attributes for a subsystem.
        # E.g., "For io.archpsec I care about cpu.target"
        name = compat.get("name")

        # Skip those we don't care about for the level
        if not name or (attributes and name not in attributes):
            continue

        for attrname, attrvalue in compat.get("attributes", {}).items():
            if attributes and attrname not in attributes[name]:
                continue
            resource_set[attrname] = attrvalue

        # Only add resource sets with at least one attribute
        if resource_set:
            resources[compat["name"]] = resource_set

    if resources:
        task["resources"] = resources

    js["tasks"][0] = task
    return js


def new_simple_jobspec(command, nodes=1, name=None, tasks=1, jobspec_version=1):
    """
    Generate a new simple jobspec from basic parameters.
    """
    if isinstance(command, str):
        command = shlex.split(command)

    if not command:
        raise ValueError("A command must be provided.")

    # If we don't have a name, derive one
    if name is None:
        name = command[0]

    if nodes < 1 or tasks < 1:
        raise ValueError("Nodes and tasks for the job must be >= 1")

    node_resource = {
        "type": "node",
        "count": nodes,
    }
    slot = {
        "type": "slot",
        "count": 1,
        "label": name,
    }
    task_resource = {
        "type": "core",
        "count": tasks,
    }
    slot["with"] = [task_resource]

    node_resource["with"] = [slot]
    task_resources = [
        {
            "command": command,
            "slot": name,
            "count": {
                "per_slot": 1,
            },
        }
    ]
    return {
        "version": jobspec_version,
        "resources": [node_resource],
        "tasks": task_resources,
        "attributes": {},
    }
