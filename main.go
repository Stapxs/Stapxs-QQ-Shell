package main

import (
	"fmt"

	"github.com/Stapxs/Stapxs-QQ-Shell/pages"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(pages.InitialModel(), tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Printf("程序运行出错: %v", err)
	}
}
