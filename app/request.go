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

func parseRequest(r io.Reader) (Request, error) {
	scanner := bufio.NewScanner(r)

	if didScanStatus := scanner.Scan(); !didScanStatus {
		if err := scanner.Err(); err != nil {
			return Request{}, fmt.Errorf("Failed to read request data %w", err)
		}
		return Request{}, fmt.Errorf("Empty request")
	}
	statusLine := scanner.Bytes()
	statusParts := bytes.Split(statusLine, []byte(" "))
	if len(statusParts) != 3 {
		return Request{}, fmt.Errorf("Malformed status line '%b'", statusLine)
	}
	method, path, protocol := statusParts[0], statusParts[1], statusParts[2]

	headers := make(map[string]string, 0)
	for scanner.Scan() {
		headerLine := scanner.Bytes()
		if len(headerLine) == 0 || bytes.Equal(headerLine, []byte{'\r', '\n'}) {
			break
		}

		headerParts := bytes.SplitN(headerLine, []byte{':'}, 2)
		if len(headerParts) != 2 {
			return Request{}, fmt.Errorf("Malformed header '%b'", headerLine)
		}
		key, value := headerParts[0], headerParts[1]
		headers[string(key)] = string(value)
	}

	return Request{
		method:   string(method),
		path:     string(path),
		protocol: string(protocol),
		headers:  headers,
		body:     r,
	}, nil
}
