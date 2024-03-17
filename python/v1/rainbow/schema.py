rainbow_config_v1 = {
    "$schema": "http://json-schema.org/draft-07/schema#",
    "$id": "http://github.com/flux-framework/rfc/tree/master/data/spec_24/schema.json",
    "title": "rainbow-config-01",
    "description": "Rainbow config version 1.0",
    "type": "object",
    "required": ["scheduler", "graphdatabase"],
    "properties": {
        "scheduler": {
            "description": "metadata for the rainbow scheduler",
            "type": "object",
            "properties": {
                "name": {"type": "string"},
                "secret": {"type": "string"},
                "algorithm": {
                    "type": "object",
                    "properties": {
                        "name": {"type": "string"},
                        "options": {"type": "object"},
                    },
                },
                "user": {"type": "object"},
            },
            "additionalProperties": False,
        },
        "graphdatabase": {
            "description": "metadata for the rainbow graph database",
            "type": "object",
            "properties": {
                "name": {"type": "string"},
                "options": {"type": "object"},
            },
            "additionalProperties": False,
        },
        "clusters": {
            "description": "listing of known clusters to submit to",
            "type": "array",
            "items": {
                "type": "object",
                "properties": {
                    "name": {"type": "string"},
                    "token": {"type": "string"},
                },
                "additionalProperties": False,
            },
            "additionalProperties": False,
        },
    },
}

# Note that this has experimental features added, they are flagged
# So it is not technically v1 :)
jobspec_v1 = {
    "$schema": "http://json-schema.org/draft-07/schema#",
    "$id": "http://github.com/flux-framework/rfc/tree/master/data/spec_24/schema.json",
    "title": "jobspec-01",
    "description": "Flux jobspec version 1",
    "definitions": {
        "intranode_resource_vertex": {
            "description": "schema for resource vertices within a node, cannot have child vertices",
            "type": "object",
            "required": ["type", "count"],
            "properties": {
                "type": {"enum": ["core", "gpu"]},
                "count": {"type": "integer", "minimum": 1},
                "unit": {"type": "string"},
            },
            "additionalProperties": False,
        },
        "node_vertex": {
            "description": "schema for the node resource vertex",
            "type": "object",
            "required": ["type", "count", "with"],
            "properties": {
                "type": {"enum": ["node"]},
                "count": {"type": "integer", "minimum": 1},
                "unit": {"type": "string"},
                "with": {
                    "type": "array",
                    "minItems": 1,
                    "maxItems": 1,
                    "items": {"oneOf": [{"$ref": "#/definitions/slot_vertex"}]},
                },
            },
            "additionalProperties": False,
        },
        "slot_vertex": {
            "description": "special slot resource type - label assigns to task slot",
            "type": "object",
            "required": ["type", "count", "with", "label"],
            "properties": {
                "type": {"enum": ["slot"]},
                "count": {"type": "integer", "minimum": 1},
                "unit": {"type": "string"},
                "label": {"type": "string"},
                "exclusive": {"type": "boolean"},
                "with": {
                    "type": "array",
                    "minItems": 1,
                    "maxItems": 2,
                    "items": {"oneOf": [{"$ref": "#/definitions/intranode_resource_vertex"}]},
                },
            },
            "additionalProperties": False,
        },
    },
    "type": "object",
    # NOTE that I removed resources, I don't see why they need to be required
    "required": ["version", "resources", "tasks"],
    # "required": ["version", "resources", "attributes", "tasks"],
    "properties": {
        "version": {
            "description": "the jobspec version",
            "type": "integer",
            "enum": [1],
        },
        "resources": {
            "description": "requested resources",
            "type": "array",
            "minItems": 1,
            "maxItems": 1,
            "items": {
                "oneOf": [
                    {"$ref": "#/definitions/node_vertex"},
                    {"$ref": "#/definitions/slot_vertex"},
                ]
            },
        },
        "attributes": {
            "description": "system and user attributes",
            "type": ["object", "null"],
            "properties": {
                "system": {
                    "type": "object",
                    "properties": {
                        "duration": {"type": "number", "minimum": 0},
                        "cwd": {"type": "string"},
                        "environment": {"type": "object"},
                    },
                },
                "user": {"type": "object"},
            },
            "additionalProperties": False,
        },
        "tasks": {
            "description": "task configuration",
            "type": "array",
            "maxItems": 1,
            "items": {
                "type": "object",
                "required": ["slot", "count"],
                # Command is not required in favor of having batch and script too
                # "required": ["command", "slot", "count"],
                "properties": {
                    "command": {
                        "type": ["string", "array"],
                        "minItems": 1,
                        "items": {"type": "string"},
                    },
                    # RESOURCES ARE EXPERIMENTAL
                    "resources": {"type": "object"},
                    "batch": {"type": "string"},
                    "script": {"type": "string"},
                    "slot": {"type": "string"},
                    "count": {
                        "type": "object",
                        "additionalProperties": False,
                        "properties": {
                            "per_slot": {"type": "integer", "minimum": 1},
                            "total": {"type": "integer", "minimum": 1},
                        },
                    },
                },
                "additionalProperties": False,
            },
        },
    },
}
