package serialization

import (
	pb "chatterbox-cli/proto"

	"google.golang.org/protobuf/proto"
)

// serializeChatEvent
func SerializeChatEvent(event *pb.ChatEvent) ([]byte, error) {
	// Marshal the protobuf to binary format
	serializedData, err := proto.Marshal(event)
	if err != nil {
		return nil, err
	}
	return serializedData, nil
}
