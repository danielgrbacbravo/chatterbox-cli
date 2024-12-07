package client

import (
	"bufio"
	"chatterbox-cli/parser"
	pb "chatterbox-cli/proto"
	"chatterbox-cli/receiver"
	"chatterbox-cli/sender"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"strings"
)

func Client() {

	json, err := os.ReadFile("user.json")
	if err != nil {
		panic(err)
	}

	user, err := parser.ParseUserFromJson(json)
	if err != nil {
		panic(err)
	}

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

	sendChatEventChan := make(chan *pb.ChatEvent)
	receiveChatEventChan := make(chan *pb.ChatEvent)

	go receiver.ReceiveChatEvents(conn, receiveChatEventChan)
	go sender.SendChatEvent(conn, sendChatEventChan)

	// Start message receiver goroutine
	go func() {
		fmt.Println("Message receiver started...")
		for chatEvent := range receiveChatEventChan {
			switch event := chatEvent.Event.(type) {
			case *pb.ChatEvent_UserMessage:
				// Clear current line
				fmt.Printf("\r\033[K") // Clear the current line
				// Print received message
				fmt.Printf("\n%s: %s\n", event.UserMessage.User.DisplayName, event.UserMessage.Message)
				// Reprint prompt
				fmt.Print("Enter message: ")
			}
		}
		fmt.Println("Message receiver stopped")
	}()

	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("Enter message: ")

			userInput, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading input:", "err", err)
				continue
			}

			userInput = strings.TrimSpace(userInput)
			if userInput != "" {
				message := &pb.Message{
					User:    user,
					Message: userInput,
				}

				chatEvent := &pb.ChatEvent{
					EventID: 1,
					Event: &pb.ChatEvent_UserMessage{
						UserMessage: message,
					},
				}
				sendChatEventChan <- chatEvent
			}
		}
	}()

	// Wait for program termination
	select {}
}
