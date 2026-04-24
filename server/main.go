package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"m1k1o/neko/internal/config"
	"m1k1o/neko/internal/server"
)

var (
	// Version is set at build time via ldflags
	Version = "dev"
	// Commit is set at build time via ldflags
	Commit = "unknown"
	// BuildDate is set at build time via ldflags
	BuildDate = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "neko",
	Short: "neko - self hosted virtual browser",
	Long: `neko is a self hosted virtual browser that runs in docker and uses
WebRTC to stream the desktop to connected users.`,
	Run: run,
}

func init() {
	// Configure zerolog to use console writer for human-readable output
	// Also include timestamps so it's easier to correlate logs with events
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})

	// Bind configuration flags
	config.Bind(rootCmd)
}

func run(cmd *cobra.Command, args []string) {
	// Initialize configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load configuration")
	}

	// Set global log level
	zerolog.SetGlobalLevel(cfg.LogLevel())

	log.Info().
		Str("version", Version).
		Str("commit", Commit).
		Str("build_date", BuildDate).
		Msg("starting neko server")

	// Create and start the server
	neko, err := server.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create server")
	}

	if err := neko.Start(); err != nil {
		log.Fatal().Err(err).Msg("failed to start server")
	}

	log.Info().Str("bind", cfg.Bind).Msg("neko server is running")

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down neko server")

	if err := neko.Shutdown(); err != nil {
		log.Error().Err(err).Msg("error during server shutdown")
	}

	log.Info().Msg("neko server stopped")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
