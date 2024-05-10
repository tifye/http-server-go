package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

type Request struct {
	method   string
	path     string
	params   map[string]string
	protocol string
	headers  map[string]string
	body     io.Reader
}

func parseRequest(r io.Reader) (*Request, error) {
	bodyReader := bufio.NewReader(r)

	statusLine, err := bodyReader.ReadBytes('\n')
	if err != nil {
		return &Request{}, fmt.Errorf("Failed to read request data %w", err)
	}
	if bytes.Equal(statusLine, []byte("")) {
		return &Request{}, fmt.Errorf("Empty request")
	}

	statusParts := bytes.Split(statusLine, []byte(" "))
	if len(statusParts) != 3 {
		return &Request{}, fmt.Errorf("Malformed status line '%b'", statusLine)
	}
	method, path, protocol := statusParts[0], statusParts[1], statusParts[2]

	headers := make(map[string]string, 0)
	for {
		headerLine, err := bodyReader.ReadBytes('\n')
		if err != nil {
			return &Request{}, fmt.Errorf("Failed to read request data %w", err)
		}
		headerLine = bytes.Trim(headerLine, "\r\n")

		if len(headerLine) == 0 {
			break
		}

		headerParts := bytes.SplitN(headerLine, []byte{':'}, 2)
		if len(headerParts) != 2 {
			return &Request{}, fmt.Errorf("Malformed header '%s'", string(headerLine))
		}
		key, value := headerParts[0], headerParts[1]
		headers[string(key)] = string(bytes.Trim(value, " "))
	}

	return &Request{
		method:   string(method),
		path:     string(path),
		protocol: string(protocol),
		headers:  headers,
		body:     bodyReader,
	}, nil
}
