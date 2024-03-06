from .memory import MemoryBackend


def get_backend(name, options):
    """
    Get a named backend.
    """
    name = name.lower()
    if name == "memory":
        return MemoryBackend(**options)
    raise ValueError(f"Backend {name} is not a known graph database backend")
