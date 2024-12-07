package receiver

import (
	"chatterbox-cli/serialization"
	"crypto/tls"
	"encoding/binary"
	"io"

	pb "chatterbox-cli/proto"

	log "github.com/charmbracelet/log"
)

func ReceiveChatEvents(conn *tls.Conn, chatEvents chan *pb.ChatEvent) {
	defer close(chatEvents)
	log.Info("Starting chat event receiver")

	lengthBuf := make([]byte, 4)
	for {
		// Read exactly 4 bytes for the length prefix
		_, err := io.ReadFull(conn, lengthBuf)
		if err != nil {
			if err == io.EOF {
				log.Info("Connection closed")
				return
			}
			log.Error("Error reading message length:", "err", err)
			return
		}

		length := binary.BigEndian.Uint32(lengthBuf)

		// Read the message payload
		rawData := make([]byte, length)
		_, err = io.ReadFull(conn, rawData)
		if err != nil {
			if err == io.EOF {
				log.Info("Connection closed during message read")
				return
			}
			log.Error("Error reading message payload:", "err", err)
			return
		}

		deserializedChatEvent, err := serialization.DeserializeChatEvent(rawData)
		if err != nil {
			log.Error("Error deserializing chat event:", "err", err)
			continue
		}

		chatEvents <- deserializedChatEvent
	}
}
