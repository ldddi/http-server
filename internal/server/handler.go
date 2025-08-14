package server

type Handler interface {
	Serve(Response, *Request)
}
