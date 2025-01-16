package pages

import (
	"github.com/Stapxs/Stapxs-QQ-Shell/utils"
	"runtime/debug"
	"sort"

	"github.com/76creates/stickers/flexbox"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type View interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (View, tea.Cmd)
	View() string
}

var WindowWidth = 0
var WindowHeight = 0

// 全局样式
var (
	mainColor            = lipgloss.AdaptiveColor{Light: "#636e79", Dark: "#cee4fc"}
	mainColorDark        = lipgloss.AdaptiveColor{Light: "#394046", Dark: "#96a6b7"}
	mainFontColor        = lipgloss.AdaptiveColor{Light: "#51534f", Dark: "#ffffff"}
	mainReverseFontColor = lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#3a3a3a"}

	baseLink = lipgloss.NewStyle().Foreground(mainColor).Underline(true)

	testStyle = lipgloss.NewStyle().Background(mainColor)
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
		currentView: utils.CurrentView,
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
	if utils.CurrentView != m.currentView {
		m.currentView = utils.CurrentView
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
			utils.CurrentView = "main"
			utils.ErrorMsg = "渲染视图 " + m.currentView + "View 异常"
			filteredStack := utils.FilterStack(debug.Stack(), "github.com/Stapxs/Stapxs-QQ-Shell")
			utils.ErrorFullTrace = filteredStack
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
	// 直接遍历顺序会随机变化，先进行排序
	keys := make([]string, 0, len(control))
	for key := range control {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		controlStr += key + " " + control[key] + " • "
	}
	controlRow.GetCell(1).SetContent(helpStyle(controlStr))
	if len(errorStr) > 0 {
		controlRow.GetCell(2).SetContent(helpStyle("> " + errorStr[0]))
	}
}
