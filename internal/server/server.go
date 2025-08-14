package server

import "regexp"

type router struct {
	pattern *regexp.Regexp
	handler Handler
}

type ServerMux struct {
	routers []router
}

func NewServerMux() *ServerMux {
	return &ServerMux{}
}
