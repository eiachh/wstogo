package controller

import (
	tea "github.com/charmbracelet/bubbletea"
	"wstogo/internal/events"
	"wstogo/internal/http"
	"wstogo/internal/ws"
)

type Controller struct {
	wsConn   *ws.WSConn
	incoming chan string
}

func New() *Controller {
	return &Controller{
		incoming: make(chan string, 10),
	}
}

func (c *Controller) Connect(addr, path string) (error, string) {
	c.wsConn = &ws.WSConn{}
	err, resStr := c.wsConn.Handshake(addr, path)
	if err != nil {
		return err, ""
	}
	go c.readLoop()
	return nil, resStr
}

func (c *Controller) HandleEvent(e events.Event) {
	switch e := e.(type) {
	case events.MessageEvent:
		if c.wsConn != nil {
			_ = c.wsConn.WriteMessage(e.Msg)
		}
	}
}

// GetHandshakeReq builds full HTTP req for logging in TUI.
func (c *Controller) GetHandshakeReq(path string) string {
	req := http.NewRequest("GET", path, "HTTP/1.1", nil)
	req.AddHeader("Host", "localhost") // placeholder
	req.AddHeader("Upgrade", "websocket")
	req.AddHeader("Connection", "Upgrade")
	req.AddHeader("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==") // example
	req.AddHeader("Sec-WebSocket-Version", "13")
	req.AddHeader("Origin", "http://localhost")
	return req.String()
}

func (c *Controller) readLoop() {
	for {
		if c.wsConn == nil {
			break
		}
		msg, err := c.wsConn.ReadMessage()
		if err != nil {
			break
		}
		c.incoming <- msg
	}
}

func (c *Controller) ListenCmd() tea.Cmd {
	return func() tea.Msg {
		msg := <-c.incoming
		return events.IncomingEvent{Content: msg}
	}
}
