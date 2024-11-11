package parser

import (
	pb "chatterbox-cli/proto"
	"encoding/json"
)

func ParseUserFromJson(jsonInput []byte) (*pb.User, error) {
	var user pb.User
	json.Unmarshal(jsonInput, &user)
	return &user, nil
}
