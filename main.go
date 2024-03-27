package main

import (
	"encoding/gob"
	"flag"
	"go-chat-cli/client"
	"go-chat-cli/server"

	log "github.com/charmbracelet/log"
)

type Messsage struct {
	Username string
	Message  string
}

func main() {
	gob.Register(Messsage{})
	log.SetLevel(log.DebugLevel)
	var username string
	var dialAddress string
	flag.StringVar(&dialAddress, "address", "", "Address to dial")
	flag.StringVar(&username, "username", "", "Username to use")
	flag.Parse()
	if dialAddress != "" {
		log.Debug("Dialing to address: ", "addr", dialAddress)
	}

	if dialAddress == "" {
		server.Server()
	} else {
		client.Client(username, dialAddress)
	}
}
