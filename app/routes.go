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
	router.GET("/", func(req *Request, res ResponseWriter) {
		res.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	})
	router.GET("/meep", func(req *Request, res ResponseWriter) {
		res.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	})
	router.GET("/meep/mino", func(req *Request, res ResponseWriter) {
		res.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	})
	router.GET("/:meep/mino/:mino", func(req *Request, res ResponseWriter) {
		res.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	})
	router.POST("/files/:filename", func(req *Request, res ResponseWriter) {
		fmt.Println("meep")
		filename, _ := req.params["filename"]
		fp := filepath.Join(config.publicDir, filename)

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
	})
	router.GET("/files/:filename", func(req *Request, res ResponseWriter) {
		filename, _ := req.params["filename"]

		fp := filepath.Join(config.publicDir, filename)
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
	})
	router.GET("/echo/:text", func(req *Request, res ResponseWriter) {
		text, _ := req.params["text"]

		buf := bytes.NewBufferString("HTTP/1.1 200 OK\r\n")
		buf.WriteString("Content-Type: text/plain\r\n")
		buf.WriteString(fmt.Sprintf("Content-Length: %d\r\n", len(text)))
		buf.WriteString("\r\n")
		buf.WriteString(text)
		buf.WriteTo(res)
	})
	router.GET("/user-agent", func(req *Request, res ResponseWriter) {
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
	})
	return &router
}
