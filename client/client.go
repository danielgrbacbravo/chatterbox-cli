package client

import (
	"bufio"
	"go-chat-cli/message"
	"net"
	"os"

	log "github.com/charmbracelet/log"
)

func Client(username, dialAddress string) {
	conn, err := net.Dial("tcp", dialAddress)
	if err != nil {
		log.Error("Error dialing:", "err", err)
		return
	}
	defer conn.Close()
	// needs to be able to send and receive messages
	go listenForMessages(conn)
	for {
		// read message from stdin
		reader := bufio.NewReader(os.Stdin)
		log.Info("Enter message: ")
		text, _ := reader.ReadString('\n')
		msg := message.Message{Username: username, Message: text}
		err := msg.SendMessage(conn)
		if err != nil {
			log.Error("Error sending message:", "err", err)
			return
		}
		log.Info("Message sent")
	}
}

func listenForMessages(conn net.Conn) {
	for {
		msg, err := message.ReadMessage(conn)
		if err != nil {
			log.Error("Error reading message:", "err", err)
			return
		}
		log.Info("Message received:", "message", msg.Message)
	}
}
