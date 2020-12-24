# TLSF
## Forward a local TCP connection to a remote TLS-enabled TCP connection
### Motivation
Remove the TLS layer of a TCP connection, to allow interaction with tools which don't have TLS/mTLS capability, such as netcat.
### Usage
```
$ tlsf -h
Usage: tlsf [-no-verify] [-cacert ca_cert] [-cert client_cert] [-key client_key] remote_host:port bind_address:port
        -ca-cert: client CA certificate PEM file location (optional)
        -cert: client certificate PEM file location (optional)
        -key: client key PEM file location (optional)
        -no-verify: skip verifying server certificate (optional, default to false)
```
### Examples
```
# simple
$ tlsf github.com:443 localhost:8000
$ curl localhost:8000 -H 'Host: github.com'

# skip verifying server cert
$ tlsf -no-verify example.com:8443 localhost:8000

# client cert authentication required (mTLS)
$ tlsf -cert cert.pem -key key.pem example.com:8443 localhost:8000
```

### Build
```
git clone https://github.com/wanlong-li/tlsf.git

go build -o dist/tlsf tlsf/cmd/tlsf/main.go

```

