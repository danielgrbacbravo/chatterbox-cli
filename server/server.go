package server

import (
	"chatterbox-cli/serialization"
	"crypto/tls"
	"encoding/binary"
	"io"
	"sync"

	pb "chatterbox-cli/proto"

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

	// Client management logic
	clientsMu.Lock()
	clients = append(clients, conn)
	clientIndex := len(clients) - 1
	clientsMu.Unlock()

	log.Info("Client joined:", "address", conn.RemoteAddr())

	// Proper client cleanup on disconnect
	defer func() {
		clientsMu.Lock()
		if clientIndex < len(clients) {
			clients = append(clients[:clientIndex], clients[clientIndex+1:]...)
		}
		clientsMu.Unlock()
		log.Info("Client left:", "address", conn.RemoteAddr())
	}()

	lengthBuf := make([]byte, 4)
	for {
		// Read exactly 4 bytes for the length prefix
		_, err := io.ReadFull(conn, lengthBuf)
		if err != nil {
			if err == io.EOF {
				log.Info("Connection closed")
				return
			}
			log.Error("Error reading message length:", "err", err)
			return
		}

		length := binary.BigEndian.Uint32(lengthBuf)

		// Read the message payload
		rawData := make([]byte, length)
		_, err = io.ReadFull(conn, rawData)
		if err != nil {
			if err == io.EOF {
				log.Info("Connection closed during message read")
				return
			}
			log.Error("Error reading message:", "err", err)
			return
		}

		// Process the message
		deserializedChatEvent, err := serialization.DeserializeChatEvent(rawData)
		if err != nil {
			log.Error("Error deserializing chat event:", "err", err)
			continue
		}

		log.Info("Received chat event:", "chatEvent", deserializedChatEvent.String())
		broadcastChatEvent(conn, deserializedChatEvent)
	}
}

func broadcastChatEvent(sender *tls.Conn, chatEvent *pb.ChatEvent) {
	log.Info("Starting broadcast to all clients")
	clientsMu.Lock()
	defer clientsMu.Unlock()

	log.Info("Number of connected clients:", "count", len(clients))

	serializedChatEvent, err := serialization.SerializeChatEvent(chatEvent)
	if err != nil {
		log.Error("Error serializing chat event:", "err", err)
		return
	}

	message := make([]byte, 4+len(serializedChatEvent))
	binary.BigEndian.PutUint32(message[:4], uint32(len(serializedChatEvent)))
	copy(message[4:], serializedChatEvent)

	log.Info("Prepared message for broadcast:", "messageLength", len(message))

	var failedClients []int
	for i, client := range clients {
		if client != sender { // Skip the sender
			log.Info("Attempting to send to client:", "address", client.RemoteAddr())
			n, err := client.Write(message)
			if err != nil {
				log.Error("Error sending message to client:", "err", err, "client", client.RemoteAddr())
				failedClients = append(failedClients, i)
			} else {
				log.Info("Successfully sent message to client:", "address", client.RemoteAddr(), "bytes", n)
			}
		} else {
			log.Info("Skipping sender:", "address", client.RemoteAddr())
		}
	}

	// Remove failed clients in reverse order
	for i := len(failedClients) - 1; i >= 0; i-- {
		failedIndex := failedClients[i]
		clients = append(clients[:failedIndex], clients[failedIndex+1:]...)
		log.Info("Removed failed client at index:", "index", failedIndex)
	}
}
