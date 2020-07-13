### Secondary development coredns-redisc

#### 1, download the project
First, git clone https://github.com/coredns/coredns (you need to download the coredns main library project before running external plugins)

#### 2, configure the plugin
Modify the configuration file coredns/plugin.cfg, for example, add our plug-in on the front line of the forward plug-in (note: because caddy used by coredns will determine the execution order according to the plug-in configuration order), the advantage before putting it in the forward plug-in is that it can be controlled DNS recursively parses queries upwards.

```
vim coredns/plugin.cfg
...
#The development environment recommends using the local directory name coredns-redisc directly, without using the github.com/wl4g/coredns-redisc address.
coredns-redisc:coredns-redisc
#coredns-redisc:github.com/wl4g/coredns-redisc
forward:forward
...
```

#### 3, compile (merge plugin)
Before executing make, you can modify the Makefile to modify the configuration to achieve cross compilation, such as:

```
Add "GOOS=linux GOARCH=amd64" after SYSTEM:=, then the binary file of the Linux system is generated:
SYSTEM:=GOOS=linux GOARCH=amd64
SYSTEM:=GOOS=windows GOARCH=amd64
SYSTEM:=GOOS=darwin GOARCH=amd64
```

#### 4, configuration file Corefile

For more configuration items, please refer to the coredns official website. For example, we give a general example:

```
.:53 {
    # Load zones records from local /etc/hosts.
    hosts {
        fallthrough
    }
    # Load zones records from redis-cluster.
    coredns-redisc {
        address localhost:6379,localhost:6380,localhost:6381,localhost:7379,localhost:7380,localhost:7381
        password "123456"
        connect_timeout 30000
        read_timeout 30000
        ttl 360
        prefix _dns:
    }
    # Up recursive DNS query server list.
    # e.g. Google dns servers: 8.8.8.8, china telecom dns servers: 114.114.114.114,202.96.134.133,202.96.212.68
    forward. 8.8.8.8 114.114.114.114
    log
}
```

#### 5, Start running

If everything is normal, the coredns execution file will be generated in the coredns/ directory after compilation and start running:

```
./coredns -conf Corefile
```

#### 6, Tests run

Add test data:
```
redis-cli> hset example.net. me "{\"a\":[{\"ttl\":300, \"ip\":\"10.0.0.166\"}]}"
```

dns client query test:
```
dig me.example.net


; <<>> DiG 9.11.4-P2-RedHat-9.11.4-9.P2.el7 <<>> me.example.net
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 2609
;; flags: qr rd ra; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 1

;; OPT PSEUDOSECTION:
; EDNS: version: 0, flags:; udp: 4096
;; QUESTION SECTION:
;me.example.net. IN A

;; ANSWER SECTION:
me.example.net. 600 IN A 10.0.0.166

;; Query time: 2664 msec
;; SERVER: 100.100.2.138#53(100.100.2.138)
;; WHEN: Mon Jul 13 12:58:47 CST 2020
;; MSG SIZE rcvd: 53
```