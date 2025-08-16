package server

import (
	"fmt"
	"net"
	"strconv"
)

type ResponseWriter interface {
	StatusCode(int)
	Header(string, string)
	Body([]byte)
	Write([]byte) (int, error)
	Send()
	Flush() error
}

type simpleResponseWriter struct {
	conn       net.Conn
	headers    map[string]string
	body       []byte
	statusCode int
	headerSent bool
}

func (w *simpleResponseWriter) StatusCode(code int) {
	w.statusCode = code
}

func (w *simpleResponseWriter) Header(k, v string) {
	if w.headers == nil {
		w.headers = make(map[string]string)
	}
	w.headers[k] = v
}

func (w *simpleResponseWriter) Body(body []byte) {
	w.body = body
}

func (w *simpleResponseWriter) Write(data []byte) (int, error) {
	return w.conn.Write(data)
}

func (w *simpleResponseWriter) Flush() error {
	return nil
}

func (w *simpleResponseWriter) Send() {
	if w.statusCode == 0 {
		w.statusCode = 200
	}

	if w.headers == nil {
		w.headers = make(map[string]string)
	}

	if w.headers["Content-type"] == "" {
		w.headers["Content-type"] = "text/plain"
	}
	w.headers["Content-Length"] = strconv.Itoa(len(w.body))

	fmt.Fprintf(w.conn, "HTTP/1.1 %d %s\r\n", w.statusCode, statusTxt(w.statusCode))
	for k, v := range w.headers {
		fmt.Fprintf(w.conn, "%s: %s\r\n", k, v)
	}
	fmt.Fprint(w.conn, "\r\n")
	fmt.Fprint(w.conn, string(w.body))
}

func statusTxt(code int) string {
	switch code {
	case 200:
		return "OK"
	case 201:
		return "Created"
	case 400:
		return "Bad Request"
	case 404:
		return "No Found"
	case 405:
		return "Method Not Allowed"
	default:
		return "Unknown"
	}
}
