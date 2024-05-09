package main

import (
	"fmt"
	"io"
	"strings"
)

type ResponseWriter io.Writer
type Handler func(*Request, ResponseWriter)

type Router struct {
	methodTrees     map[string]*RouteNode
	notFoundHandler Handler
}

func newRouter() Router {
	return Router{
		methodTrees: make(map[string]*RouteNode, 0),
		notFoundHandler: func(req *Request, resp ResponseWriter) {
			resp.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
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

func extractRouteParams(handlerPath string, requestPath string) map[string]string {
	handlerParts := strings.Split(handlerPath, "/")
	requestParts := strings.Split(requestPath, "/")
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
