// Demon's Souls API
//
// The boostrap and game API endpoints for Demon's Souls
//
//     Schemes: http
//	   Version: 1.0.0
//     basePath: /
//
//     Consumes:
//     - text/plain
//
//     Produces:
//     - text/plain
//
// swagger:meta
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/danmrichards/dessego/internal/crypto"
	"github.com/danmrichards/dessego/internal/database"
	"github.com/danmrichards/dessego/internal/server/bootstrap"
	"github.com/danmrichards/dessego/internal/server/game"
	"github.com/danmrichards/dessego/internal/service/character"
	"github.com/danmrichards/dessego/internal/service/gamestate"
	"github.com/danmrichards/dessego/internal/service/ghost"
	"github.com/danmrichards/dessego/internal/service/msg"
	"github.com/danmrichards/dessego/internal/service/replay"
	"github.com/danmrichards/dessego/internal/service/sos"
	"github.com/rs/zerolog"
)

const (
	defaultHostGame      = "127.0.0.1"
	defaultPortBootstrap = "18000"
	defaultPortUS        = "18666"
	defaultPortEU        = "18667"
	defaultPortJP        = "18668"
	defaultDBPath        = "./db/dessego.db"
)

var (
	hostGame      string
	portBootstrap string
	portUS        string
	portEU        string
	portJP        string
	dbPath        string
	seed          bool
)

func main() {
	flag.BoolVar(&seed, "seed", false, "Seed database tables with legacy data")
	flag.StringVar(&hostGame, "host", envOrDefault("DESSEGO_PUBLIC_HOST", defaultHostGame), "Public host advertised to game clients")
	flag.StringVar(&portBootstrap, "bootstrap-port", envOrDefault("DESSEGO_BOOTSTRAP_PORT", defaultPortBootstrap), "Bootstrap server TCP port")
	flag.StringVar(&portUS, "us-port", envOrDefault("DESSEGO_US_PORT", defaultPortUS), "US game server TCP port")
	flag.StringVar(&portEU, "eu-port", envOrDefault("DESSEGO_EU_PORT", defaultPortEU), "EU game server TCP port")
	flag.StringVar(&portJP, "jp-port", envOrDefault("DESSEGO_JP_PORT", defaultPortJP), "JP game server TCP port")
	flag.StringVar(&dbPath, "db", envOrDefault("DESSEGO_DB_PATH", defaultDBPath), "SQLite database path")
	flag.Parse()

	l := zerolog.New(os.Stdout)
	gameServers := map[string]string{
		"US": portUS,
		"EU": portEU,
		"JP": portJP,
	}

	db, err := database.NewSQLite(dbPath)
	if err != nil {
		fatal(l, err)
	}
	defer db.Close()

	// Track the servers, so we can close them down later.
	servers := make([]io.Closer, 0, 4)
	serverErrs := make(chan error, 4)

	// Bootstrap server; used to allow Demon's Souls to configure it's network
	// client.
	var bs *bootstrap.Server
	bs, err = bootstrap.NewServer(portBootstrap, hostGame, gameServers, l)
	if err != nil {
		fatal(l, err)
	}
	servers = append(servers, bs)

	l.Info().Msg("bootstrap server listening on " + portBootstrap)
	go func() {
		if serveErr := bs.Serve(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			serverErrs <- fmt.Errorf("bootstrap server: %w", serveErr)
		}
	}()

	// Dependencies for the gamestate server.
	rd, err := crypto.NewDecrypter(crypto.DefaultAESKey)
	if err != nil {
		fatal(l, err)
	}

	c, err := character.NewSQLiteService(db)
	if err != nil {
		fatal(l, err)
	}

	var mo []msg.Option
	if seed {
		mo = append(mo, msg.Seed())
	}
	ms, err := msg.NewSQLiteService(db, l, mo...)
	if err != nil {
		fatal(l, err)
	}

	var ro []replay.Option
	if seed {
		ro = append(ro, replay.Seed())
	}
	rs, err := replay.NewSQLiteService(db, l, ro...)
	if err != nil {
		fatal(l, err)
	}

	// Create a gamestate server for each supported region
	for region, port := range gameServers {
		gs, err := game.NewServer(
			port,
			rd,
			c,
			gamestate.NewMemory(),
			ms,
			ghost.NewMemory(l),
			rs,
			sos.NewManager(l),
			l,
		)
		if err != nil {
			fatal(l, err)
		}
		servers = append(servers, gs)

		l.Info().Msg(region + " game server listening on " + port)
		go func(region string) {
			if serveErr := gs.Serve(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
				serverErrs <- fmt.Errorf("%s game server: %w", region, serveErr)
			}
		}(region)
	}

	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	select {
	case sig := <-sigChan:
		l.Info().Str("signal", sig.String()).Msg("shutting down servers")
	case serveErr := <-serverErrs:
		fatal(l, serveErr)
	}

	for _, s := range servers {
		if err = s.Close(); err != nil {
			l.Error().Err(err).Msg("close server")
		}
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func fatal(l zerolog.Logger, err error) {
	l.Fatal().Err(err).Msg("fatal error")
}
