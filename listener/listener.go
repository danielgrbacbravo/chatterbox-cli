package listener

import (
	pb "chatterbox-cli/proto"
	"fmt"
)

// print out the chat event
// Change the channel to be bi-directional so you can both send and receive data
func PrintChatEvents(chatEvents chan *pb.ChatEvent) {
	for chatEvent := range chatEvents {
		// Print the chat event
		fmt.Println("Received chat event:", chatEvent.GetUserMessage().String())
	}
}
