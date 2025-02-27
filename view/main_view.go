package view

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/Stapxs/Stapxs-QQ-Shell/utils/runtime"

	"github.com/charmbracelet/lipgloss"

	"github.com/76creates/stickers/flexbox"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	errorTextList = []string{
		"Bug 已经被打包到笨蛋罐头里！",
		"呜呜呜，完蛋了完蛋了……",
		"诶嘿！",
		"看不见看不见！",
	}
	errorText = errorTextList[rand.Intn(len(errorTextList))]
)

// MainModel 主视图元素
type MainModel struct {
	flexBox *flexbox.FlexBox // 布局框架
	spinner spinner.Model    // 加载指示器
	timer   timer.Model      // 加载计时器
}

// InitialMainModel
// @Description: 初始化主视图
// @return *MainModel
func InitialMainModel() *MainModel {
	// 初始化布局
	flexBox := flexbox.New(30, 10)
	rows := []*flexbox.Row{
		flexBox.NewRow().AddCells(
			flexbox.NewCell(1, 10),
		),
		flexBox.NewRow().AddCells(
			flexbox.NewCell(1, 1),
			flexbox.NewCell(14, 1),
			flexbox.NewCell(15, 1).SetStyle(tipStyle),
			flexbox.NewCell(1, 1),
		),
	}
	flexBox.AddRows(rows)
	// 初始化加载指示器
	newSpinner := spinner.New()
	newSpinner.Style = lipgloss.NewStyle().Foreground(mainColor)
	newSpinner.Spinner = spinner.Line

	return &MainModel{
		flexBox: flexBox,
		spinner: newSpinner,
		timer:   timer.NewWithInterval(time.Second*3, time.Millisecond),
	}
}

func (m MainModel) Init() tea.Cmd {
	// 启动加载指示器
	return tea.Batch(m.spinner.Tick, m.timer.Init())
}

func (m MainModel) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.flexBox.SetWidth(msg.Width)
		m.flexBox.SetHeight(msg.Height)
		return &m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return &m, cmd
	case timer.TickMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return &m, cmd
	default:
		return &m, nil
	}
}

func (m MainModel) View() string {
	// 渲染加载视图
	str := fmt.Sprintf("%s%s%s",
		m.spinner.View(),
		" ",
		lipgloss.NewStyle().Foreground(mainColor).Render("初始化中..."),
	)

	// 如果存在 ErrorMsg，此视图将作为错误页使用
	// 渲染错误页视图
	if runtime.ErrorMsg == "" && m.timer.Timedout() {
		runtime.CurrentView = "chat"
	}

	errorFbStyleTitle := lipgloss.NewStyle().Foreground(mainReverseFontColor).Background(mainColor).Margin(1, 0, 0, 3).Padding(0, 1).Render
	errorFbStyle := lipgloss.NewStyle().Foreground(mainFontColor).Margin(1, 0, 0, 3).Render
	errorStyleTitle := lipgloss.NewStyle().Foreground(mainReverseFontColor).Background(lipgloss.Color("#fc5c65")).Margin(1, 0, 0, 3).Padding(0, 1).Render
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#fc5c65")).Margin(1, 0, 0, 3).Render

	if runtime.ErrorMsg == "" {
		m.flexBox.GetRow(1).GetCell(1).SetContent(helpStyle(" • ctrl+c: 快速退出 • "))
		m.flexBox.GetRow(1).GetCell(2).SetContent(str)
		var title = "" +
			" _____ _                  _____ _____    _____ _       _ _  \n" +
			"|   __| |_ ___ ___ _ _   |     |     |  |   __| |_ ___| | | \n" +
			"|__   |  _| .'| . |_'_|  |  |  |  |  |  |__   |   | -_| | | \n" +
			"|_____|_| |__,|  _|_,_|  |__  _|__  _|  |_____|_|_|___|_|_| \n" +
			"              |_|           |__|  |__|                        "
		m.flexBox.GetRow(0).GetCell(0).SetStyle(lipgloss.NewStyle().Align(lipgloss.Center).AlignVertical(lipgloss.Center).Foreground(mainColor))
		m.flexBox.GetRow(0).GetCell(0).SetContent(
			lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Render(title),
		)
	} else {
		m.flexBox.GetRow(1).GetCell(1).SetContent(helpStyle(" • ctrl+c: 退出 • "))
		m.flexBox.GetRow(1).GetCell(2).SetContent(helpStyle("> " + runtime.ErrorMsg))
		m.flexBox.GetRow(0).GetCell(0).SetContent("" +
			errorFbStyleTitle("严重异常中断") +
			errorFbStyle(">> 程序发生了一个严重的异常，"+errorText+" <<") +
			errorFbStyle("你可以将此页面截图或复制提交至仓库 issue 来让 Stapxs QQ Shell 变得更好！\n") +
			errorFbStyle("        > ") +
			baseLink.Render("https://github.com/Stapxs/Stapxs-QQ-Shell") + "\n" +
			errorStyleTitle("摘要") +
			errorStyle(":: "+runtime.ErrorMsg+"\n"+runtime.ErrorFullTrace),
		)
	}
	return m.flexBox.Render()
}
