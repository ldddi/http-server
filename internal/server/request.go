package server

type Request struct {
	Headers  map[string]string
	Method   string
	Params   map[string]string
	Query    map[string]string
	Path     string
	Body     []byte
	Protocol string
}
