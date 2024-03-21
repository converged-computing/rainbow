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
