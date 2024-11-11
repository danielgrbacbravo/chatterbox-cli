package serialization

import (
	pb "chatterbox-cli/proto"

	"google.golang.org/protobuf/proto"
)

func DeserializeChatEvent(data []byte) (*pb.ChatEvent, error) {
	var event pb.ChatEvent
	// Unmarshal the binary data into a ChatEvent object
	err := proto.Unmarshal(data, &event)
	if err != nil {
		return nil, err
	}
	return &event, nil
}
