package server

type Response interface {
	StatusCode(int)
	Header(string, string)
	Body([]byte)
	Write([]byte) (int, error)
	Send()
	Flush() error
}
