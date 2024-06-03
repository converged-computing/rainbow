import time

import grpc

from rainbow.protos import rainbow_pb2, rainbow_pb2_grpc


def with_time(name, host, funcName, request, metadata):
    """
    Make an API call and time it.

    A function (or metric name) is required, along with the
    name of the function. Kwargs are optional, and should be metadata.

    name: is the name of the API endpoint for the stub
    host: hostname for client endpoint
    funcName: the "key" of the metric (function being timed)
    request: request to go to stub
    metadata: arbitrary dict of metadata
    """
    with grpc.insecure_channel(host) as channel:
        stub = rainbow_pb2_grpc.RainbowSchedulerStub(channel)
        func = getattr(stub, name)
        start = time.time()
        response = func(request)
        end = time.time()
        saveRequest = rainbow_pb2.SaveMetricRequest(
            name=funcName, value=str(end - start), metadata=metadata
        )
        saveResponse = stub.SaveMetric(saveRequest)
    return response, saveResponse


def save_metric(host, name, value, metadata):
    """
    Save time to the database
    """
    saveRequest = rainbow_pb2.SaveMetricRequest(name=name, value=value, metadata=metadata)
    with grpc.insecure_channel(host) as channel:
        stub = rainbow_pb2_grpc.RainbowSchedulerStub(channel)
        response = stub.SaveMetric(saveRequest)
    return response
