package sender

import (
	"bytes"
	pb "chatterbox-cli/proto"
	"chatterbox-cli/serialization"
	"crypto/tls"
	"encoding/binary"
	"log"
)

func SendUserMessages(conn *tls.Conn, userInputChan chan string, user *pb.User) {
	for {
		userInput := <-userInputChan

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

		// Serialize the chatEvent
		serializedChatEvent, err := serialization.SerializeChatEvent(chatEvent)
		// Send the chat message to the server

		// Create a buffer to hold length + serialized data
		var buf bytes.Buffer

		// Write the length of the serialized message as a uint32
		err = binary.Write(&buf, binary.BigEndian, uint32(len(serializedChatEvent)))
		if err != nil {
			log.Fatalf("Error writing message length: %s", err)
		}

		// Write the actual serialized data to the buffer
		_, err = buf.Write(serializedChatEvent)
		if err != nil {
			log.Fatalf("Error writing serialized message: %s", err)
		}

		// Send the complete buffer to the server
		_, err = conn.Write(buf.Bytes())
		if err != nil {
			log.Fatalf("Error sending message to server: %s", err)
		}

	}

}
