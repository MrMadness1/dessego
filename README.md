# DeSSE Go [![License](http://img.shields.io/badge/license-mit-blue.svg)](https://raw.githubusercontent.com/danmrichards/dessego/master/LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/danmrichards/dessego)](https://goreportcard.com/report/github.com/danmrichards/dessego)
A Demon's Souls server emulator implemented in Go and using SQLite for persistent
data storage.

## Acknowledgements
This is heavily based upon the work done by ymgve in the [desse][1] project. This
was the first working version of a Demon's Souls server after the official
shutdown. We're in the debt of the packet analysis and efforts from ymgve.

## Features
This server implements the core feature set for Demon's Souls:

* Login
* Character creation
* World tendency
* Blood messages
* Wandering ghosts
* Blood stains
* Summons

It should be noted that the full summon/multiplayer flow is not handled by this
server and relies on the Sony Playstation Network matchmaking system. There
is every chance they'll drop support for PS3 Demon's Souls at some point.

## Requirements
* [Go][2] 1.13+
* Docker with Compose (recommended for server deployments)

## Installation
```bash
$ go get -u github.com/danmrichards/dessego/cmd/server/...
```

## Building From Source
Clone this repo and build the binary:

```bash
$ make build
```

## Usage
```bash
Usage of ./bin/dessego-linux-amd64:
  -seed
        Seed database tables with legacy data
  -host string
        Public host advertised to game clients (default "127.0.0.1")
  -db string
        SQLite database path (default "./db/dessego.db")
```

The server can also be configured with environment variables:

| Variable | Default | Purpose |
| --- | --- | --- |
| `DESSEGO_PUBLIC_HOST` | `127.0.0.1` | Hostname or IPv4 address advertised to game clients |
| `DESSEGO_BOOTSTRAP_PORT` | `18000` | Bootstrap server TCP port |
| `DESSEGO_US_PORT` | `18666` | US game server TCP port |
| `DESSEGO_EU_PORT` | `18667` | EU game server TCP port |
| `DESSEGO_JP_PORT` | `18668` | JP game server TCP port |
| `DESSEGO_DB_PATH` | `./db/dessego.db` | SQLite database path |

Command-line flags override the matching environment variables.

## Docker Compose

Copy `compose.example.yaml`, set `DESSEGO_PUBLIC_HOST` to the public hostname or
IPv4 address that clients can reach, and start the service:

```bash
docker compose up -d --build
```

The example uses host networking because the bootstrap response advertises the
game ports directly. It persists the SQLite database in the `dessego_data`
volume and restarts the container unless it is explicitly stopped.

Expose these ports to the host running the container:

| Protocol | Port | Service |
| --- | ---: | --- |
| TCP | 18000 | Bootstrap |
| TCP | 18666 | US game service |
| TCP | 18667 | EU game service |
| TCP | 18668 | JP game service |

## Connecting from Demon's Souls
### Native PS3
To start with you'll need some sort of DNS proxy where you can configure the following URLs to route to your dessego server:

```
ds-eu-c.scej-online.jp
ds-eu-g.scej-online.jp
c.demons-souls.com
g.demons-souls.com
cmnap.scej-online.jp
demons-souls.scej-online.jp
```

Then you can follow these steps:
1. On your PS3 Navigate to the following menu: `Settings > Network > Internet Connection Settings > Custom > Enter Manually -> Scroll to DNS Section -> Manual`
2. Set Primary DNS to the host/port of your DNS proxy

### RPCS3
First ensure you have [RPCS3][3] installed and Demon's Souls is working (figure that one out yourselves!)

1. Open RPCS3 & Create a custom configuration of your game, proceed into the Network settings and set the following options as said:
2. Set Network Status to Connected
3. Set PSN Status to RPCN
4. Set DNS to `8.8.8.8`
5. Set IP/Host Switch to `ds-eu-c.scej-online.jp=<DESSEGO_IP>&&ds-eu-g.scej-online.jp=<DESSEGO_IP>&&c.demons-souls.com=<DESSEGO_IP>&&g.demons-souls.com=<DESSEGO_IP>&&cmnap.scej-online.jp=<DESSEGO_IP>&&demons-souls.scej-online.jp=<DESSEGO_IP>` where `DESSEGO_IP` is the host/port where you're running dessego.
6. Save Configuration
7. Go to RPCS3 main menu, proceed to 'Configuration', then 'RPCN'
8. Set Host to `np.rpcs3.net`
9. Set NPID to your preferred username
10. Set Password to your preferred password for RPCN
11. Click Create Account
12. You will be asked to enter you email, to which you will receive a email with a token inside it
13. In the Token Field, enter the token you received in the email.
14. Save, and Start the game

[1]: https://github.com/ymgve/desse
[2]: https://go.dev
[3]: https://rpcs3.net
