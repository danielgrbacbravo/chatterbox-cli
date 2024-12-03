package sender

import (
	"bytes"
	pb "chatterbox-cli/proto"
	"chatterbox-cli/serialization"
	"crypto/tls"
	"encoding/binary"

	"log"
)

func SendChatEvent(conn *tls.Conn, chatEvent chan *pb.ChatEvent) {
    for {
        // Wait for new messages from the channel
        event := <-chatEvent

        // Serialize the chatEvent
        serializedChatEvent, err := serialization.SerializeChatEvent(event)
        if err != nil {
            log.Printf("Error serializing message: %s", err)
            continue
        }

        // Create a buffer to hold length + serialized data
        var buf bytes.Buffer

        // Write the length of the serialized message as a uint32
        err = binary.Write(&buf, binary.BigEndian, uint32(len(serializedChatEvent)))
        if err != nil {
            log.Printf("Error writing message length: %s", err)
            continue
        }

        // Write the actual serialized data to the buffer
        _, err = buf.Write(serializedChatEvent)
        if err != nil {
            log.Printf("Error writing serialized message: %s", err)
            continue
        }

        // Send the complete buffer to the server
        _, err = conn.Write(buf.Bytes())
        if err != nil {
            log.Printf("Error sending message to server: %s", err)
            continue
        }
    }
}
