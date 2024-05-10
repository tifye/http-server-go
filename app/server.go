package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
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

	handler := router.findHandler(request)
	if handler == nil {
		fmt.Printf("No handler for %s %s", request.method, request.path)
		return
	}

	handler(request, conn)
}
