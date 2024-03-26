package main

import (
	"bufio"
	"encoding/gob"
	"flag"
	"net"
	"os"

	log "github.com/charmbracelet/log"
)

type Messsage struct {
	Username string
	Message  string
}

func main() {
	gob.Register(Messsage{})
	log.SetLevel(log.DebugLevel)
	var username string
	var dialAddress string
	flag.StringVar(&dialAddress, "address", "", "Address to dial")
	flag.StringVar(&username, "username", "", "Username to use")
	flag.Parse()
	if dialAddress != "" {
		log.Debug("Dialing to address: ", "addr", dialAddress)
	}
	if dialAddress == "" {
		Server()
	} else {
		Client(username, dialAddress)
	}
}

func Server() {
	log.Debug("Starting server")
	// print the server address
	var ip = getOutboundIP()
	log.Debug("Server address:", "addr", ip.String())
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

func Client(username string, dialAddress string) {
	log.Debug("Starting client")
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

	for {
		log.Info("Enter message to send:")
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Error("Error reading input:", "err", err)
			break
		}
		msg := Messsage{Username: username, Message: text}
		sendMessage(conn, msg)
		log.Info("Message sent:", "message", text)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close() // Close the connection when the function returns
	for {
		msg, err := readMessage(conn)
		if err != nil {
			break
		}

		log.Info("Message received:", "username", msg.Username, "message", msg.Message)
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

func sendMessage(conn net.Conn, msg Messsage) {
	enc := gob.NewEncoder(conn)
	err := enc.Encode(msg)
	if err != nil {
		log.Error("Error sending message to server:", "err", err)
	}
}

func readMessage(conn net.Conn) (Messsage, error) {
	dec := gob.NewDecoder(conn)
	var msg Messsage
	err := dec.Decode(&msg)
	if err != nil {
		log.Error("Error reading message from server:", "err", err)
		return msg, err
	}
	return msg, nil
}
