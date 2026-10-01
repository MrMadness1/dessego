# DeSSE Go — MGN preservation fork

A Demon's Souls game-service emulator written in Go, with SQLite persistence.
This fork maintains the original protocol while improving deployment, native
PS3 compatibility, parser safety, and multiplayer-state reliability.

## Project status

As of October 1, 2026 (operator-reported gameplay results):

- Native US PS3 login, summon-sign visibility, and two-console co-op connection
  have been demonstrated on MGN.
- RPCS3 account validation through MGN RPCN and connection to the MGN game
  service have been demonstrated.
- Full-zone/boss completion, death, cancellation, helper return, and repeated
  summoning remain acceptance tests, not blanket guarantees.
- Native PS3 ↔ RPCS3 co-op is unverified and is a separate research track.
- EU/JP listeners exist; the US results do not establish equivalent regional
  gameplay coverage or cross-region matchmaking.

The game service is not a replacement for Sony PSN. It handles game-specific
requests and summon information, but does not implement the complete native
PS3 authentication, NP room creation, signaling, or peer networking stack.
Native PS3 sessions still depend on their platform services; RPCS3 uses RPCN.
Connecting both platforms to this backend does not by itself enable cross-play.

The September 30 reliability pass is merged as PR #5 (merge revision
`7d8e2d4004ba0ac43d975fb74891a2a2d489d090`): malformed-request and AES
validation, stale/competing summon handling, replay/message/ghost safeguards,
request bounds, and SQLite resource cleanup have regression coverage.
CI includes race tests, vet, and a container build. This is a working community
preservation project, not a claim of security audit or hostile-load readiness.

## MGN player quick start

These public endpoints serve different purposes:

| Endpoint | Purpose |
| --- | --- |
| `dns.mgn.pub` / `137.220.35.216` | Public preservation DNS on the VPS |
| `des.mgn.pub` | Demon's Souls game-service destination |
| `rpcn.mgn.pub` | MGN RPCN account and emulator matchmaking service |

The DNS address is not the game-server address. The public DNS node is separate
from the home-hosted game services. DNS records direct the game's original
service names to the game destination; do not substitute the DNS VPS address
for the game destination in manual host overrides.

### Native PS3

1. Open **Settings → Network Settings → Internet Connection Settings**.
2. Choose the custom setup path, retain your normal addressing settings, and
   select manual DNS.
3. Set **Primary DNS = 137.220.35.216**.
4. Do not configure an unrelated public resolver as a fallback and assume it
   will provide the preservation overrides. Ask the operator for a supported
   secondary resolver if required.
5. Complete setup, start Demon's Souls, and check for the MGN welcome message.

This redirects game-service discovery. It does not make a stock/native PS3
speak RPCN or remove its PSN requirements. Match game region/version and normal
summoning eligibility when testing multiplayer.

### RPCS3

1. Create/open the game's custom configuration and its Network settings.
2. Set **Network Status = Connected** and **PSN Status = RPCN**.
3. Set **DNS = 137.220.35.216** and leave **IP/Hosts switches empty**.
   MGN DNS resolves the game service names; manual game-server IP overrides
   are not needed for this verified setup.
4. Open RPCS3's RPCN account/server manager (network-services menu or RPCN
   toolbar entry, depending on your build).
5. Add/select the MGN server at **rpcn.mgn.pub** and create or select an account
   on that server. Official RPCN accounts are not automatically MGN accounts.
6. Follow the selected server's account/token prompts, then use **Test Account**.
7. Save the game configuration, launch, and check for the MGN welcome message.

Do not assume a fixed menu layout or that every private server sends email.
The official RPCN service is an alternative account/matchmaking service, not
the MGN server. All emulator players in a test should use the same RPCN service.

MGN's currently verified RPCN compatibility baseline is protocol **32**, pinned
at upstream revision `696f692355d4f1ecfea691ab85f42292d9b853b2` (binary version
1.10.0). Protocol-33 v1.11.0 produced a protocol mismatch with the tested client.
This is a dated compatibility record, not a permanent recommendation to avoid
new versions. Validate client/server protocol compatibility before upgrades.

## Generic self-hosting

### Requirements and source build

- Docker Engine with the Compose plugin is recommended for Linux deployments.
- Docker and CI use Go **1.24** as the tested toolchain. The historical
  `go 1.15` declaration in go.mod is not a guarantee that older toolchains are
  supported by this fork.
- Native builds require a C compiler and SQLite build prerequisites because
  the SQLite driver uses CGO; `make build` also requires Make.

Clone this fork rather than installing the historical upstream with go get:

~~~sh
git clone https://github.com/MrMadness1/dessego.git
cd dessego
make build
~~~

For a reproducible deployment, check out a reviewed commit before building.
Keep the original module/import paths; they preserve upstream source identity.

### Configuration

| Variable | Default | Purpose |
| --- | --- | --- |
| `DESSEGO_PUBLIC_HOST` | `127.0.0.1` | Reachable hostname or IPv4 advertised to clients |
| `DESSEGO_BOOTSTRAP_PORT` | `18000` | Bootstrap TCP listener |
| `DESSEGO_US_PORT` | `18666` | US game TCP listener |
| `DESSEGO_EU_PORT` | `18667` | EU game TCP listener |
| `DESSEGO_JP_PORT` | `18668` | JP game TCP listener |
| `DESSEGO_DB_PATH` | `./db/dessego.db` | Persistent SQLite path |

