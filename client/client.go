package client

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
)

func Client() {
	// Configure the TLS connection
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // Disable verification for self-signed certs; use `false` for production
	}

	// Connect to the TLS server
	conn, err := tls.Dial("tcp", "localhost:5051", tlsConfig)
	if err != nil {
		log.Fatalf("Error connecting to server: %s", err)
	}
	defer conn.Close()

	// Send a message to the server
	fmt.Fprintf(conn, "Hello from TLS client!\n")

	// Read and print server responses
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		fmt.Println("Server:", scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading from server: %s", err)
	}
}
