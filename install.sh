#!/bin/bash -x

#curl -X POST http://127.0.0.1:31600/api/daemon/stop/ship
#curl -X POST http://127.0.0.1:31600/api/daemon/disable/ship
#curl -X POST http://127.0.0.1:31600/api/daemon/uninstall/ship
#curl -X POST http://127.0.0.1:31600/api/daemon/env/$DAEMON \
#     -H "DaemonEnviron: SHIP_DOCK_KEYPATH=$HOME/.ssh/id_rsa"
if [[ "$OSTYPE" == "linux"* ]]; then
    DAEMON=ship
    BINARY=go-$DAEMON-ms
    SRC=$HOME/go/bin
    DST=/usr/local/bin
    go install
    curl -X POST http://127.0.0.1:31600/api/daemon/stop/$DAEMON
    curl -X POST http://127.0.0.1:31600/api/daemon/uninstall/$DAEMON
    sudo cp $SRC/$BINARY $DST
    curl -X POST http://127.0.0.1:31600/api/daemon/install/$DAEMON?path=$DST/$BINARY
    curl -X POST http://127.0.0.1:31600/api/daemon/enable/$DAEMON
    curl -X POST http://127.0.0.1:31600/api/daemon/start/$DAEMON
    curl -X GET http://127.0.0.1:31600/api/daemon/info/$DAEMON
    curl -X GET http://127.0.0.1:31600/api/daemon/env/$DAEMON
fi
