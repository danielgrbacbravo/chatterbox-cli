package login

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	log "github.com/charmbracelet/log"
)

var username string
var dialAddress string

func FetchLoginData() (string, string) {
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

	return username, dialAddress
}

type (
	errMsg error
)

const (
	ip   = iota
	user = iota
)

const (
	hotPink  = lipgloss.Color("#FF06B7")
	darkGray = lipgloss.Color("#767676")
)

var (
	inputStyle    = lipgloss.NewStyle().Foreground(hotPink)
	continueStyle = lipgloss.NewStyle().Foreground(darkGray)
)

type model struct {
	inputs  []textinput.Model
	focused int
	err     error
}

func initialModel() model {
	var inputs []textinput.Model = make([]textinput.Model, 2)
	// username field
	inputs[user] = textinput.New()
	inputs[user].Placeholder = "Username"
	inputs[user].CharLimit = 20
	inputs[user].Width = 30
	inputs[user].Prompt = ""
	// IP address field
	inputs[ip] = textinput.New()
	inputs[ip].Focus()
	inputs[ip].Placeholder = "192.168.10.5"
	inputs[ip].CharLimit = 15
	inputs[ip].Width = 30
	inputs[ip].Prompt = ""

	return model{
		inputs:  inputs,
		focused: 0,
		err:     nil,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd = make([]tea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.focused == len(m.inputs)-1 {
				// set username and dial address
				username = m.inputs[user].Value()
				dialAddress = m.inputs[ip].Value()
				return m, tea.Quit
			}
			m.nextInput()
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyShiftTab, tea.KeyCtrlP:
			m.prevInput()
		case tea.KeyTab, tea.KeyCtrlN:
			m.nextInput()
		}

		for i := range m.inputs {
			if i == m.focused {
				m.inputs[i].Focus()
			} else {
				m.inputs[i].Blur()
			}
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	for i := range m.inputs {
		if i == m.focused {
			m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
		}
	}
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	return fmt.Sprintf(
		`
 %s
 %s

 %s
 %s

 %s
`,
		inputStyle.Width(30).Render("Username"),
		m.inputs[user].View(),
		inputStyle.Width(30).Render("IP Address"),
		m.inputs[ip].View(),
		continueStyle.Render("Continue ->"),
	) + "\n"
}

// nextInput focuses the next input field
func (m *model) nextInput() {
	m.focused = (m.focused + 1) % len(m.inputs)
}

// prevInput focuses the previous input field
func (m *model) prevInput() {
	m.focused--
	// Wrap around
	if m.focused < 0 {
		m.focused = len(m.inputs) - 1
	}
}
