package server

import (
	"net"

	"go-chat-cli/message"

	log "github.com/charmbracelet/log"
)

type Messsage struct {
	Username string
	Message  string
}

var clients = make([]*net.Conn, 0)

func Server() {
	log.Debug("Starting server")
	// print the server address
	var ip = getOutboundIP()
	log.Debug("Server address:", "addr", ip.String())

	// construct chat room info message

	// Listen on TCP port 8080 on all available unicast and
	// anycast IP addresses of the local system.
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Error("Error listening:", "err", err)
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

func handleConnection(conn net.Conn) {
	var username string
	defer conn.Close()
	clients = append(clients, &conn)
	defer func() {
		// send a leave message to all clients
		msg := constructLeaveMessage(username)
		msg.BroadcastMessage(clients)
		// remove the connection from the clients slice

		log.Warn("closing connection ", "addr", conn.RemoteAddr().String())
		for i, client := range clients {
			if client == &conn {
				clients = append(clients[:i], clients[i+1:]...)
				break
			}
		}
	}()
	for {
		msg, err := message.ReadMessage(conn)
		if err != nil {
			return
		}
		if username == "" {
			username = msg.Username
			log.Info("username registered to connection handler", "username", username)
		}

		err = msg.BroadcastMessage(clients)
		if err != nil {
			log.Error("Error broadcasting message:", "err", err)
			return
		}
	}
}

// Get preferred outbound ip of this machine
func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

func listClients() {
	for _, client := range clients {
		log.Info("Client connected", "addr", (*client).RemoteAddr().String())
	}
}

func constructLeaveMessage(username string) message.Message {
	return message.Message{
		Username:    username,
		Message:     "",
		MessageType: "leave",
	}
}
