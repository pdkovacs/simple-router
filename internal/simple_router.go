package simple_router

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
)

type RouteMap map[string]string
type RouteDefinition struct {
	HeaderSelector  string
	RouteBySelector RouteMap
}

type SimpleRouter struct {
	proxies        map[string]*httputil.ReverseProxy
	regexes        map[string]*regexp.Regexp
	headerSelector string
}

func (sr *SimpleRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	selectorHeader := r.Header.Get(sr.headerSelector)

	for key := range sr.getMatchingProxyKeys(selectorHeader) {
		proxy := sr.proxies[key]
		proxy.ServeHTTP(w, r)
	}
}

func (sr *SimpleRouter) getMatchingProxyKeys(requestedSelector string) map[string]struct{} {
	matches := map[string]struct{}{}
	for proxyKey := range sr.proxies {
		rgexp := sr.regexes[proxyKey]
		if rgexp.MatchString(requestedSelector) {
			matches[proxyKey] = struct{}{}
		}
	}
	return matches
}

func NewRouter(rtDefinition RouteDefinition) SimpleRouter {
	simpleRtr := SimpleRouter{
		proxies:        make(map[string]*httputil.ReverseProxy),
		regexes:        make(map[string]*regexp.Regexp),
		headerSelector: rtDefinition.HeaderSelector,
	}
	for selector, target := range rtDefinition.RouteBySelector {
		targetURL, err := url.Parse(target)
		if err != nil {
			panic(err)
		}
		simpleRtr.proxies[selector] = httputil.NewSingleHostReverseProxy(targetURL)
		simpleRtr.regexes[selector] = regexp.MustCompile(selector)
	}

	return simpleRtr
}
