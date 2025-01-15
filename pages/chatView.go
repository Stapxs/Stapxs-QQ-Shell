package pages

import (
	"github.com/76creates/stickers/flexbox"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	style1 = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true).BorderForeground(lipgloss.Color("#fc5c65"))
	style2 = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true).BorderForeground(lipgloss.Color("#fd9644"))
	style3 = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true).BorderForeground(lipgloss.Color("#fed330"))
)

type ChatModel struct {
	flexBox *flexbox.FlexBox
}

func InitialChatModel() ChatModel {
	flexBox := flexbox.New(0, 0)
	rows := []*flexbox.Row{
		flexBox.NewRow().AddCells(
			flexbox.NewCell(3, 6).SetStyle(style1),
			flexbox.NewCell(6, 6).SetStyle(style2),
		),
		flexBox.NewRow().AddCells(
			flexbox.NewCell(2, 1).SetStyle(style3),
		),
	}

	flexBox.AddRows(rows)
	return ChatModel{
		flexBox: flexBox,
	}
}

func (m ChatModel) Init() tea.Cmd {
	m.flexBox.SetWidth(WindowWidth)
	m.flexBox.SetHeight(WindowHeight)
	return nil
}

func (m ChatModel) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.flexBox.SetWidth(msg.Width)
		m.flexBox.SetHeight(msg.Height)
		return m, nil
	default:
		return m, nil
	}
}

func (m ChatModel) View() (s string) {
	return m.flexBox.Render()
}
