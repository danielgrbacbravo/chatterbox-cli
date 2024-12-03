package server

import (
	"bufio"
	"bytes"
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

	// Client management logic (unchanged)
	clientsMu.Lock()
	clients = append(clients, conn)
	clientsMu.Unlock()
	log.Info("Client joined:", "address", conn.RemoteAddr())
	defer func() { /* client removal logic */ }()

	reader := bufio.NewReader(conn)
	for {
		// Read the length prefix (assuming it's a uint32 for this example)
		var length uint32
		if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
			log.Error("Error reading message length:", "err", err)
			break
		}

		// Read the actual message based on the length
		rawData := make([]byte, length)
		if _, err := io.ReadFull(reader, rawData); err != nil {
			log.Error("Error reading message:", "err", err)
			break
		}

		// Process the message (same as before)
		deserializedChatEvent, err := serialization.DeserializeChatEvent(rawData)
		if err != nil {
			log.Error("Error deserializing chat event:", "err", err)
			continue
		}

		log.Info("Received chat event:", "chatEvent", deserializedChatEvent.String())
		broadcastChatEvent(conn, deserializedChatEvent)
	}
}

func broadcastChatEvent(conn *tls.Conn, chatEvent *pb.ChatEvent) {
	log.Info("Broadcasting message:", "chatEvent", chatEvent)
	clientsMu.Lock()
	defer clientsMu.Unlock()

	serializedChatEvent, err := serialization.SerializeChatEvent(chatEvent)
	if err != nil {
		log.Error("Error serializing chat event:", "err", err)
		return
	}

	var buf bytes.Buffer
	err = binary.Write(&buf, binary.BigEndian, uint32(len(serializedChatEvent)))
	if err != nil {
		log.Error("Error writing message length:", "err", err)
	}

	_, err = buf.Write(serializedChatEvent)
	if err != nil {
		log.Error("Error writing message:", "err", err)
	}

	for _, client := range clients {
		if client != conn {
			_, err := client.Write(buf.Bytes())
			if err != nil {
				log.Error("Error sending message to client:", "err", err, "client", client.RemoteAddr())
			}
		}
	}
}
