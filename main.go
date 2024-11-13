package main

import (
	"chatterbox-cli/client"
	"chatterbox-cli/server"
	"fmt"
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

	// json, err := os.ReadFile("user.json")
	// if err != nil {
	// 	panic(err)
	// }

	// user, err := parser.ParseUserFromJson(json)
	// if err != nil {
	// 	panic(err)
	// }

	// message := &pb.Message{
	// 	User:    user,
	// 	Message: "Hello World",
	// }

	// chatEvent := &pb.ChatEvent{
	// 	EventID: 1,
	// 	Event: &pb.ChatEvent_UserMessage{
	// 		UserMessage: message,
	// 	},
	// }

	// serializedChatEvent, err := serialization.SerializeChatEvent(chatEvent)
	// if err != nil {
	// 	panic(err)
	// }

	// deserializedChatEvent, err := serialization.DeserializeChatEvent(serializedChatEvent)
	// if err != nil {
	// 	panic(err)
	// }

	// serverUpdate := &pb.ServerUpdate{
	// 	Reason: 1,
	// 	Motd:   "Hello World",
	// }

	// chatEvent2 := &pb.ChatEvent{
	// 	EventID: 2,
	// 	Event: &pb.ChatEvent_ServerUpdate{
	// 		ServerUpdate: serverUpdate,
	// 	},
	// }

	// // serializedChatEvent2, err := serialization.SerializeChatEvent(chatEvent2)
	// if err != nil {
	// 	panic(err)
	// }

	// deserializedChatEvent2, err := serialization.DeserializeChatEvent(serializedChatEvent2)

	// if err != nil {
	// 	panic(err)
	// }

	if serverMode {
		server.Server()
	} else {
		client.Client()
	}

}
