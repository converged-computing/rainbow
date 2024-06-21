from contextlib import contextmanager

import grpc


@contextmanager
def grpc_channel(host, use_ssl=False):
    """
    Yield a channel, either with or without ssl, and close properly.
    """
    if use_ssl:
        channel = grpc.secure_channel(host, grpc.ssl_channel_credentials())
    else:
        channel = grpc.insecure_channel(host)
    try:
        yield channel
    finally:
        channel.close()
