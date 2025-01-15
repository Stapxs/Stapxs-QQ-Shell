package pages

import (
	"github.com/76creates/stickers/flexbox"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"runtime/debug"
	"strings"
)

type View interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (View, tea.Cmd)
	View() string
}

var CurrentView = "main"
var WindowWidth = 0
var WindowHeight = 0
var ErrorMsg = ""
var ErrorFullTrace = ""

// 全局样式
var (
	mainColor = lipgloss.Color("#636e79")

	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	textStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("69")).Render
	// 控制指示器样式
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
	tipStyle  = lipgloss.NewStyle().Align(lipgloss.Right)
)

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
		case "ctrl+c":
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
	defer func() string {
		if r := recover(); r != nil {
			CurrentView = "main"
			ErrorMsg = "渲染视图 " + m.currentView + "View 异常"
			filteredStack := filterStack(debug.Stack(), "github.com/Stapxs/Stapxs-QQ-Shell")
			ErrorFullTrace = filteredStack
		}
		return ""
	}()
	if view, exists := m.views[m.currentView]; exists {
		return view.View()
	}
	return "载入入口视图失败"
}

// ========================================

func SetControlBar(flexBox *flexbox.FlexBox, control map[string]string, errorStr ...string) {
	// flexBox 的最后一行的第二列是控制指示器，第三列是错误信息
	controlRow := flexBox.GetRow(flexBox.RowsLen() - 1)
	// 判断它有没有三列
	if controlRow.CellsLen() < 3 {
		return
	}
	// 拼接控制指示器
	controlStr := " • "
	for key, value := range control {
		controlStr += key + ": " + value + " • "
	}
	controlRow.GetCell(1).SetContent(helpStyle(controlStr))
	if len(errorStr) > 0 {
		controlRow.GetCell(2).SetContent(helpStyle("> " + errorStr[0]))
	}
}

// ========================================

func filterStack(stack []byte, packageName string) string {
	lines := strings.Split(string(stack), "\n")
	var filteredLines []string
	for _, line := range lines {
		if strings.Contains(line, packageName) {
			line = strings.Replace(line, packageName+"/", "", 1)
			filteredLines = append(filteredLines, line)
		}
	}
	return strings.Join(filteredLines, "\n")
}