Flags `-host` and `-db` override their matching environment variables.
The `-seed` flag loads legacy data; do not run seeding against a production
player database without first reviewing its effects and taking a backup.

### Docker Compose

Review `compose.example.yaml`, set a reachable `DESSEGO_PUBLIC_HOST`, and run:

~~~sh
docker compose -f compose.example.yaml up -d --build
~~~

The example uses host networking and the `dessego_data` persistent volume.
The container runs non-root (UID/GID 10001); any bind-mounted data directory
must be writable by that identity. Do not delete the data volume during an
image update. The restart policy is `unless-stopped`.

The default game-service ports are TCP **18000, 18666, 18667, 18668**.
Expose only the listeners your deployment needs. These are not DNS or RPCN
ports. DNS cannot advertise a different TCP port through an A record; native
clients' bootstrap path and advertised game endpoints must remain reachable.
Do not move an existing web reverse proxy or blindly forward TCP 80/443:
confirm the native bootstrap access path for your deployment first.

### DNS and alternate emulator overrides

Configure these original game hostnames to resolve to your reachable game
service destination, not to your DNS resolver's address:

~~~text
ds-eu-c.scej-online.jp
ds-eu-g.scej-online.jp
c.demons-souls.com
g.demons-souls.com
cmnap.scej-online.jp
demons-souls.scej-online.jp
~~~

Native PS3 users enter your resolver's numeric IP in manual DNS settings.
For RPCS3, the six-name IP/Host Switch can be used as an alternate/debug path:

~~~text
ds-eu-c.scej-online.jp=<GAME_SERVER_IPV4>&&ds-eu-g.scej-online.jp=<GAME_SERVER_IPV4>&&c.demons-souls.com=<GAME_SERVER_IPV4>&&g.demons-souls.com=<GAME_SERVER_IPV4>&&cmnap.scej-online.jp=<GAME_SERVER_IPV4>&&demons-souls.scej-online.jp=<GAME_SERVER_IPV4>
~~~

Replace every placeholder with the game destination's numeric IPv4 address.
For MGN, obtain that destination from `des.mgn.pub`/operator guidance, not from
`dns.mgn.pub`. Do not append a URL scheme or port to the IPv4 substitution.
Confirm host-switch syntax against your RPCS3 build if using a nonstandard
port arrangement. Generic deployments choose their own compatible RPCN
service independently of these game-server DNS overrides.

## Operations and validation

- Record the source commit, image digest, deployment definition, toolchain,
  client versions, game region/version, and demonstrated acceptance tests.
- Stage upgrades separately; avoid unattended production rebuilds from main.
  RPCN protocol compatibility must be tested independently of game-server CI.
- Before updates, preserve SQLite data, configuration, and the previous image.
  Stop writes for a file-copy backup or use SQLite's backup mechanism; a raw
  copy of an actively written database is not a reliable backup strategy.
- Keep an off-server backup, verify SQLite integrity, and rehearse restoration.
  Keep diagnostic logs/captures private and redact identifiers and tokens.
- Verify startup/listeners, login, sign visibility, actual summoning, helper
  return, cancellation, death, zone/boss progression, and re-summoning. Repeat
  after reboot. A welcome screen is not proof of full multiplayer correctness.
- Monitor restart counts, disk usage, listener availability, and application
  failures. Bound diagnostic capture size and retention.

Persistent character/message/replay storage is shared; regional transient
managers are separate. A restart can clear transient session state even when
persistent player data survives. Plan maintenance accordingly.

## Known limitations and next work

Public-facing hardening still needs measured work: HTTP timeouts, controlled
unknown-route aborts, privacy-aware timestamped logs, graceful shutdown, parser
fuzzing, explicit SQLite concurrency settings, schema migrations, query/index
measurements, retention, and listener-aware health checks. Container hardening
and rate limits should be tested against real PS3 timing before promotion.
The protocol's fixed AES key is not client authentication; treat incoming
identifiers and payloads as untrusted.

Cross-play research should compare synchronized, narrowly filtered client and
server captures with application logs. Encrypted traffic and direct peer
connections mean server packet captures alone cannot explain every failure.
PS3 ↔ PS3 and RPCS3 ↔ RPCS3 are separate baselines before mixed-platform tests.
Sony-independent native PS3 matchmaking remains a separate preservation goal.
For other games, reuse the capture/testing/deployment methodology, not assumed
Demon's Souls keys, endpoints, or wire formats.

## Acknowledgements and license

Based on [danmrichards/dessego](https://github.com/danmrichards/dessego) and
ymgve's pioneering [DeSSE](https://github.com/ymgve/desse) protocol work.
Upstream authors' work and attribution are retained. See [LICENSE](LICENSE)
for the MIT license. [RPCS3](https://rpcs3.net) and
[RPCN](https://github.com/RipleyTom/rpcn) are independent upstream projects.
