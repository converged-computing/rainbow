FROM ubuntu:jammy

LABEL maintainer="Vanessasaurus <@vsoch>"

# Match the default user id for a single system so we aren't root
ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=1000
ENV USERNAME=${USERNAME}
ENV USER_UID=${USER_UID}
ENV USER_GID=${USER_GID}
USER root
ENV DEBIAN_FRONTEND=noninteractive
ENV GO_VERSION=1.20.14

# Install extra buildrequires for flux-sched:
    RUN apt-get update && \
    apt-get -qq install -y --no-install-recommends \
        wget \
        curl \
        less \
        git \
        build-essential \
        ca-certificates \
        unzip && \
    apt-get clean -y && apt -y autoremove

# Install go 19.10
RUN wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz  && tar -xvf go${GO_VERSION}.linux-amd64.tar.gz && \
    mv go /usr/local && rm go${GO_VERSION}.linux-amd64.tar.gz

# ENV GOROOT=/usr/local/go
# ENV GOPATH=/go
ENV PATH=/usr/local/go/bin:$PATH

# Go dependencies for protobuf
RUN apt -y update && apt -y upgrade && apt install --no-install-recommends -y protobuf-compiler curl && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1

# IMPORTANT: install this way and not from package manager because the include directory
# (e.g., with timestamp.proto) needs to be relative to install root in include/
RUN export PB_REL="https://github.com/protocolbuffers/protobuf/releases" && \
    curl -LO $PB_REL/download/v3.15.8/protoc-3.15.8-linux-x86_64.zip && \
    unzip protoc-3.15.8-linux-x86_64.zip -d /usr/local

# Assuming installing to /usr/local
ENV LD_LIBRARY_PATH=/usr/lib:/usr/local/lib
ENV PATH=$PATH:/workspaces/rainbow/bin

# When we upgrade go
#RUN wget --no-check-certificate https://go.dev/dl/go1.21.7.linux-amd64.tar.gz && tar -xvf go1.21.7.linux-amd64.tar.gz && \
#         mv go /usr/local && rm go1.21.7.linux-amd64.tar.gz

ENV PATH=$PATH:/usr/local/go/bin:/home/vscode/go/bin

# Add the group and user that match our ids
RUN groupadd -g ${USER_GID} ${USERNAME} && \
    adduser --disabled-password --uid ${USER_UID} --gid ${USER_GID} --gecos "" ${USERNAME} && \
    echo "${USERNAME} ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers
