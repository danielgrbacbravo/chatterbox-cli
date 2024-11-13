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

	message := &pb.Message{
		User:    user,
		Message: "Hello World",
	}

	chatEvent := &pb.ChatEvent{
		EventID: 1,
		Event: &pb.ChatEvent_UserMessage{
			UserMessage: message,
		},
	}

	fmt.Println("chatEventCreation: ", chatEvent)

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

	// begin go routine for receiving chat events
	chatEventChan := make(chan *pb.ChatEvent)
	userInputChan := make(chan string)
	// begin go routine for receiving chat events
	go receiver.ReceiveChatEvents(conn, chatEventChan)
	// begin go routine for printing chat events
	go sender.SendUserMessages(conn, userInputChan, user)
	// Goroutine to listen for user input
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("Enter message: ")
			userInput, _ := reader.ReadString('\n')
			userInput = strings.TrimSpace(userInput)
			// check if text isnt empty
			if userInput != "" {
				userInputChan <- userInput
			}
		}
	}()

	// Wait for program termination
	select {}
}
