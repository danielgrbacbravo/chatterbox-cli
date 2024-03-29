package server

import (
	"crypto/ecdsa"
	"math/big"
	"net"

	"go-chat-cli/crypto"
	"go-chat-cli/message"

	log "github.com/charmbracelet/log"
)

var clients = make([]*net.Conn, 0)
var serverPrivateKey *ecdsa.PrivateKey
var serverPublicKey ecdsa.PublicKey

// map of client public keys to their connections
var sharedKeys = make(map[net.Conn]*big.Int)

func Server() {
	var ip = getOutboundIP()
	log.Info("Server starting...")
	log.Info("Server address:", "addr", ip.String())

	// construct server public key
	log.Info("generating server public key...")
	serverPrivateKey = crypto.GeneratePrivateKey()
	serverPublicKey = crypto.GeneratePublicKey(serverPrivateKey)
	log.Info("server public key generated")

	// Listen on TCP port 8080 on all available unicast and
	// anycast IP addresses of the local system.
	listener, err := net.Listen("tcp", ":5051")
	if err != nil {
		log.Error("Error listening:", "err", err)
		return
	}
	log.Debug("listener created successfully", "addr", listener.Addr().String())
	defer listener.Close()

	for {
		// Wait for a connection.
		conn, err := listener.Accept()
		if err != nil {
			log.Error("Error accepting connection:", "err", err)
			return
		}
		log.Info("creating new thread for connection", "addr", conn.RemoteAddr().String())
		go handleConnection(conn) // Handle the connection in a new goroutine.
	}
}

func handleConnection(conn net.Conn) {
	var username string
	var clientPublicKey ecdsa.PublicKey
	var sharedKey *big.Int

	// establish connection with client
	crypto.SendPublicKey(conn, serverPublicKey)
	log.Info("server public key sent to client")
	defer conn.Close()
	clients = append(clients, &conn)

	// construct the join message

	// defer the leave message
	defer func() {
		msg := constructLeaveMessage(username)
		crypto.BroadcastMessage(msg, clients, sharedKeys)
		// remove the connection from the clients slice
		log.Warn("closing connection ", "addr", conn.RemoteAddr().String())
		for i, client := range clients {
			if client == &conn {
				clients = append(clients[:i], clients[i+1:]...)
				break
			}
		}
	}()
	// read the client public key
	for {
		msg, err := message.ReadMessage(conn)
		if err != nil {
			return
		}
		if msg.MessageType == "public_key" {
			clientPublicKey = crypto.ConvertToPublicKey(msg)
			log.Info("client public key received")
			sharedKey = crypto.GenerateSharedSecret(serverPrivateKey, clientPublicKey)
			log.Info("Diffie Hellman key exchange complete")
			sharedKeys[conn] = sharedKey
			break
		}
	}

	// read messages from the connection
	for {
		msg, err := message.ReadMessage(conn)
		if err != nil {
			return
		}
		// decrypt message
		log.Debug("Message received", "msg", msg.Message, "username", msg.Username, "type", msg.MessageType)
		log.Debug("decrypting message ...")
		decryptedMessage := crypto.DecryptMessage(msg, sharedKey)
		log.Debug("decrypted message", "msg", decryptedMessage.Message, "username", decryptedMessage.Username, "type", decryptedMessage.MessageType)

		if username == "" {
			// decrypt message
			// decrypt message using shared key
			username = decryptedMessage.Username
			log.Info("username registered to connection handler", "username", username)
		}

		// re-encrypt message using shared key and broadcast
		crypto.BroadcastMessage(decryptedMessage, clients, sharedKeys)
		log.Debug("Message broadcasted", "msg", msg.Message, "username", msg.Username, "type", msg.MessageType)
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

func constructJoinMessage(username string) message.Message {
	return message.Message{
		Username:    username,
		Message:     "",
		MessageType: "join",
	}
}
