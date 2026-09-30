FROM golang:1.24-bookworm AS build

WORKDIR /src
COPY go.mod go.sum ./
COPY vendor ./vendor
COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build \
    -mod=vendor \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/dessego \
    ./cmd/server

FROM debian:bookworm-slim

RUN groupadd --system --gid 10001 dessego \
    && useradd --system --uid 10001 --gid 10001 --home-dir /app dessego \
    && install -d -o 10001 -g 10001 -m 0750 /app /data

WORKDIR /app

COPY --from=build /out/dessego /usr/local/bin/dessego
COPY --chown=10001:10001 internal ./internal

USER 10001:10001

VOLUME ["/data"]
EXPOSE 18000/tcp 18666/tcp 18667/tcp 18668/tcp

STOPSIGNAL SIGTERM
ENTRYPOINT ["/usr/local/bin/dessego"]
CMD ["-seed"]
