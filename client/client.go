package client

import (
	"chatterbox-cli/crypto"
	"chatterbox-cli/message"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"net"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	log "github.com/charmbracelet/log"
)

var conn net.Conn
var publicKey ecdsa.PublicKey
var privateKey *ecdsa.PrivateKey

// server public key
var serverPublicKey ecdsa.PublicKey

func Client(username, dialAddress string) {

	// construct private key
	privateKey = crypto.GeneratePrivateKey()
	publicKey = crypto.GeneratePublicKey(privateKey)

	conn, err := net.Dial("tcp", dialAddress)
	if err != nil {
		return
	}
	defer conn.Close()

	initialModel := initialModel()
	initialModel.username = username
	initialModel.conn = conn

	// begin aes encryption handshake
	sendJoinMessage(conn, username)

	programChan := make(chan *tea.Program, 1) // create a channel to pass the Bubbletea program

	go listenForMessages(conn, programChan) // pass the channel to the goroutine

	p := tea.NewProgram(initialModel)
	programChan <- p // send the Bubbletea program over the channel

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func listenForMessages(conn net.Conn, programChan chan *tea.Program) {
	p := <-programChan // receive the Bubbletea program from the channel
	for {
		msg, err := message.ReadMessage(conn)
		if err != nil {
			return
		}
		p.Send(incomingMsg(msg))
	}
}

type (
	errMsg error
)

type incomingMsg message.Message

type model struct {
	viewport    viewport.Model
	username    string
	messages    []string
	textarea    textarea.Model
	senderStyle lipgloss.Style
	err         error
	conn        net.Conn
	serverKey   ecdsa.PublicKey
	sharedKey   *big.Int
}

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(60, 10)
	vp.SetContent(`Welcome to the chat room!
Type a message and press Enter to send.`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return model{
		textarea:    ta,
		username:    "",
		messages:    []string{},
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:         nil,
		conn:        nil,
		serverKey:   ecdsa.PublicKey{},
		sharedKey:   nil,
	}
}
func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case incomingMsg:
		if msg.MessageType == "public_key" {
			// server has sent public key
			// save it to the model
			m.serverKey = crypto.ConvertToPublicKey(message.Message(msg))
			crypto.SendPublicKey(m.conn, publicKey)
			// condition at current state
			// both client and server have exchanged public keys
			m.sharedKey = crypto.GenerateSharedSecret(privateKey, m.serverKey)
			// diffie hellman key exchange complete
			break
		}
		// decrypt message
		decryptedMessage := crypto.DecryptMessage(message.Message(msg), m.sharedKey)

		// check if message is a join message
		if decryptedMessage.MessageType == "join" {
			m.messages = append(m.messages, m.senderStyle.Render(decryptedMessage.Username+" joined the chat"))
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()
			break
		}
		// check if message is a leave message
		if decryptedMessage.MessageType == "leave" {
			m.messages = append(m.messages, m.senderStyle.Render(decryptedMessage.Username+" left the chat"))
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()
			break
		}
		// check if the message is from the current user
		if decryptedMessage.Username == m.username {
			break
		}
		// add message to the model
		m.messages = append(m.messages, m.senderStyle.Render(decryptedMessage.Username+": ")+decryptedMessage.Message)
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.textarea.Reset()
		m.viewport.GotoBottom()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			msg := message.Message{Username: m.username, Message: m.textarea.Value(), MessageType: "message"}
			if m.conn == nil {
				log.Error("Connection is nil, cannot send message")
				break
			}
			// add message to the model
			m.messages = append(m.messages, m.senderStyle.Render("You: ")+m.textarea.Value())
			// encrypt message
			encryptedMsg := crypto.EncyptMessage(msg, m.sharedKey)
			encryptedMsg.SendMessage(m.conn)
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.textarea.Reset()
			m.viewport.GotoBottom()
			// make sure that msg is not type message.messages after this case
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s",
		m.viewport.View(),
		m.textarea.View(),
	) + "\n\n"
}

func sendJoinMessage(conn net.Conn, username string) {
	msg := message.Message{Username: username, MessageType: "join"}
	msg.SendMessage(conn)
}
