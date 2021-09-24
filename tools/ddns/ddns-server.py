#!/usr/bin/python3
#coding=utf-8
import sys
import os
import hashlib
from flask import Flask, jsonify, request
from flask import json
from rediscluster import RedisCluster

# This script is run into the docker container running bind.
app = Flask(__name__)

@app.route('/')
def index():
    return "Wecome to DoPaas-CoreDNS for DDNS!"

@app.route('/dns/update', methods=['POST'])
def updateDns():
    # Get parameters
    sign = request.args.get('sign')
    r = request.args.get('r')
    #ipaddr = request.args.get("ipaddr")
    ipaddr = request.remote_addr

    # Check signture
    key = 'au43hwe9dfkl'
    orgin = "%s%s" % (r,key)
    hl = hashlib.md5()
    hl.update(orgin.encode(encoding='utf-8'))
    _sign = hl.hexdigest()

    if sign != _sign:
        print('Calculation signture: %s, request singture: %s' % (sign, _sign))
        return jsonify({"code":"Illegal signture"})

    print("DNS update for : "+ ipaddr)
    #os.system("%s %s"%("/root/dns-update-tool.sh", ipaddr))

    redisClient = RedisCluster(startup_nodes=[
        {"host": "127.0.0.1", "port": 6369},
        {"host": "127.0.0.1", "port": 6380},
        {"host": "127.0.0.1", "port": 6381},
        {"host": "127.0.0.1", "port": 7379},
        {"host": "127.0.0.1", "port": 7380},
        {"host": "127.0.0.1", "port": 7381}],
        password='zzx!@#$%')
    print("Saving DNS resolve ipaddr for : " + ipaddr)
    zoneJson = "{\"a\":[{\"ttl\":600, \"ip\":\"%s\"}]}" % (ipaddr)
    redisClient.hset("_coredns:anjiancloud.owner.", "*", zoneJson)
    redisClient.hset("_coredns:anjiancloud.owner.", "@", zoneJson)
    redisClient.connection_pool.disconnect()
    print("Closed redis cluster connection pool for - " + str(redisClient))

    return jsonify({"code":"ok","ipaddr":ipaddr})

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=4008)