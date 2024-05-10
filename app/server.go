package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
)

type Config struct {
	publicDir string
}

func main() {
	fmt.Println("Logs from your program will appear here!")

	flags := flag.NewFlagSet("", 0)
	publicDir := flags.String("directory", "", "")
	flags.Parse(os.Args[1:])

	config := Config{
		publicDir: *publicDir,
	}
	router := setupRouter(config)

	listener, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	connChan := make(chan net.Conn)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("Error accepting connection: ", err.Error())
				cancel()
				return
			}
			connChan <- conn
		}
	}()

	for {
		select {
		case conn := <-connChan:
			go serve(ctx, router, conn)
		case <-ctx.Done():
			fmt.Println("Exiting listener goroutine")
			return
		}
	}
}

func serve(ctx context.Context, router *Router, conn net.Conn) {
	defer conn.Close()

	request, err := parseRequest(conn)
	if err != nil {
		fmt.Printf("Failed to parse request got %s\n", err)
		os.Exit(1)
	}

	handler := router.findHandler(&request)
	if handler == nil {
		fmt.Printf("No handler for %s %s", request.method, request.path)
		return
	}

	handler(&request, conn)
}

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
		n, err := io.ReadFull(req.body, contents)
		if err != nil {
			fmt.Printf("err reading request body %s\n", err)
			res.Write([]byte("HTTP/1.1 500 Server Error\r\n\r\n"))
			return
		}

		fmt.Printf("content length is %s", sizeStr)
		fmt.Printf("read %d bytes from body", n)

		err = os.WriteFile(fp, contents, 0644)
		if err != nil {
			fmt.Printf("err writing to file: %s", err)
			res.Write([]byte("HTTP/1.1 500 Server Error\r\n\r\n"))
			return
		}

		fmt.Println("mino")
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
