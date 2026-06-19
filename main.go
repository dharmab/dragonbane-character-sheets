package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"dragonbane-char/internal/character"
	"dragonbane-char/internal/ui"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: dragonbane-char <character.json>\n")
		os.Exit(1)
	}
	path := os.Args[1]

	char, err := character.Load(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading %s: %v\n", path, err)
		os.Exit(1)
	}

	p := tea.NewProgram(ui.New(char, path), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
