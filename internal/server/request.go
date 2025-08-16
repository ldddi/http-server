package server

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"strings"
)

const maxBodyBytes = 64 * 1024

var (
	ErrBadRequestLine   = errors.New("malformed request line")
	ErrHeaderTooLarge   = errors.New("header section too large")
	ErrBadContentLength = errors.New("invalid Content-Length")
	ErrBodyTooLarge     = errors.New("request body too large")
	ErrUnexpectedEOF    = errors.New("unexpected EOF")
)

type Request struct {
	Headers  map[string]string
	Method   string
	Params   map[string]string
	Query    map[string]string
	Path     string
	Body     string
	Protocol string
}

func ReadRequest(r *bufio.Reader) (*Request, error) {
	line, err := ReadLine(r)
	if err != nil {
		return nil, err
	}
	fields := strings.Fields(line)
	if len(fields) != 3 {
		return nil, ErrBadRequestLine
	}

	method, rawURL, protocol := fields[0], fields[1], fields[2]
	path := rawURL
	queryString := ""
	if idx := strings.Index(path, "?"); idx != -1 {
		path = rawURL[:idx]
		queryString = rawURL[idx+1:]
	}

	// headers
	headers := make(map[string]string)
	for {
		line, err := ReadLine(r)
		if err != nil {
			return nil, err
		}
		if line == "" {
			break
		}

		idx := strings.IndexByte(line, ':')
		if idx == -1 {
			continue
		}
		headers[line[:idx]] = line[idx+1:]
	}

	// body
	var body string
	if bodyLen, ok := headers["Content-Length"]; ok {
		len, err := strconv.Atoi(bodyLen)
		if err != nil {
			return nil, err
		}

		if len > maxBodyBytes {
			return nil, ErrBodyTooLarge
		}

		buf := make([]byte, len)
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}
		body = string(buf)
	}

	return &Request{
		Method:   method,
		Path:     path,
		Protocol: protocol,
		Headers:  headers,
		Body:     body,
		Query:    parseQuery(queryString),
	}, nil
}

func ReadLine(r *bufio.Reader) (string, error) {
	const maxLine = 4096
	s, err := r.ReadString('\n')
	if err != nil {
		if err == io.EOF && len(s) == 0 {
			return "", err
		}
		return "", ErrUnexpectedEOF
	}

	if len(s) > maxLine {
		return "", ErrHeaderTooLarge
	}

	return strings.TrimRight(s, "\r\n"), nil
}

func parseQuery(qs string) map[string]string {
	params := make(map[string]string)
	if qs == "" {
		return params
	}

	pairs := strings.Split(qs, "&")
	for _, pair := range pairs {
		sl := strings.SplitN(pair, "=", 2)
		if len(sl) == 2 {
			params[sl[0]] = sl[1]
		}
	}
	return params
}
