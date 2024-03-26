package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"

	log "github.com/charmbracelet/log"
)

func main() {
	var dialAddress string
	flag.StringVar(&dialAddress, "address", "", "Address to dial")
	flag.Parse()
	if dialAddress != "" {
		log.Info("Dialing to address: ", "addr", dialAddress)
	}
	if dialAddress == "" {
		Server()
	} else {
		Client(dialAddress)
	}
}

func Server() {
	log.Info("Starting server")
	// Listen on TCP port 8080 on all available unicast and
	// anycast IP addresses of the local system.
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer listener.Close()

	for {
		// Wait for a connection.
		conn, err := listener.Accept()
		if err != nil {
			log.Error("Error accepting connection:", "err", err)
			return
		}
		go handleConnection(conn) // Handle the connection in a new goroutine.
	}
}

func Client(dialAddress string) {
	log.Info("Starting client")

	// Connect to the server
	// Dial the server at the address provided
	conn, err := net.Dial("tcp", dialAddress)
	if err != nil {
		log.Error("Error dialing server:", "err", err)
		return
	}
	defer conn.Close()

	// read stdin
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter Message: ")
	text, _ := reader.ReadString('\n')
	log.Info("Message sent:", "message", text)

	// Create a new writer for the connection
	// Write a message to the server
	writer := bufio.NewWriter(conn)
	_, err = writer.WriteString(text)
	if err != nil {
		log.Error("Error writing message:", "err", err)
		return
	}
	writer.Flush()
}

func handleConnection(conn net.Conn) {
	defer conn.Close() // Close the connection when the function returns

	// Create a new reader for the connection
	reader := bufio.NewReader(conn)

	for {
		// Read a message from the client
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Error("Error reading message:", "err", err)
			return
		}
		log.Info("Message received:", "message", message)
	}
}
