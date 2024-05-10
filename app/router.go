package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type ResponseWriter struct {
	Headers map[string]string
	writer  io.WriteCloser
	status  int
}
type Handler func(*Request, *ResponseWriter)

func (r *ResponseWriter) Status(status int) {
	if !isValidStatusCode(status) {
		panic(fmt.Sprintf("Invalid http status code: %d", status))
	}
	r.status = status
}

func (r *ResponseWriter) WriteHeader(status int) (n int, err error) {
	if !isValidStatusCode(r.status) {
		r.Status(status)
	}
	return r.Write([]byte{})
}

func (r *ResponseWriter) Write(b []byte) (n int, err error) {
	if r.status < 100 {
		r.status = http.StatusOK
	}
	var buf bytes.Buffer
	_, err = buf.WriteString(fmt.Sprintf("HTTP/1.1 %d %s\r\n", r.status, http.StatusText(r.status)))
	if err != nil {
		return buf.Len(), err
	}

	for key, value := range r.Headers {
		_, err = buf.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
		if err != nil {
			return buf.Len(), err
		}
	}

	if _, err = buf.WriteString("\r\n"); err != nil {
		return buf.Len(), err
	}
	if n, err = buf.Write(b); err != nil {
		return buf.Len(), err
	}

	nb, err := buf.WriteTo(r.writer)
	return int(nb), err
}

func isValidStatusCode(status int) bool {
	return status >= 100 && status <= 599
}

type Router struct {
	methodTrees     map[string]*RouteNode
	notFoundHandler Handler
}

func newRouter() Router {
	return Router{
		methodTrees: make(map[string]*RouteNode, 0),
		notFoundHandler: func(req *Request, resp *ResponseWriter) {
			resp.WriteHeader(http.StatusNotFound)
		},
	}
}

func (router Router) addHandler(method string, path string, handler Handler) {
	if len(path) >= 1 && !strings.HasPrefix(path, "/") {
		panic("path must start with '/'")
	}

	methodTree, ok := router.methodTrees[method]
	if !ok {
		root := newRouteNode("", "", false, nil)
		methodTree = &root
		router.methodTrees[method] = &root
	}

	methodTree.addNode(path, handler)
}

func (router Router) findHandler(request *Request) Handler {
	root, ok := router.methodTrees[request.method]
	if !ok || root == nil {
		return router.notFoundHandler
	}

	if len(request.path) == 0 || !strings.HasPrefix(request.path, "/") {
		return router.notFoundHandler
	}

	node := root.match(request.path)
	if node == nil || node.handler == nil {
		return router.notFoundHandler
	}

	request.params = extractRouteParams(node.path, request.path)
	return node.handler
}

func (router Router) GET(path string, handler Handler) {
	router.addHandler("GET", path, handler)
}

func (router Router) POST(path string, handler Handler) {
	router.addHandler("POST", path, handler)
}

func extractRouteParams(handlerPath string, requestPath string) map[string]string {
	handlerParts := strings.Split(strings.TrimPrefix(handlerPath, "/"), "/")
	requestParts := strings.Split(strings.TrimPrefix(requestPath, "/"), "/")
	if len(handlerParts) != len(requestParts) {
		panic(fmt.Sprintf("Handler/Request mismatch handler='%s' request='%s'", handlerPath, requestPath))
	}

	params := make(map[string]string)
	for i, hPart := range handlerParts {
		if !strings.HasPrefix(hPart, ":") {
			continue
		}

		params[hPart[1:]] = requestParts[i]
	}

	return params
}
