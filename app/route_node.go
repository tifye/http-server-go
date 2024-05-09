package main

import (
	"strings"
)

type RouteNode struct {
	handler    Handler
	path       string
	segment    string
	isWildcard bool
	children   []*RouteNode
}

func newRouteNode(path string, segment string, isWildcard bool, handler Handler) RouteNode {
	return RouteNode{
		path:       path,
		segment:    segment,
		isWildcard: isWildcard,
		handler:    handler,
	}
}

func (node *RouteNode) match(path string) *RouteNode {
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		if node.handler != nil {
			return node
		}
		return nil
	}

	before, after, _ := strings.Cut(path, "/")
	for _, child := range node.children {
		if !child.isWildcard && child.segment != before {
			continue
		}

		if match := child.match(after); match != nil {
			return match
		}
	}
	return nil
}

// tail = self + rest
func (node *RouteNode) addNode(tailPath string, handler Handler) {
	tailPath = strings.Trim(tailPath, "/")
	if tailPath == "" {
		if node.handler != nil {
			panic("handler already exists")
		}
		node.handler = handler
		return
	}

	before, after, found := strings.Cut(tailPath, "/")
	for _, child := range node.children {
		if child.segment == before {
			child.addNode(after, handler)
			return
		}
	}

	child := newRouteNode(strings.Join([]string{node.path, before}, "/"), before, strings.HasPrefix(before, ":"), nil)
	node.children = append(node.children, &child)
	if found {
		child.addNode(after, handler)
		return
	}
	child.handler = handler
}
