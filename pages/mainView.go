package pages

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/76creates/stickers/flexbox"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
)

type MainModel struct {
	flexBox *flexbox.FlexBox
	spinner spinner.Model
	timer   timer.Model
}

func InitialMainModel() MainModel {
	newSpinner := spinner.New()
	newSpinner.Style = spinnerStyle
	newSpinner.Spinner = spinner.Pulse

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
		timer:   timer.NewWithInterval(time.Second*2, time.Millisecond),
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
	str := fmt.Sprintf("%s%s%s",
		m.spinner.View(),
		" ",
		textStyle("初始化中..."),
	)
	if ErrorMsg == "" && m.timer.Timedout() {
		CurrentView = "chat"
	}

	errorStyleTitle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Background(lipgloss.Color("#fc5c65")).Margin(1, 0, 0, 3).Render
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#fc5c65")).Margin(1, 0, 0, 3).Render
	// 此视图有时候也用于全局错误提示
	if ErrorMsg == "" {
		m.flexBox.GetRow(1).GetCell(1).SetContent(str)
	} else {
		m.flexBox.GetRow(1).GetCell(1).SetContent(helpStyle(" • ctrl+c: 退出 • "))
		m.flexBox.GetRow(1).GetCell(2).SetContent(helpStyle("> " + ErrorMsg))
		m.flexBox.GetRow(0).GetCell(0).SetContent(errorStyleTitle("错误摘要") + errorStyle(ErrorFullTrace))
	}
	return m.flexBox.Render()
}
