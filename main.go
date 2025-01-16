package main

import (
	"fmt"
	"github.com/Stapxs/Stapxs-QQ-Shell/pages"
	tea "github.com/charmbracelet/bubbletea"
	"os/exec"
	"time"
)

func main() {
	_ = exec.Command("title", "Stapxs QQ Shell")

	p := tea.NewProgram(
		pages.InitialModel(),
		tea.WithAltScreen(),
	)

	// 每秒发送一个 tick 消息用于被动刷新页面
	go func() {
		for {
			time.Sleep(1 * time.Second)
			p.Send("tick")
		}
	}()

	if err := p.Start(); err != nil {
		fmt.Printf("程序运行出错: %v", err)
	}
}
