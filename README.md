## A coredns enterprise-level plug-in that can obtain zone resolution records from redis-cluster, which can be integrated with super-cloudops [https://github.com/wl4g/super-cloudops](https://github.com/wl4g/ super-cloudops), provides DevOps unified management web GUI

English version goes [here](./README.md)

[Secondary Development] (./INSTALL_CN.md)

### Configuration example

For more configuration items, please refer to the coredns official website. For example, we give a general example:

```hocon
.:53 {
    loadbalance round_robin
    # Load zones records from local /etc/hosts.
    hosts {
        fallthrough
    }
    # Load zones records from redis-cluster.
    coredns_agent {
        address localhost:6379,localhost:6380,localhost:6381,localhost:7379,localhost:7380,localhost:7381
        password "123456"
        connect_timeout 5000
        read_timeout 10000
        write_timeout 5000
        max_retries 10
        pool_size 10
        ttl 360
        prefix _dns:
        local_cache_expire_ms 5000
    }
    # Up recursive DNS query server list.
    # e.g. Google dns servers: 8.8.8.8, china telecom dns servers: 114.114.114.114,202.96.134.133,202.96.212.68
    forward . 8.8.8.8 114.114.114.114
    reload 6s
    log . "{local}:{port} - {>id} '{type} {class} {name} {proto} {size} {>do} {>bufsize}' {rcode} {>rflags} {rsize} {duration}"
    errors
}
```

* `address` redis cluster node address host:port or ip:port, default: localhost:6379,localhost:6380,localhost:6381,localhost:7379,localhost:7380,localhost:7381
* `password` redis cluster password, default: empty
* `connect_timeout` connection timeout time, default: 5000ms
* `read_timeout` data read timeout, default: 10000ms
* `write_timeout` data write timeout, default: 5000ms
* `max_retries` Maximum number of retries, default: 10
* `pool_size` redis connection pool size, default: 10
* `ttl` zones resolve cache ttl, default: 360sec
* `prefix` zones resolution record data is stored in redis-cluster key prefix, default: _dns:
* `local_cache_expire_ms` zones resolving and record the validity period of the local cache, default: 5000ms (Note: In order to improve performance, the loading sequence of zones map data is in order: localCache -> redisCache -> db)


### Reverse resolution

Currently does not support direction resolution

### Proxy resolution

Currently does not support direction resolution

### Zones resolving records are stored in redis-cluster data format

Each zone is stored as a hash map in redis-cluster, with zone as the key. Note: According to the https://tools.ietf.org/html/rfc6763 protocol, it ends with a "." suffix. Such as:

```
redis-cli>KEYS *
1) "example.com."
2) "example.net."
redis-cli>
```

#### dns RRs

Stored in redis cluster in json string format, *@* is used for RR value of the region itself. Such as:

#### A

```json
{
    "a":{
        "ip": "1.2.3.4",
        "ttl": 360
    }
}
```

#### AAAA

```json
{
    "aaaa":{
        "ip": "::1",
        "ttl": 360
    }
}
```

#### CNAME

```json
{
    "cname":{
        "host": "x.example.com.",
        "ttl": 360
    }
}
```

#### TXT

```json
{
    "TXT":{
        "text": "this is a text",
        "ttl": 360
    }
}
```

#### NS

```json
{
    "ns":{
        "host": "ns1.example.com.",
        "ttl": 360
    }
}
```

#### MX

```json
{
    "mx":{
        "host": "mx1.example.com",
        "priority": 10,
        "ttl": 360
    }
}
```

#### SRV

```json
{
    "srv":{
        "host": "sip.example.com.",
        "port": 555,
        "priority": 10,
        "weight": 100,
        "ttl": 360
    }
}
```

#### SOA

```json
{
    "soa":{
        "ttl": 100,
        "mbox": "hostmaster.example.com.",
        "ns": "ns1.example.com.",
        "refresh": 44,
        "retry": 55,
        "expire": 66
    }
}
```

#### CAA

```json
{
    "caa":{
        "flag": 0,
        "tag": "issue",
        "value": "letsencrypt.org"
    }
}
```

### Parsing example

```
$ORIGIN example.net.
 example.net. 300 IN SOA <SOA RDATA>
 example.net. 300 NS ns1.example.net.
 example.net. 300 NS ns2.example.net.
 *.example.net. 300 TXT "this is a wildcard"
 *.example.net. 300 MX 10 host1.example.net.
 sub.*.example.net. 300 TXT "this is not a wildcard"
 host1.example.net. 300 A 5.5.5.5
 _ssh.tcp.host1.example.net. 300 SRV <SRV RDATA>
 _ssh.tcp.host2.example.net. 300 SRV <SRV RDATA>
 subdel.example.net. 300 NS ns1.subdel.example.net.
 subdel.example.net. 300 NS ns2.subdel.example.net.
 host2.example.net CAA 0 issue "letsencrypt.org"
```

The above zone data should be stored in redis-cluster as follows:

```
redis-cli> hgetall example.net.
 1) "_ssh._tcp.host1"
 2) "{\"srv\":[{\"ttl\":300, \"target\":\"tcp.example.com.\",\"port\":123,\"priority\" :10,\"weight\":100}]}"
 3) "*"
 4) "{\"txt\":[{\"ttl\":300, \"text\":\"this is a wildcard\"}],\"mx\":[{\"ttl\" :300, \"host\":\"host1.example.net.\",\"preference\": 10}]}"
 5) "host1"
 6) "{\"a\":[{\"ttl\":300, \"ip\":\"5.5.5.5\"}]}"
 7) "sub.*"
 8) "{\"txt\":[{\"ttl\":300, \"text\":\"this is not a wildcard\"}]}"
 9) "_ssh._tcp.host2"
10) "{\"srv\":[{\"ttl\":300, \"target\":\"tcp.example.com.\",\"port\":123,\"priority\" :10,\"weight\":100}]}"
11) "subdel"
12) "{\"ns\":[{\"ttl\":300, \"host\":\"ns1.subdel.example.net.\"},{\"ttl\":300, \ "host\":\"ns2.subdel.example.net.\"}]}"
13) "@"
14) "{\"soa\":{\"ttl\":300, \"minttl\":100, \"mbox\":\"hostmaster.example.net.\",\"ns\": \"ns1.example.net.\",\"refresh\":44,\"retry\":55,\"expire\":66},\"ns\":[{\"ttl\": 300, \"host\":\"ns1.example.net.\"},{\"ttl\":300, \"host\":\"ns2.example.net.\"}]}"
15) "host2"
16)"{\"caa\":[{\"flag\":0, \"tag\":\"issue\", \"value\":\"letsencrypt.org\"}]}"
redis-cli>
```
