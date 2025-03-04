package main

import (
	"fmt"
	"os/exec"
	"runtime/debug"
	"time"

	"github.com/Stapxs/Stapxs-QQ-Shell/utils"
	"github.com/Stapxs/Stapxs-QQ-Shell/utils/runtime"
	"github.com/Stapxs/Stapxs-QQ-Shell/view"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// 判断是否为调试模式，使用 go run 命令运行时会自动开启调试模式
	if info, ok := debug.ReadBuildInfo(); ok && info.Path != "" {
		runtime.Debug = true
	}
	// 设置窗口
	_ = exec.Command("title", "Stapxs QQ Shell")
	// 初始化程序
	p := tea.NewProgram(
		view.InitialModel(),
		tea.WithAltScreen(),
	)
	go func() {
		// 每 500 毫秒发送一次更新消息，用于刷新视图
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
