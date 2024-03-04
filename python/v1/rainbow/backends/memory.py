import grpc

import rainbow.protos.memory_pb2 as memory_pb2
import rainbow.protos.memory_pb2_grpc as memory_pb2_grpc

from .base import GraphBackend

# The memory database backend provides an interface to interact with an in memory cluster database


class MemoryBackend(GraphBackend):
    """
    A MemoryBackend is the rainbow default.

    This graph database backend is primarily for development.
    """

    def satisfies(self, jobspec):
        """
        Determine if a jobspec can be satisfied by the graph.
        """
        # Prepare a satisfy request with the jobspec
        # TODO if auth is in the graph, that needs to be done here too
        request = memory_pb2.SatisfyRequest(payload=jobspec.to_str())

        # Host should be set from the database_options from the client
        with grpc.insecure_channel(self.host) as channel:
            stub = memory_pb2_grpc.MemoryGraphStub(channel)
            response = stub.Satisfy(request)
        return response
