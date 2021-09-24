#!/usr/bin/python3
#coding=utf-8
from re import split
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

@app.route('/ddns/update', methods=['POST'])
def updateDNS():
    # Get parameters
    sign = request.args.get('sign')
    r = request.args.get('r')
    ipaddr = request.remote_addr

    # Check signture
    key = os.getenv('COREDNS_DDNS_KEY', 'abcdefghijklmnopqrstuvwxyz')
    orgin = "%s%s" % (r,key)
    hl = hashlib.md5()
    hl.update(orgin.encode(encoding='utf-8'))
    _sign = hl.hexdigest()

    if sign != _sign:
        print('Calculation signture: %s, request singture: %s' % (sign, _sign))
        return jsonify({"code":"Illegal signture"})

    print("DNS update for : "+ ipaddr)
    #os.system("%s %s"%("/bin/ddns-updater.sh", ipaddr))

    redisNodes = os.getenv('COREDNS_DDNS_REDIS_NODES', '127.0.0.1:6379,127.0.0.1:6380,127.0.0.1:6381,127.0.0.1:7379,127.0.0.1:7380,127.0.0.1:7381')
    redisPasswd = os.getenv('COREDNS_DDNS_REDIS_PASSWORD', '123456')
    objNodes = list(map(lambda hap: {"host":hap.split(':')[0],"port":hap.split(':')[1]}, redisNodes.split(',')))
    redisClient = RedisCluster(startup_nodes= objNodes, password=redisPasswd)
    print("Updating DNS resolve ipaddr for: %s" % (ipaddr))

    zoneJson = "{\"a\":[{\"ttl\":600, \"ip\":\"%s\"}]}" % (ipaddr)
    corednsPrefix = os.getenv('COREDNS_DDNS_PREFIX', '_coredns:')
    domain = os.getenv('COREDNS_DDNS_DOMAIN', 'example.com')
    redisClient.hset(corednsPrefix+domain+".", "*", zoneJson)
    redisClient.hset(corednsPrefix+domain+".", "@", zoneJson)
    redisClient.connection_pool.disconnect()
    print("Closed redis cluster connection pool for - %s" % (str(redisClient)))

    return jsonify({"code":"ok","ipaddr":ipaddr})

if __name__ == '__main__':
    app.run(host=os.getenv('COREDNS_DDNS_LISTEN_ADDR', '0.0.0.0'), port=os.getenv('COREDNS_DDNS_LISTEN_PORT', 4008))