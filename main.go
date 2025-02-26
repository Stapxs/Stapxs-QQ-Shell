package main

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/Stapxs/Stapxs-QQ-Shell/utils"
	"github.com/Stapxs/Stapxs-QQ-Shell/view"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	_ = exec.Command("title", "Stapxs QQ Shell")

	p := tea.NewProgram(
		view.InitialModel(),
		tea.WithAltScreen(),
	)

	// 页面数据刷新外部信号
	go func() {
		for {
			time.Sleep(1 * time.Millisecond * 500)
			p.Send(utils.UpdateMsg{})
		}
	}()

	_, err := p.Run()
	if err != nil {
		fmt.Printf("程序运行出错: %v", err)
	}
}
