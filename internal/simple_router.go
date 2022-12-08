package simple_router

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"

	"github.com/rs/zerolog"
)

type SelectorType int

const (
	SelectorTypeNone SelectorType = iota
	SelectorTypeHeader
	SelectorTypeCookie
)

type RouteSelector struct {
	Name string
	Type SelectorType
}

type RouteMap map[string]string
type RouteDefinition struct {
	Selector        RouteSelector
	RouteBySelector RouteMap
	DefaultTarget   string
}

type reverseProxyInfo struct {
	target   string
	revProxy *httputil.ReverseProxy
}
type SimpleRouter struct {
	proxies      map[string]*reverseProxyInfo
	defaultProxy *reverseProxyInfo
	regexes      map[string]*regexp.Regexp
	selector     RouteSelector
	logger       zerolog.Logger
}

func (sr *SimpleRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	selectorValue, err := sr.getSelectorValue(r)
	if err != nil {
		sr.logger.Error().Msgf("failed to get selector value: %v", err)
		w.WriteHeader(500)
		return
	}
	selectorValue, err = url.QueryUnescape(selectorValue)
	if err != nil {
		sr.logger.Error().Msgf("failed to get selector value: %v", err)
		w.WriteHeader(500)
		return
	}
	sr.logger.Debug().Msgf("[%s] received selector %s to serve", sr.selector.Name, selectorValue)

	matchingProxyInfos := []*reverseProxyInfo{}
	for _, key := range sr.getMatchingProxyKeys(selectorValue) {
		matchingProxyInfos = append(matchingProxyInfos, sr.proxies[key])
	}

	if len(matchingProxyInfos) == 0 {
		matchingProxyInfos = append(matchingProxyInfos, sr.defaultProxy)
	}

	for _, proxyInfo := range matchingProxyInfos {
		sr.logger.Debug().Msgf("sending %s to target %s...", selectorValue, proxyInfo.target)
		proxyInfo.revProxy.ServeHTTP(w, r)
	}
}

func (sr *SimpleRouter) getSelectorValue(r *http.Request) (string, error) {
	switch sr.selector.Type {
	case SelectorTypeHeader:
		return r.Header.Get(sr.selector.Name), nil
	case SelectorTypeCookie:
		cookie, err := r.Cookie(sr.selector.Name)
		if err == http.ErrNoCookie {
			return "", nil
		}
		return cookie.Value, nil
	default:
		return "", fmt.Errorf("no such selector type: %v", sr.selector.Type)
	}
}

func (sr *SimpleRouter) getMatchingProxyKeys(requestedSelector string) []string {
	matches := []string{}
	for proxyKey := range sr.proxies {
		rgexp := sr.regexes[proxyKey]
		if rgexp.MatchString(requestedSelector) {
			matches = append(matches, proxyKey)
			continue
		}
	}
	if len(matches) > 0 {
		return matches
	}
	return []string{}
}

func NewRouter(rtDefinition RouteDefinition, logger zerolog.Logger) (SimpleRouter, error) {
	simpleRtr := SimpleRouter{
		proxies:  make(map[string]*reverseProxyInfo),
		regexes:  make(map[string]*regexp.Regexp),
		selector: rtDefinition.Selector,
		logger:   logger,
	}

	if rtDefinition.Selector.Type == SelectorTypeNone {
		return simpleRtr, fmt.Errorf("missing selector type")
	}

	if rtDefinition.Selector.Name == "" {
		return simpleRtr, fmt.Errorf("missing selector name")
	}

	for selector, target := range rtDefinition.RouteBySelector {
		targetURL, err := url.Parse(target)
		if err != nil {
			return simpleRtr, fmt.Errorf("failed to parse URL %s: %w", target, err)
		}
		simpleRtr.proxies[selector] = &reverseProxyInfo{
			target:   target,
			revProxy: httputil.NewSingleHostReverseProxy(targetURL),
		}
		simpleRtr.regexes[selector] = regexp.MustCompile(selector)
	}

	for _, proxy := range simpleRtr.proxies {
		if proxy.target == rtDefinition.DefaultTarget {
			simpleRtr.defaultProxy = proxy
		}
	}

	if simpleRtr.defaultProxy == nil {
		targetURL, err := url.Parse(rtDefinition.DefaultTarget)
		if err != nil {
			return simpleRtr, fmt.Errorf("failed to parse URL %s: %w", rtDefinition.DefaultTarget, err)
		}
		simpleRtr.defaultProxy = &reverseProxyInfo{
			target:   rtDefinition.DefaultTarget,
			revProxy: httputil.NewSingleHostReverseProxy(targetURL),
		}
	}

	return simpleRtr, nil
}
