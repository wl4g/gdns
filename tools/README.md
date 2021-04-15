## Other scene special tools

- 1. Resoving of dynamic externet IP

> Obtain the exit IP of the company's intranet host (because the operator will change it at any time), and then use the router (or switch, etc.) port mapping, so as to achieve the same domain name resolution to the dynamic extrnet IP.

Usage:
```
ls extranetip-to-dnsserver

# ---------- Client(Company hosts side) -------
pip3 install flask
clientLog='/mnt/disk1/log/extranetip-to-dnsserver/'
mkdir -p $clientLog
startClientCommand="nohup /usr/bin/extranetip-to-dnsserver-client.py > $clientLog/client.out 2>&1 &"
echo $startClientCommand >> /etc/rc.local # CentOS7

# --------- Server(CoreDNS side) --------
pip3 install flask
pip3 install redis-py-cluster
serverLog='/mnt/disk1/log/extranetip-to-dnsserver/'
mkdir -p $serverLog
startServerCommand="nohup /usr/bin/extranetip-to-dnsserver-server.py > $serverLog/server.out 2>&1 &"
echo $startServerCommand >> /etc/rc.local # CentOS7
```
