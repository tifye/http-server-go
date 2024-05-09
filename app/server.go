package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	router := newRouter()
	router.GET("/", func(req *Request, res ResponseWriter) {
		res.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	})

	listener, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	connection, err := listener.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	defer func() {
		err = connection.Close()
		if err != nil {
			fmt.Println("Failed to close connection")
			os.Exit(1)
		}
	}()

	request, err := parseRequest(connection)
	if err != nil {
		fmt.Printf("Failed to parse request got %s\n", err)
		os.Exit(1)
	}

	handler := router.findHandler(&request)
	if handler == nil {
		fmt.Printf("No handler for %s %s", request.method, request.path)
		return
	}
	handler(&request, connection)
}

type ResponseWriter io.Writer
type Handler func(*Request, ResponseWriter)

type Router struct {
	handlers        map[string]map[string]Handler
	notFoundHandler Handler
}

type Request struct {
	method   string
	path     string
	protocol string
	headers  map[string]string
	body     io.Reader
}

func newRouter() Router {
	return Router{
		handlers: make(map[string]map[string]Handler),
		notFoundHandler: func(req *Request, resp ResponseWriter) {
			resp.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		},
	}
}

func (router Router) findHandler(request *Request) Handler {
	h, ok := router.handlers[request.method][request.path]
	if !ok {
		return router.notFoundHandler
	}
	return h
}

func (router Router) GET(path string, handler Handler) {
	const GET string = "GET"
	if len(router.handlers[GET]) == 0 {
		router.handlers[GET] = make(map[string]Handler)
	}
	router.handlers[GET][path] = handler
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
