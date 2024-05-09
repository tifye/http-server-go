package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	router := newRouter()
	router.GET("/", func(req *Request, res ResponseWriter) {
		res.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	})
	router.GET("/meep", func(req *Request, res ResponseWriter) {
		res.Write([]byte("HTTP/1.1 201 OK\r\n\r\n"))
	})
	router.GET("/meep/mino", func(req *Request, res ResponseWriter) {
		res.Write([]byte("HTTP/1.1 202 OK\r\n\r\n"))
	})
	router.GET("/:meep/mino/:mino", func(req *Request, res ResponseWriter) {
		res.Write([]byte("HTTP/1.1 202 OK\r\n\r\n"))
	})
	router.GET("/echo/:text", func(req *Request, res ResponseWriter) {
		text, ok := req.params["text"]
		if !ok {
			fmt.Println("Failed to find value for param 'text'")
			res.Write([]byte("HTTP/1.1 500 Server Error\r\n\r\n"))
			return
		}
		buf := bytes.NewBufferString("HTTP/1.1 200 OK\r\n\r\n")
		buf.WriteString("Content-Type: text/plain\r\n")
		buf.WriteString(fmt.Sprintf("Content-Length: %d\r\n", len(text)))
		buf.WriteString("\r\n")
		buf.WriteString(text)
		buf.WriteTo(res)
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
	for key, value := range request.params {
		fmt.Printf("%s: %s\n", key, value)
	}

	handler(&request, connection)
}
