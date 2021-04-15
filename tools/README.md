## Other scene special tools

- 1. Resoving of dynamic externet IP

> Obtain the exit IP of the company's intranet host (because the operator will change it at any time), and then use the router (or switch, etc.) port mapping, so as to achieve the same domain name resolution to the dynamic extrnet IP.

Usage:
```
ls extranetip-to-dnsserver

# ---------- Client(Company hosts side) -------
pip3 install flask
startClientCommand='nohup /usr/bin/client.py > /dev/null 2>&1 &'
echo $startClientCommand >> /etc/rc.local # CentOS7
`$startClientCommand`

# --------- Server(CoreDNS side) --------
pip3 install flask
startServerCommand='nohup /usr/bin/server.py > /dev/null 2>&1 &'
echo $startServerCommand >> /etc/rc.local # CentOS7
`$startServerCommand`
```
