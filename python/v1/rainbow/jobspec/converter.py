import shlex


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
