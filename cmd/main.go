package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	simple_router "simple-router/internal"

	"github.com/rs/zerolog"
)

func main() {
	logger := createRootLogger()
	config, parseOptsErr := simple_router.ReadConfiguration(os.Args)
	if parseOptsErr != nil {
		fmt.Fprintf(os.Stderr, "%v\n", parseOptsErr)
		os.Exit(1)
	}
	logger.Info().Msgf("Config: %#v", config)

	router := simple_router.NewRouter(config.RouteDef)
	s := &http.Server{
		Addr:    config.ListenAt,
		Handler: &router,
	}
	log.Fatal(s.ListenAndServe())
}

func createRootLogger() zerolog.Logger {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	var level = zerolog.DebugLevel
	return logger.Level(level).With().Timestamp().Logger()
}
