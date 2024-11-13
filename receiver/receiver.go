package receiver

import (
	"bufio"
	"chatterbox-cli/serialization"
	"crypto/tls"
	"encoding/binary"
	"io"

	pb "chatterbox-cli/proto"

	log "github.com/charmbracelet/log"
)

func ReceiveChatEvents(conn *tls.Conn, chatEvents chan *pb.ChatEvent) {
	defer close(chatEvents) // Close the channel when we're done

	reader := bufio.NewReader(conn)
	for {
		// Read the length prefix (assuming it's a uint32 for this example)
		var length uint32
		log.Info("Attempting to read message length")
		err := binary.Read(reader, binary.BigEndian, &length)
		if err != nil {
			log.Error("Error reading message length:", "err", err)
			// Possibly log the reader buffer here to see if data is being received
			break
		}
		log.Info("Message length received:", "length", length)

		// Read the actual message based on the length
		rawData := make([]byte, length)
		_, err = io.ReadFull(reader, rawData)
		if err != nil {
			// Handle the case where the connection was closed unexpectedly
			if err == io.EOF {
				log.Info("Client disconnected")
				break
			}
			log.Error("Error reading message:", "err", err)
			break
		}

		// Process the message (deserialize it)
		log.Info("Received raw data:", "rawData", rawData)
		deserializedChatEvent, err := serialization.DeserializeChatEvent(rawData)
		if err != nil {
			log.Error("Error deserializing chat event:", "err", err)
			continue
		}

		// Send the deserialized chat event to the channel
		log.Info("Received chat event:", "chatEvent", deserializedChatEvent.String())
		chatEvents <- deserializedChatEvent
	}
}
