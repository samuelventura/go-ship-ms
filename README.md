# go-ship-ms

## Test Drive

```bash
#development
ln -sf ~/github/go-ship-ms/id_rsa.key ~/go/bin/go-ship-ms.key
go install && ~/go/bin/go-ship-ms
#daemon deploy
curl -X POST http://127.0.0.1:31600/api/daemon/env/ship -H "DaemonEnviron: SHIP_NAME=demo" -H "DaemonEnviron: SHIP_DOCK_POOL=127.0.0.1:31622" -H "DaemonEnviron: SHIP_DOCK_KEYPATH=$HOME/github/go-ship-ms/id_rsa.key"
curl -X POST http://127.0.0.1:31600/api/daemon/stop/ship
curl -X POST http://127.0.0.1:31600/api/daemon/start/ship
curl -X GET http://127.0.0.1:31600/api/daemon/env/ship
```

## Helpers

```bash
#manually check DNS records
dig dock.domain.tld TXT
```
