package pages

import (
	"fmt"
	"github.com/Stapxs/Stapxs-QQ-Shell/utils"
	"time"

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
)

type MainModel struct {
	flexBox   *flexbox.FlexBox
	spinner   spinner.Model
	timer     timer.Model
	errorText string
}

func InitialMainModel() MainModel {
	spinnerStyle := lipgloss.NewStyle().Foreground(mainColor)
	newSpinner := spinner.New()
	newSpinner.Style = spinnerStyle
	newSpinner.Spinner = spinner.Line

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

	return MainModel{
		flexBox: flexBox,
		spinner: newSpinner,
		timer:   timer.NewWithInterval(time.Second*3, time.Millisecond),
		errorText: errorTextList[utils.RandInt(0, len(errorTextList)-1)],
	}
}

func (m MainModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.timer.Init())
}

func (m MainModel) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.flexBox.SetWidth(msg.Width)
		m.flexBox.SetHeight(msg.Height)
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case timer.TickMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m MainModel) View() (s string) {
	textStyle := lipgloss.NewStyle().Foreground(mainColor).Render

	str := fmt.Sprintf("%s%s%s",
		m.spinner.View(),
		" ",
		textStyle("初始化中..."),
	)
	if utils.ErrorMsg == "" && m.timer.Timedout() {
		utils.CurrentView = "chat"
	}

	errorFbStyleTitle := lipgloss.NewStyle().Foreground(mainReverseFontColor).Background(mainColor).Margin(1, 0, 0, 3).Padding(0, 1).Render
	errorFbStyle := lipgloss.NewStyle().Foreground(mainFontColor).Margin(1, 0, 0, 3).Render
	errorStyleTitle := lipgloss.NewStyle().Foreground(mainReverseFontColor).Background(lipgloss.Color("#fc5c65")).Margin(1, 0, 0, 3).Padding(0, 1).Render
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#fc5c65")).Margin(1, 0, 0, 3).Render
	// 此视图有时候也用于全局错误提示
	if utils.ErrorMsg == "" {
		m.flexBox.GetRow(1).GetCell(1).SetContent(helpStyle(" • ctrl+c: 快速退出 • "))
		m.flexBox.GetRow(1).GetCell(2).SetContent(str)
	} else {
		m.flexBox.GetRow(1).GetCell(1).SetContent(helpStyle(" • ctrl+c: 退出 • "))
		m.flexBox.GetRow(1).GetCell(2).SetContent(helpStyle("> " + utils.ErrorMsg))
		m.flexBox.GetRow(0).GetCell(0).SetContent("" +
			errorFbStyleTitle("严重异常中断") +
			errorFbStyle(">> 程序发生了一个严重的异常，"+m.errorText+" <<") +
			errorFbStyle("你可以将此页面截图或复制提交至仓库 issue 来让 Stapxs QQ Shell 变得更好！\n") +
			errorFbStyle("        > ") +
			baseLink.Render("https://github.com/Stapxs/Stapxs-QQ-Shell") + "\n" +
			errorStyleTitle("摘要") +
			errorStyle(":: "+utils.ErrorMsg+"\n"+utils.ErrorFullTrace),
		)
	}
	return m.flexBox.Render()
}
