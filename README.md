# decred-proxy

Decred mining proxy with web-interface.

**Proxy feature list:**

* Rigs availability monitoring
* Keep track of accepts, rejects, blocks stats
* Easy detection of sick rigs
* Daemon failover list

<!---
![Demo](https://raw.githubusercontent.com/bitbandi/decred-proxy/master/proxy.png)
--->

### Building on Linux

Dependencies:

  * go >= 1.4
  * geth

Export GOPATH:

    export GOPATH=$HOME/go

Install required packages:

    go get github.com/decred/dcrd/blockchain
    go get github.com/decred/dcrd/chaincfg/chainhash
    go get github.com/goji/httpauth
    go get github.com/gorilla/mux
    go get github.com/yvasiyarov/gorelic

Compile:

    go build -o decred-proxy main.go

### Building on Windows

Install required packages (look at Linux install guide above). Then compile:

    go build -o decred-proxy.exe main.go

### Building on Mac OS X

If you didn't install [Brew](http://brew.sh/), do it. Then install Golang:

    brew install go

And follow Linux installation instructions because they are the same for OS X.

### Configuration

Configuration is self-describing, just copy *config.example.json* to *config.json* and specify endpoint URL and upstream URLs.

#### Example upstream section

```javascript
"upstream": [
  {
    "pool": true,
    "name": "Suprnova",
    "url": "http://dcr.suprnova.cc:9110",
    "username": "workername",
    "password": "x",
    "timeout": "10s"
  },
  {
    "name": "backup-decred",
    "url": "http://127.0.0.1:9109",
    "username": "yourusername",
    "password": "yoursecurepassword",
    "timeout": "10s"
  }
],
```

In this example we specified [Suprnova's Decred Pool](https://dcr.suprnova.cc/) mining pool as main mining target and a local geth node as backup for solo.

#### Running

    ./decred-proxy config.json

#### Mining

    ethminer -F http://x.x.x.x:8546/miner/5/gpu-rig -G
    ethminer -F http://x.x.x.x:8546/miner/0.1/cpu-rig -C

### Pools that work with this proxy

* [MaxMiners Pool](https://dcr.maxminers.net) Pool for decred
* [SuprNova.cc](https://dcr.suprnova.cc) Suprnova's Decred Pool

Pool owners, apply for listing here. PM me for implementation details.

### TODO

**Currently it's solo-only solution.**

* Report block numbers
* Report luck per rig
* Maybe add more stats
* Maybe add charts

### Donations

* **DCR**: [DsWybs2sPWpUzTyATFKBS8ciQn9vMAN7G6n](https://mainnet.decred.org/address/DsWybs2sPWpUzTyATFKBS8ciQn9vMAN7G6n)

* **BTC**: [1EttvCv3JrorDXZFtq1JcswtjXKNQ6YtmH](https://blockchain.info/address/1EttvCv3JrorDXZFtq1JcswtjXKNQ6YtmH)

### License

The MIT License (MIT).
