package server

import (
	"crypto/ecdsa"
	"math/big"
	"net"

	"chatterbox-cli/crypto"
	"chatterbox-cli/message"

	"github.com/charmbracelet/lipgloss"
	log "github.com/charmbracelet/log"
)

const SuccessLevel = log.InfoLevel + 2
const MessageLevel = log.InfoLevel + 1
const DisconnectLevel = log.InfoLevel + 3

const decryptLevel = log.InfoLevel + 4
const encryptLevel = log.InfoLevel + 5

var clients = make([]*net.Conn, 0)
var serverPrivateKey *ecdsa.PrivateKey
var serverPublicKey ecdsa.PublicKey

// map of client public keys to their connections
var sharedKeys = make(map[net.Conn]*big.Int)

func Server(serverName string) {
	//success message
	styles := log.DefaultStyles()
	styles.Levels[SuccessLevel] = lipgloss.NewStyle().
		SetString("SUCCESS").
		Bold(true).
		Foreground(lipgloss.Color("42"))

	styles.Levels[MessageLevel] = lipgloss.NewStyle().
		SetString("MESSAGE").
		Bold(true).
		Foreground(lipgloss.Color("39"))

	styles.Levels[DisconnectLevel] = lipgloss.NewStyle().
		SetString("DISCONNECT").
		Bold(true).
		Foreground(lipgloss.Color("208"))

	styles.Levels[decryptLevel] = lipgloss.NewStyle().
		SetString("DECRYPT").
		Bold(true).
		Foreground(lipgloss.Color("199"))

	styles.Levels[encryptLevel] = lipgloss.NewStyle().
		SetString("ENCRYPT").
		Bold(true).
		Foreground(lipgloss.Color("199"))

	log.SetStyles(styles)

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
	Success("Server started successfully", "addr", listener.Addr().String())
	defer listener.Close()

	for {
		// Wait for a connection.
		conn, err := listener.Accept()
		if err != nil {
			log.Error("Error accepting connection:", "err", err)
			return
		}
		Success("creating new thread for connection", "addr", conn.RemoteAddr().String())
		go handleConnection(serverName, conn) // Handle the connection in a new goroutine.
	}
}

func handleConnection(servername string, conn net.Conn) {
	var username string
	var clientPublicKey ecdsa.PublicKey
	var sharedKey *big.Int

	// establish connection with client
	crypto.SendPublicKey(servername, conn, serverPublicKey)
	log.Debug("server public key sent to client")
	defer conn.Close()
	clients = append(clients, &conn)

	// construct the join message

	// defer the leave message
	defer func() {
		msg := constructLeaveMessage(username)
		crypto.BroadcastMessage(msg, clients, sharedKeys)
		// remove the connection from the clients slice
		Disconnect("closing connection", "addr", conn.RemoteAddr().String())
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
			Success("Diffie Hellman key exchange complete", "sharedKey", sharedKey)
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
		Message("Message received", "msg", msg.Message, "username", msg.Username, "type", msg.MessageType)
		Decrypt("decrypting message ...")
		decryptedMessage := crypto.DecryptMessage(msg, sharedKey)
		log.Debug("decrypted message", "msg", decryptedMessage.Message, "username", decryptedMessage.Username, "type", decryptedMessage.MessageType)

		if username == "" {
			// decrypt message
			// decrypt message using shared key
			username = decryptedMessage.Username
			log.Debug("username registered to connection handler", "username", username)
		}
		// re-encrypt message using shared key and broadcast
		Encrypt("encrypting message ...")
		crypto.BroadcastMessage(decryptedMessage, clients, sharedKeys)
		Message("Message broadcasted", "msg", decryptedMessage.Message, "username", decryptedMessage.Username, "type", decryptedMessage.MessageType)
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

func Success(msg string, args ...any) {
	log.Default().Log(SuccessLevel, msg, args...)
}

func Message(msg string, args ...any) {
	log.Default().Log(MessageLevel, msg, args...)
}

func Disconnect(msg string, args ...any) {
	log.Default().Log(DisconnectLevel, msg, args...)
}

func Encrypt(msg string, args ...any) {
	log.Default().Log(encryptLevel, msg, args...)
}

func Decrypt(msg string, args ...any) {
	log.Default().Log(decryptLevel, msg, args...)
}
