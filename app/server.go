package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

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

	resp := []byte("HTTP/1.1 200 OK\r\n\r\n")
	_, err = connection.Write(resp)
	if err != nil {
		fmt.Println("Failed to write to connection")
		os.Exit(1)
	}
	// err = connection.Close()
	// if err != nil {
	// 	fmt.Println("Failed to close connection")
	// 	os.Exit(1)
	// }
}
