package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/rs/zerolog"
)

func main() {
	logger := createRootLogger()
	logger.Info().Msg("Hello, World!")
}

func createRootLogger() zerolog.Logger {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	var level = zerolog.DebugLevel
	fmt.Printf("Log level: %v\n", level)
	return logger.Level(level).With().Timestamp().Logger()
}

type routeMap map[string]string
type routeDefinition struct {
	headerSelector  string
	routeBySelector routeMap
}

type simpleRouter struct {
	proxies        map[string]*httputil.ReverseProxy
	regexes        map[string]*regexp.Regexp
	headerSelector string
}

func (sr simpleRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	selectorHeader := r.Header.Get(sr.headerSelector)

	for key := range sr.getMatchingProxyKeys(selectorHeader) {
		proxy := sr.proxies[key]
		proxy.ServeHTTP(w, r)
	}
}

func (sr simpleRouter) getMatchingProxyKeys(requestedSelector string) map[string]struct{} {
	matches := map[string]struct{}{}
	for proxyKey := range sr.proxies {
		rgexp := sr.regexes[proxyKey]
		if rgexp.MatchString(requestedSelector) {
			matches[proxyKey] = struct{}{}
		}
	}
	return matches
}

func createRouter(rtDefinition routeDefinition) simpleRouter {
	simpleRtr := simpleRouter{
		proxies:        make(map[string]*httputil.ReverseProxy),
		regexes:        make(map[string]*regexp.Regexp),
		headerSelector: rtDefinition.headerSelector,
	}
	for selector, target := range rtDefinition.routeBySelector {
		targetURL, err := url.Parse(target)
		if err != nil {
			panic(err)
		}
		simpleRtr.proxies[selector] = httputil.NewSingleHostReverseProxy(targetURL)
		simpleRtr.regexes[selector] = regexp.MustCompile(selector)
	}

	return simpleRtr
}
