package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"wstogo/internal/tui"
)

var (
	addr = flag.String("addr", "localhost:8080", "server address")
	path = flag.String("path", "/echo", "websocket path")
)

func main() {
	flag.Parse()
	m := tui.NewModel(*addr, *path)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
