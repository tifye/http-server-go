package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

func setupRouter(config Config) *Router {
	router := newRouter()
	router.GET("/", handleRoot())
	router.GET("/meep", handleGetMeep())
	router.GET("/:meep/mino/:mino", handleDoubleWild())
	router.POST("/files/:filename", handlePostFile(config.publicDir))
	router.GET("/files/:filename", handleGetFile(config.publicDir))
	router.GET("/echo/:text", handleEchoText())
	router.GET("/user-agent", handleEchoUserAgent())
	return &router
}

func handleRoot() Handler {
	return func(req *Request, res ResponseWriter) {
		res.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	}
}

func handleGetMeep() Handler {
	return func(req *Request, res ResponseWriter) {
		res.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	}
}

func handleDoubleWild() Handler {
	return func(req *Request, res ResponseWriter) {
		res.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	}
}

func handlePostFile(publicDir string) Handler {
	return func(req *Request, res ResponseWriter) {
		filename, _ := req.params["filename"]
		fp := filepath.Join(publicDir, filename)

		sizeStr, _ := req.headers["Content-Length"]
		size, _ := strconv.Atoi(sizeStr)
		contents := make([]byte, size)
		_, err := io.ReadFull(req.body, contents)
		if err != nil {
			fmt.Printf("err reading request body %s\n", err)
			res.Write([]byte("HTTP/1.1 500 Server Error\r\n\r\n"))
			return
		}

		err = os.WriteFile(fp, contents, 0644)
		if err != nil {
			fmt.Printf("err writing to file: %s", err)
			res.Write([]byte("HTTP/1.1 500 Server Error\r\n\r\n"))
			return
		}

		res.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
	}
}

func handleGetFile(publicDir string) Handler {
	return func(req *Request, res ResponseWriter) {
		filename, _ := req.params["filename"]

		fp := filepath.Join(publicDir, filename)
		_, err := os.Stat(fp)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				res.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
				return
			}
			res.Write([]byte("HTTP/1.1 500 Server Error\r\n\r\n"))
		}

		data, err := os.ReadFile(fp)
		if err != nil {
			res.Write([]byte("HTTP/1.1 500 Server Error\r\n\r\n"))
			return
		}

		buf := bytes.NewBufferString("HTTP/1.1 200 OK\r\n")
		buf.WriteString("Content-Type: application/octet-stream\r\n")
		buf.WriteString(fmt.Sprintf("Content-Length: %d\r\n", len(data)))
		buf.WriteString("\r\n")
		buf.Write(data)
		buf.WriteTo(res)
	}
}

func handleEchoText() Handler {
	return func(req *Request, res ResponseWriter) {
		text, _ := req.params["text"]

		buf := bytes.NewBufferString("HTTP/1.1 200 OK\r\n")
		buf.WriteString("Content-Type: text/plain\r\n")
		buf.WriteString(fmt.Sprintf("Content-Length: %d\r\n", len(text)))
		buf.WriteString("\r\n")
		buf.WriteString(text)
		buf.WriteTo(res)
	}
}

func handleEchoUserAgent() Handler {
	return func(req *Request, res ResponseWriter) {
		userAgent, ok := req.headers["User-Agent"]
		if !ok {
			fmt.Println("Failed to find value for param 'text'")
			res.Write([]byte("HTTP/1.1 500 Server Error\r\n\r\n"))
			return
		}
		buf := bytes.NewBufferString("HTTP/1.1 200 OK\r\n")
		buf.WriteString("Content-Type: text/plain\r\n")
		buf.WriteString(fmt.Sprintf("Content-Length: %d\r\n", len(userAgent)))
		buf.WriteString("\r\n")
		buf.WriteString(userAgent)
		buf.WriteTo(res)
	}
}
