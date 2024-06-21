# Authentication

Rainbow currently supports two modes of running:

- No authentication (default)
- With certificates (with `--ssl` flag)


Without authentication, just use rainbow as is.

## With Authentication

### Self-signed Certificates

Ideally, you have generated signed certificates for your server. For self signed certificates, use:

```bash
make certs
```

And they will be generate for you in the local bin, both for the server and a client:

```bash
bin/certs/
├── ca-cert.pem
├── ca-cert.srl
├── ca-key.pem
├── client-cert.pem
├── client-key.pem
├── client-req.pem
├── server-cert.pem
├── server-key.pem
└── server-req.pem
```

### Server with TLS

Then you can run the server using these paths as the defaults:

```console
$ make server-tls
```
```console
go run cmd/server/server.go --global-token rainbow -cert /home/vanessa/Desktop/Code/rainbow/bin/certs/server-cert.pem -ca-cert /home/vanessa/Desktop/Code/rainbow/bin/certs/ca-cert.pem --key /home/vanessa/Desktop/Code/rainbow/bin/certs/server-key.pem
2024/04/15 18:17:46 creating 🌈️ server...
2024/04/15 18:17:46 🧩️ selection algorithm: random
2024/04/15 18:17:46 🧩️ graph database: memory
2024/04/15 18:17:46 ✨️ creating rainbow.db...
2024/04/15 18:17:46    rainbow.db file created
2024/04/15 18:17:46    🏓️ creating tables...
2024/04/15 18:17:46    🏓️ tables created
2024/04/15 18:17:46 ⚠️ WARNING: global-token is set, use with caution.
2024/04/15 18:17:46 starting scheduler server: rainbow v0.1.1-draft
2024/04/15 18:17:46 🔐️ adding tls credentials
2024/04/15 18:17:46 🧠️ Registering memory graph database...
2024/04/15 18:17:46 server listening: [::]:50051
```

Note the command above - you can customize this to use your own paths. Also note that when the server starts with credentials, you will see `🔐️ adding tls credentials`.

### Client with TLS

The client takes a similar path, but instead uses the client-*.pem files referenced above.
With the server running, try doing a register without credentials.

```bash
make register
```

You'll note that it hangs, and doesn't work! Now we can add the client credentials to get a different response:

```bash
go run cmd/rainbow/rainbow.go register cluster --cluster-name keebler --nodes-json ./docs/examples/scheduler/cluster-nodes.json \
    --config-path ./docs/examples/scheduler/rainbow-config.yaml --save \
    --cert /home/vanessa/Desktop/Code/rainbow/bin/certs/client-cert.pem --ca-cert /home/vanessa/Desktop/Code/rainbow/bin/certs/ca-cert.pem --key /home/vanessa/Desktop/Code/rainbow/bin/certs/client-key.pem
```
```console
2024/06/21 06:51:02 🌈️ starting client (localhost:50051)...
2024/06/21 06:51:02 🔐️ adding tls credentials
2024/06/21 06:51:02 registering cluster: keebler
2024/06/21 06:51:02 status: REGISTER_SUCCESS
2024/06/21 06:51:02 secret: 618c0e98-f4a8-401c-920e-702f6a917c76
2024/06/21 06:51:02  token: rainbow
2024/06/21 06:51:02 Saving cluster secret to ./docs/examples/scheduler/rainbow-config.yaml
```

Note that I followed the tutorial instructions [here](https://medium.com/@mertkimyonsen/securing-grpc-connection-with-ssl-tls-certificate-using-go-db3852fe89dd), although there were a few hairy bits.

[home](/README.md#rainbow-scheduler)
