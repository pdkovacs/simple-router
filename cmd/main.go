package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	simple_router "simple-router/internal"
)

func main() {
	logger := simple_router.CreateRootLogger()
	config, parseOptsErr := simple_router.ReadConfiguration(os.Args)
	if parseOptsErr != nil {
		fmt.Fprintf(os.Stderr, "%v\n", parseOptsErr)
		os.Exit(1)
	}

	router, routerCreErr := simple_router.NewRouter(config.RouteDef, logger)
	if routerCreErr != nil {
		fmt.Fprintf(os.Stderr, "%v\n", routerCreErr)
		os.Exit(1)
	}

	s := &http.Server{
		Addr:    config.ListenAt,
		Handler: &router,
	}
	log.Fatal(s.ListenAndServe())
}
