package main

import (
	"chatterbox-cli/client"
	"chatterbox-cli/parser"
	pb "chatterbox-cli/proto"
	"chatterbox-cli/serialization"
	"chatterbox-cli/server"
	"fmt"
	"os"
)

func readServerMode(prompt string) bool {
	fmt.Print(prompt)
	fmt.Println()
	var serverMode bool
	var input string
	fmt.Scanln(&input)
	if input == "y" {
		serverMode = true
	} else {
		serverMode = false
	}
	return serverMode
}

func main() {

	// ask if you want to run in server mode or client mode
	serverMode := readServerMode("Do you want to run in server mode? (y/n): ")

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

	fmt.Println("chatEventCration: ", chatEvent)

	serializedChatEvent, err := serialization.SerializeChatEvent(chatEvent)
	if err != nil {
		panic(err)
	}

	fmt.Print("\n chatEventSerialization: ", serializedChatEvent)

	deserializedChatEvent, err := serialization.DeserializeChatEvent(serializedChatEvent)
	if err != nil {
		panic(err)
	}
	fmt.Println("\n chatEventDeserialization: ", deserializedChatEvent)

	serverUpdate := &pb.ServerUpdate{
		Reason: 1,
		Motd:   "Hello World",
	}

	chatEvent2 := &pb.ChatEvent{
		EventID: 2,
		Event: &pb.ChatEvent_ServerUpdate{
			ServerUpdate: serverUpdate,
		},
	}

	serializedChatEvent2, err := serialization.SerializeChatEvent(chatEvent2)
	if err != nil {
		panic(err)
	}

	fmt.Print("\n chatEvent2Serialization: ", serializedChatEvent2)

	deserializedChatEvent2, err := serialization.DeserializeChatEvent(serializedChatEvent2)

	if err != nil {
		panic(err)
	}

	fmt.Println("\n chatEvent2Deserialization: ", deserializedChatEvent2)

	if serverMode {
		server.Server()
	} else {
		client.Client()
	}

}
