package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"
)

type router struct {
	pattern *regexp.Regexp
	handler Handler
}

type ServeMux struct {
	routes []router
}

func NewServeMux() *ServeMux {
	return &ServeMux{}
}

func handleConnection(conn net.Conn, handler Handler) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		req, err := ReadRequest(reader)
		if err != nil {
			fmt.Fprintf(conn, "HTTP/1.1 400 Bad Request\r\n\r\n%s\n", err)
			return
		}

		if req == nil {
			return
		}

		w := &simpleResponseWriter{conn: conn}
		handler.Serve(w, req)
		w.Flush()

		close := strings.ToLower(req.Headers["Connection"]) == "close"
		if close || req.Protocol == "HTTP/1.0" {
			return
		}
	}
}

func ListenAndServe(addr string, hander Handler) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(conn, hander)
	}
}

func (m *ServeMux) Serve(w ResponseWriter, r *Request) {
	for _, rt := range m.routes {
		if rt.pattern.MatchString(r.Path) {
			matches := rt.pattern.FindStringSubmatch(r.Path)
			params := make(map[string]string)
			for i, name := range rt.pattern.SubexpNames() {
				if i > 0 && name != "" {
					params[name] = matches[i]
				}
			}
			r.Params = params
			rt.handler.Serve(w, r)
			return
		}
	}
	fmt.Fprintln(w, "HTTP/1.1 404 Not Found\r\nContent-Type: text/plain\r\n\r\nNot Found")
}

func (m *ServeMux) HandleFunc(path string, handler func(ResponseWriter, *Request)) {
	converted := convertPattern(path)
	re := regexp.MustCompile(converted)
	m.routes = append(m.routes, router{pattern: re, handler: HandlerFunc(handler)})
}

func convertPattern(p string) string {
	// Replace ":param" with regex group
	re := regexp.MustCompile(`:([a-zA-Z0-9_]+)`)
	return "^" + re.ReplaceAllString(p, `(?P<$1>[^/]+)`) + "$"
}
