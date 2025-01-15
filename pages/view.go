package pages

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type View interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (View, tea.Cmd)
	View() string
}

var CurrentView = "main"
var WindowWidth int = 0
var WindowHeight int = 0

type BaseModel struct {
	currentView string
	views       map[string]View
}

func InitialModel() BaseModel {
	views := map[string]View{
		"main": InitialMainModel(), // 加载视图
		"chat": InitialChatModel(), // 聊天视图
	}
	return BaseModel{
		currentView: CurrentView,
		views:       views,
	}
}

func (m BaseModel) Init() tea.Cmd {
	if view, exists := m.views[m.currentView]; exists {
		return view.Init()
	}
	return nil
}

func (m BaseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// 如果当前视图不等于当前视图, 则更新当前视图
	if CurrentView != m.currentView {
		m.currentView = CurrentView
		m.Init()
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		WindowWidth = msg.Width
		WindowHeight = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	if view, exists := m.views[m.currentView]; exists {
		updatedView, cmd := view.Update(msg)
		m.views[m.currentView] = updatedView
		return m, cmd
	}
	return m, nil
}

func (m BaseModel) View() string {
	if view, exists := m.views[m.currentView]; exists {
		return view.View()
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("载入入口视图失败 • q: exit")
}
