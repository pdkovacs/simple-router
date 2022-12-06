package simple_router

import (
	"fmt"
	"strings"

	flags "github.com/jessevdk/go-flags"
)

type Options struct {
	ListenAt       string   `short:"l" long:"listen-at" default:"localhost:0" description:"The address:port to listen at"`
	RouteStrings   []string `short:"r" long:"route" description:"Route definition: <selector-pattern>;<target>" required:"true"`
	HeaderSelector string   `short:"h" long:"header" default:"Authorization" description:"The header containing the descriminator onto which the routing will be applied"`
}

type Configuration struct {
	RouteDef RouteDefinition
	ListenAt string
}

func ReadConfiguration(args []string) (Configuration, error) {
	config := Configuration{}

	opts := Options{}
	p := flags.NewParser(&opts, 0)
	args, err := p.ParseArgs(args)
	if flagsError, ok := err.(*flags.Error); ok && flagsError.Type == flags.ErrRequired {
		return config, err
	}

	if err != nil {
		return config, fmt.Errorf("failed to parse command line arguments: %w", err)
	}
	if len(args) > 1 {
		return config, fmt.Errorf("unexpected arguments: %v", strings.Join(args, ", "))
	}

	routeMap := make(RouteMap)
	for _, routeDef := range opts.RouteStrings {
		targetStart := strings.LastIndex(routeDef, ";")
		selector := routeDef[:targetStart]
		target := routeDef[targetStart+1:]
		routeMap[selector] = string(target)
	}

	routeDefinition := RouteDefinition{}
	routeDefinition.RouteBySelector = routeMap
	routeDefinition.HeaderSelector = opts.HeaderSelector

	config.RouteDef = routeDefinition
	config.ListenAt = opts.ListenAt

	return config, nil
}
