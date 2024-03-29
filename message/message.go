package message

import (
	"encoding/gob"
	"net"
)

// Message is a struct that represents a message
type Message struct {
	Username    string
	Message     string
	MessageType string
}

// send message to a connection

func (m *Message) SendMessage(conn net.Conn) error {
	enc := gob.NewEncoder(conn)
	return enc.Encode(m)
}

func ReadMessage(conn net.Conn) (Message, error) {
	dec := gob.NewDecoder(conn)
	var msg Message
	err := dec.Decode(&msg)
	return msg, err
}
