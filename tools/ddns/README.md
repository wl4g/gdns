# Dynamic Domain Name System Tool

> Obtain the exit IP of the company's intranet host (because the operator will change it at any time), and then use the router (or switch, etc.) port mapping, so as to achieve the same domain name resolution to the dynamic extrnet IP.

## Compiling Installcation

```bash
# Installing to /bin/ddns-client
./install.sh -c -b

# Installing to /bin/ddns-server
./install.sh -s -b
```

## Environments Configuration

| Environment | Default | Side | Description |
| --- | --- | --- | --- |
| COREDNS_DDNS_KEY | abcdefghijklmnopqrstuvwxyz | client/server | Key for interface security signature |
| COREDNS_DDNS_REDIS_NODES | 127.0.0.1:6379,127.0.0.1:6380,127.0.0.1:6381,127.0.0.1:7379,127.0.0.1:7380,127.0.0.1:7381 | server | Redis cluster node collection, separated by commas |
| COREDNS_DDNS_REDIS_PASSWORD | 123456 | server | Redis cluster password |
| COREDNS_DDNS_PREFIX | _coredns:_ | server | Update the DDNS resolution value to the key prefix of redis |
| COREDNS_DDNS_DOMAIN | example.com | server | Update the first level domain name saved to redis through DDNS resolution |
| COREDNS_DDNS_LISTEN_ADDR | 0.0.0.0 | server | Address to start service listening |
| COREDNS_DDNS_LISTEN_PORT | 4008 | server | Port to start service listening |
| COREDNS_DDNS_SERVER_ADDR | 127.0.0.1 | client | Server address (exclude port) |
| COREDNS_DDNS_SERVER_PORT | 4008 | client | Server port |
