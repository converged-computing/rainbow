ARG base="ubuntu:jammy"
FROM $base
USER root
LABEL MAINTAINER Author <vsoch>
ARG arch=amd64
ENV arch=$arch

# install go 20.10
RUN apt-get update && apt-get install -y wget python3-pip
RUN wget https://go.dev/dl/go1.20.10.linux-${arch}.tar.gz  && tar -xvf go1.20.10.linux-${arch}.tar.gz && \
         mv go /usr/local && rm go1.20.10.linux-${arch}.tar.gz

ENV PATH=/usr/local/go/bin:$PATH
WORKDIR /code
COPY . /code
RUN make build && \
    cp ./bin/rainbow /usr/bin && \
    cp ./bin/rainbow-scheduler /usr/local/bin/

# ensure we install the python bindings
RUN cd /code/python/v1 && \
    python3 -m pip install .

ENTRYPOINT ["rainbow-scheduler"]

# Anticipate different running contexts
EXPOSE 80
EXPOSE 8080
EXPOSE 443

# Recommended to add --secret here! If you need to persist the database, mount that
# or we can add support for non sqlite.
CMD ["--address", "0.0.0.0:8080", "--name", "rainbow"]
