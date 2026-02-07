package tui

import (
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"wstogo/internal/controller"
	"wstogo/internal/events"
)

type Model struct {
	controller *controller.Controller
	width      int
	height     int
	addr       string
	path       string
	logs       []string
	incoming   []string
	input      string
	connected  bool
}

func NewModel(addr, path string) Model {
	c := controller.New()
	return Model{
		controller: c,
		addr:       addr,
		path:       path,
		logs:       []string{"Press 'c' to connect"},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "c":
			// always allow connect
			reqStr := m.controller.GetHandshakeReq(m.path)
			cleanReq := strings.ReplaceAll(reqStr, "\r", "")
			m.logs = append(m.logs,
				"---",
				"Connecting to "+m.addr+m.path,
				"---",
				cleanReq,
				"---",
			)
			if len(m.logs) > 20 {
				m.logs = m.logs[1:]
			}
			return m, connectCmd(m.addr, m.path, m.controller)
		case "enter":
			if m.connected && m.input != "" {
				m.controller.HandleEvent(events.MessageEvent{Msg: m.input})
				m.input = ""
			}
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		case "v":
			if m.connected {
				if out, err := exec.Command("xclip", "-o", "-selection", "clipboard").Output(); err == nil {
					m.input += strings.TrimSpace(string(out))
				}
			}
		default:
			if m.connected {
				m.input += msg.String()
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case events.LogEvent:
		// clean multi-line logs (e.g. HTTP with \r\n) for TUI display
		clean := strings.ReplaceAll(msg.Content, "\r", "")
		m.logs = append(m.logs, clean)
		if len(m.logs) > 20 {
			m.logs = m.logs[1:]
		}
	case events.HandshakeDone:
		// dedicated signal: set connected, log res, start listener
		m.logs = append(m.logs, msg.Res)
		if len(m.logs) > 20 {
			m.logs = m.logs[1:]
		}
		m.connected = true
		return m, m.controller.ListenCmd()
	case events.IncomingEvent:
		m.incoming = append(m.incoming, msg.Content)
		if len(m.incoming) > 10 {
			m.incoming = m.incoming[1:]
		}
		return m, m.controller.ListenCmd()
	}
	return m, nil
}

func (m Model) View() string {
	// 3-panel layout with lipgloss: left=logs/handshake, mid=incoming, right=input
	// adjust widths for borders; no title method in this lipgloss ver, use content
	w := m.width / 3
	leftStyle := lipgloss.NewStyle().
		Width(w).
		Height(m.height - 4).
		Border(lipgloss.RoundedBorder())
	midStyle := lipgloss.NewStyle().
		Width(w).
		Height(m.height - 4).
		Border(lipgloss.RoundedBorder())
	rightStyle := lipgloss.NewStyle().
		Width(w).
		Height(m.height - 4).
		Border(lipgloss.RoundedBorder())

	// left panel: logs (with title in content)
	leftContent := "Handshake/Process\n---\n" + strings.Join(m.logs, "\n")
	left := leftStyle.Render(leftContent)

	// mid: incoming
	midContent := strings.Join(m.incoming, "\n")
	if len(m.incoming) == 0 {
		midContent = "no messages yet"
	}
	midContent = "Incoming WS Data\n---\n" + midContent
	mid := midStyle.Render(midContent)

	// right: input
	rightContent := "Send (Enter to send, v=paste)\n---\n" + m.input
	right := rightStyle.Render(rightContent)

	// join horizontal
	return lipgloss.JoinHorizontal(lipgloss.Top, left, mid, right)
}

func connectCmd(addr, path string, ctrl *controller.Controller) tea.Cmd {
	return func() tea.Msg {
		err, resStr := ctrl.Connect(addr, path)
		if err != nil {
			return events.LogEvent{Content: "Connect error: " + err.Error()}
		}
		// dedicated signal with res for log
		cleanRes := strings.ReplaceAll(resStr, "\r", "")
		return events.HandshakeDone{Res: cleanRes}
	}
}
