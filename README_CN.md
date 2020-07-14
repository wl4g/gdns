## 一个coredns企业级插件，实现从redis-cluster获取zones解析记录，可与super-cloudops整合[https://github.com/wl4g/super-cloudops](https://github.com/wl4g/super-cloudops)，提供DevOps统一管理web GUI

English version goes [here](./README.md)

[二次开发](./INSTALL_CN.md)

### 配置示例

更多配置项可参考coredns官网查看，如 我们给出常规示例：

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

* `address` redis集群节点地址 host:port or ip:port，默认: localhost:6379,localhost:6380,localhost:6381,localhost:7379,localhost:7380,localhost:7381
* `password` redis集群密码，默认：空
* `connect_timeout` 连接超时时间，默认：5000ms
* `read_timeout` 数据读取超时时间，默认：10000ms
* `write_timeout` 数据写入超时时间，默认：5000ms
* `max_retries` 最大重试次数，默认：10
* `pool_size` redis连接池大小，默认：10
* `ttl` zones解析缓存ttl，默认：360sec
* `prefix` zones解析记录数据存储在redis-cluster的key前缀，默认：_dns:
* `local_cache_expire_ms` zones解析记录本地高速缓存的有效期，默认：5000ms (说明: 为了提高性能, zones映射数据加载顺序依次为:  localCache -> redisCache -> db)


### 反向解析

目前暂不支持反向解析

### 代理

目前暂不支持代理解析

### zones解析记录存储在redis-cluster中的数据格式

每个zone作为散列映射存储在redis-cluster中，以zone作为key。注：按照https://tools.ietf.org/html/rfc6763规约，以“.”后缀结尾。如：

```
redis-cli>KEYS *
1) "example.com."
2) "example.net."
redis-cli>
```

#### dns RRs

以json字符串格式存储在redis集群中，*@* 用于区域自身的RR值。如：

#### A

```json
{
    "a":{
        "ip" : "1.2.3.4",
        "ttl" : 360
    }
}
```

#### AAAA

```json
{
    "aaaa":{
        "ip" : "::1",
        "ttl" : 360
    }
}
```

#### CNAME

```json
{
    "cname":{
        "host" : "x.example.com.",
        "ttl" : 360
    }
}
```

#### TXT

```json
{
    "txt":{
        "text" : "this is a text",
        "ttl" : 360
    }
}
```

#### NS

```json
{
    "ns":{
        "host" : "ns1.example.com.",
        "ttl" : 360
    }
}
```

#### MX

```json
{
    "mx":{
        "host" : "mx1.example.com",
        "priority" : 10,
        "ttl" : 360
    }
}
```

#### SRV

```json
{
    "srv":{
        "host" : "sip.example.com.",
        "port" : 555,
        "priority" : 10,
        "weight" : 100,
        "ttl" : 360
    }
}
```

#### SOA

```json
{
    "soa":{
        "ttl" : 100,
        "mbox" : "hostmaster.example.com.",
        "ns" : "ns1.example.com.",
        "refresh" : 44,
        "retry" : 55,
        "expire" : 66
    }
}
```

#### CAA

```json
{
    "caa":{
        "flag" : 0,
        "tag" : "issue",
        "value" : "letsencrypt.org"
    }
}
```

### 解析示例

```
$ORIGIN example.net.
 example.net.                 300 IN  SOA   <SOA RDATA>
 example.net.                 300     NS    ns1.example.net.
 example.net.                 300     NS    ns2.example.net.
 *.example.net.               300     TXT   "this is a wildcard"
 *.example.net.               300     MX    10 host1.example.net.
 sub.*.example.net.           300     TXT   "this is not a wildcard"
 host1.example.net.           300     A     5.5.5.5
 _ssh.tcp.host1.example.net.  300     SRV   <SRV RDATA>
 _ssh.tcp.host2.example.net.  300     SRV   <SRV RDATA>
 subdel.example.net.          300     NS    ns1.subdel.example.net.
 subdel.example.net.          300     NS    ns2.subdel.example.net.
 host2.example.net                    CAA   0 issue "letsencrypt.org"
```

以上zones数据应存储在redis-cluster中，如下所示：

```
redis-cli> hgetall example.net.
 1) "_ssh._tcp.host1"
 2) "{\"srv\":[{\"ttl\":300, \"target\":\"tcp.example.com.\",\"port\":123,\"priority\":10,\"weight\":100}]}"
 3) "*"
 4) "{\"txt\":[{\"ttl\":300, \"text\":\"this is a wildcard\"}],\"mx\":[{\"ttl\":300, \"host\":\"host1.example.net.\",\"preference\": 10}]}"
 5) "host1"
 6) "{\"a\":[{\"ttl\":300, \"ip\":\"5.5.5.5\"}]}"
 7) "sub.*"
 8) "{\"txt\":[{\"ttl\":300, \"text\":\"this is not a wildcard\"}]}"
 9) "_ssh._tcp.host2"
10) "{\"srv\":[{\"ttl\":300, \"target\":\"tcp.example.com.\",\"port\":123,\"priority\":10,\"weight\":100}]}"
11) "subdel"
12) "{\"ns\":[{\"ttl\":300, \"host\":\"ns1.subdel.example.net.\"},{\"ttl\":300, \"host\":\"ns2.subdel.example.net.\"}]}"
13) "@"
14) "{\"soa\":{\"ttl\":300, \"minttl\":100, \"mbox\":\"hostmaster.example.net.\",\"ns\":\"ns1.example.net.\",\"refresh\":44,\"retry\":55,\"expire\":66},\"ns\":[{\"ttl\":300, \"host\":\"ns1.example.net.\"},{\"ttl\":300, \"host\":\"ns2.example.net.\"}]}"
15) "host2"
16)"{\"caa\":[{\"flag\":0, \"tag\":\"issue\", \"value\":\"letsencrypt.org\"}]}"
redis-cli>
```
