# Other scene special tools

- ## 1. Resoving of dynamic externet IP

> Obtain the exit IP of the company's intranet host (because the operator will change it at any time), and then use the router (or switch, etc.) port mapping, so as to achieve the same domain name resolution to the dynamic extrnet IP.

Usage:

```bash
cd $PROJECT_HOME/outlink-dns

# ---------- Client(e.g Company intranet hosts client side) -------
pip3 install flask
clientLogDir='/mnt/disk1/log/outlink-dns'
mkdir -p $clientLogDir
startClientCommand="nohup /usr/bin/outlink-dns-client.py > $clientLogDir/client.out 2>&1 &"
echo $startClientCommand >> /etc/rc.local # CentOS7

# --------- Server(e.g CoreDNS extranet server side) --------
pip3 install flask
pip3 install redis-py-cluster
serverLogDir='/mnt/disk1/log/outlink-dns'
mkdir -p $serverLogDir
startServerCommand="nohup /usr/bin/outlink-dns-server.py > $serverLogDir/server.out 2>&1 &"
echo $startServerCommand >> /etc/rc.local # CentOS7
```
