package pages

import (
	"fmt"
	"time"

	"github.com/76creates/stickers/flexbox"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	textStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("69")).Render
	timeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render

	tipStyle = lipgloss.NewStyle().Align(lipgloss.Right)
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
		timer:   timer.NewWithInterval(time.Second*1, time.Millisecond),
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
	str := fmt.Sprintf("%s%s%s%s",
		m.spinner.View(),
		" ",
		textStyle("正在加载..."),
		timeStyle(" ("+m.timer.View()+")"),
	)
	errStr := ""
	if m.timer.Timedout() {
		CurrentView = "chat"
	}
	m.flexBox.GetRow(1).GetCell(1).SetContent(str)
	m.flexBox.GetRow(1).GetCell(2).SetContent(errStr)
	return m.flexBox.Render()
}
