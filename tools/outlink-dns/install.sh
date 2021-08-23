#!/bin/bash
# Copyright 2017 ~ 2025 the original author or authors<Wanglsir@gmail.com, 983708408@qq.com>.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Global definition.
CURR_DIR=$(cd "`dirname $0`"/;pwd)
OUTLINK_LOG_DIR='/mnt/disk1/log/outlink-dns'; mkdir -p $OUTLINK_LOG_DIR

# Installation to client. (e.g: Company intranet hosts client side)
function installClient() {
    cd $CURR_DIR
    local useBinary=$1
    local installFile='/bin/outlink-dns-client'
    # Compiling install.
    pip3 install flask
    case $useBinary in
        -b|--binary)
            pip3 install pyinstaller
            pyinstaller -F outlink-dns-client.py
            sudo cp -r "dist/outlink-dns-client" $installFile
            sudo rm -rf __pycache__ build dist *.spec # Clean
        ;;
        *)
            sudo cp -r "outlink-dns-client.py" $installFile
        ;;
    esac
    local startCmd="nohup $installFile > $OUTLINK_LOG_DIR/client.out 2>&1 &"
    sudo echo $startCmd >> /etc/rc.local # CentOS7
    echo "Starting for outlink client ..."
    bash -c "$startCmd"
    echo "Installed outlink client to $installFile successfully !"
}

# Installation to server(e.g CoreDNS extranet server side)
function installServer() {
    cd $CURR_DIR
    local useBinary=$1
    local installFile='/bin/outlink-dns-server'
    # Compiling install.
    pip3 install flask
    pip3 install redis-py-cluster
    case $useBinary in
        -b|--binary)
            pip3 install pyinstaller
            pyinstaller -F outlink-dns-server.py
            sudo cp -r "dist/outlink-dns-server" $installFile
            sudo rm -rf __pycache__ build dist *.spec # Clean
        ;;
        *)
            sudo cp -r "outlink-dns-server.py" $installFile
        ;;
    esac
    local startCmd="nohup $installFile > $OUTLINK_LOG_DIR/server.out 2>&1 &"
    sudo echo $startCmd >> /etc/rc.local # CentOS7
    echo "Starting for outlink server ..."
    bash -c "$startCmd"
    echo "Installed outlink server to $installFile successfully !"
}

# --- Main entries. ---
installType=$1
useBinary=$2
case $installType in
  -c|--client)
    installClient "$useBinary"
  ;;
  -s|--server)
    installServer "$useBinary"
  ;;
  -h|--help|*)
    echo "Usage: {./$(basename $0) [-c|--client [-b|--binary]] or [-s|--server [-b|--binary]]}"
  ;;
esac