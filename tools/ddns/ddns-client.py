#!/usr/bin/python3
#coding=utf-8
import sys
import os
import threading
import datetime as dt
import hashlib
import random
from timeit import Timer
from flask import Flask, jsonify
from flask import json
import http.client,urllib.parse

# Put this into the server where the public IP will change.
# 也可直接使用Linux cron调度
def updateDNS(): 
  try:
    # Hasing sign
    key = 'au43hwe9dfkl'
    r = random.random()
    orgin = "%s%s" % (r,key)
    hl = hashlib.md5()
    hl.update(orgin.encode(encoding='utf-8'))
    sign = hl.hexdigest()

    # Build parameters
    url = '/dns/update?sign=%s&r=%s' % (sign,r)
    jsonData = json.dumps("{}")

    conn = http.client.HTTPConnection("127.0.0.2", 4008)
    conn.request('POST', url, jsonData, {'Content-Type':'application/json'})
    res = conn.getresponse()
    resData = res.read().decode('utf-8')
    conn.close()

    print(dt.datetime.now().strftime("%Y-%m-%d %H:%M:%S")+" - DNS update request : "+url+", response status: "+ str(res.status) +", data: "+resData)
  finally:
    global timer
    delay = random.random()*1000+2000
    timer = threading.Timer(delay, updateDNS)
    timer.start()

if __name__ == "__main__":
    updateDNS()