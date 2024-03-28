package main

import (
	"encoding/gob"
	"flag"
	"go-chat-cli/client"
	"go-chat-cli/message"
	"go-chat-cli/server"

	log "github.com/charmbracelet/log"
)

func main() {
	gob.Register(message.Message{})
	log.SetLevel(log.DebugLevel)
	var username string
	var dialAddress string
	var isServer bool
	flag.StringVar(&dialAddress, "address", "", "Address to dial")
	flag.StringVar(&username, "username", "", "Username to use")
	flag.BoolVar(&isServer, "server", false, "Run as server")
	flag.Parse()

	if isServer {
		server.Server()
		return
	}

	if dialAddress == "" {
		log.Error("Please provide an address to dial")
		return
	}

	if username == "" {
		log.Error("Please provide a username")
		return
	}
	client.Client(username, dialAddress)

}
