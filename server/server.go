package server

import (
	"bufio"
	"crypto/tls"
	"sync"

	log "github.com/charmbracelet/log"
)

var (
	clients   = make([]*tls.Conn, 0) // Slice to hold all client connections
	clientsMu sync.Mutex             // Mutex for synchronizing access to clients slice
)

// Server starts the TLS server
func Server() {
	// Load the TLS certificate and private key
	cert, err := tls.LoadX509KeyPair("certs/server.crt", "certs/server.key")
	if err != nil {
		log.Error("Error loading certificate:", "err", err)
		return
	}

	// Configure TLS settings
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}

	// Start listening on a TLS-enabled listener
	listener, err := tls.Listen("tcp", ":5051", tlsConfig)
	if err != nil {
		log.Error("Error listening:", "err", err)
		return
	}
	defer listener.Close()
	log.Info("TLS server started on :5051")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error("Error accepting connection:", "err", err)
			continue
		}

		log.Info("Accepted TLS connection from:", "address", conn.RemoteAddr())
		go handleConnection(conn.(*tls.Conn)) // Handle the connection in a new goroutine
	}
}

func handleConnection(conn *tls.Conn) {
	defer conn.Close()

	// Add the client connection to the list
	clientsMu.Lock()
	clients = append(clients, conn)
	clientsMu.Unlock()
	log.Info("Client joined:", "address", conn.RemoteAddr())

	// Remove the client connection on function exit
	defer func() {
		clientsMu.Lock()
		defer clientsMu.Unlock()
		for i, client := range clients {
			if client == conn {
				clients = append(clients[:i], clients[i+1:]...)
				break
			}
		}
		log.Info("Client left:", "address", conn.RemoteAddr())
	}()

	// Read messages from the client
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		msg := scanner.Text()
		log.Info("Received message:", "message", msg, "from", conn.RemoteAddr())
		broadcastMessage(conn, msg)
	}

	// Check for any scanning error (like a client disconnection)
	if err := scanner.Err(); err != nil {
		log.Error("Error reading from connection:", "err", err, "address", conn.RemoteAddr())
	}
}

// broadcastMessage sends a message to all clients except the sender
func broadcastMessage(sender *tls.Conn, message string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for _, client := range clients {
		if client != sender {
			_, err := client.Write([]byte(message + "\n"))
			if err != nil {
				log.Error("Error sending message to client:", "err", err, "client", client.RemoteAddr())
			}
		}
	}
}
