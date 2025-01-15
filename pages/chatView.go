package pages

import (
	"strings"

	"github.com/76creates/stickers/flexbox"
	"github.com/Stapxs/Stapxs-QQ-Shell/utils/websocket"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var WebSocketClient *websocket.Client

type ChatModel struct {
	errorStr string
	flexBox  *flexbox.FlexBox
	// 输入框
	inputs     []textinput.Model
	focusIndex int
}

// Websocket 消息类型
type wsConnectedMsg struct{}
type wsErrorMsg struct {
	err error
}
type wsMessageMsg struct {
	message string
}

func InitialChatModel() ChatModel {
	// 基础布局
	flexBox := flexbox.New(30, 10)
	rows := []*flexbox.Row{
		flexBox.NewRow().AddCells(
			flexbox.NewCell(15, 10),
			flexbox.NewCell(30, 10),
		),
		flexBox.NewRow().AddCells(
			flexbox.NewCell(1, 1),
			flexbox.NewCell(14, 1),
			flexbox.NewCell(10, 1).SetStyle(tipStyle),
			flexbox.NewCell(1, 1),
		),
	}
	flexBox.AddRows(rows)
	// 输入框
	inputs := make([]textinput.Model, 2)
	var t textinput.Model
	for i := range inputs {
		t = textinput.New()
		switch i {
		case 0:
			t.Placeholder = "连接地址"
			t.Focus()
		case 1:
			t.Placeholder = "连接密钥"
		}
		inputs[i] = t
	}
	return ChatModel{
		errorStr: "未连接",
		flexBox:  flexBox,
		inputs:   inputs,
	}
}

func (m ChatModel) Init() tea.Cmd {
	m.flexBox.SetWidth(WindowWidth)
	m.flexBox.SetHeight(WindowHeight)
	return textinput.Blink
}

func (m ChatModel) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.flexBox.SetWidth(msg.Width)
		m.flexBox.SetHeight(msg.Height)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "enter", "up", "down":
			if WebSocketClient == nil {
				s := msg.String()
				if s == "up" || s == "shift+tab" {
					m.focusIndex--
				} else {
					m.focusIndex++
				}

				if m.focusIndex > len(m.inputs) {
					m.focusIndex = 0
				} else if m.focusIndex < 0 {
					m.focusIndex = len(m.inputs)
				}

				// 如果是最后一个输入框，按下 enter 键则连接到 OneBot
				if m.focusIndex == len(m.inputs) {
					if WebSocketClient != nil {
						WebSocketClient.Close()
					}
					address := m.inputs[0].Value() + "?access_token=" + m.inputs[1].Value()
					WebSocketClient = websocket.NewClient(address)
					if err := WebSocketClient.Connect(); err != nil {
						WebSocketClient = nil
						m.errorStr = "连接失败: " + err.Error()
						m.focusIndex = 0
					}
				}

				cmdList := make([]tea.Cmd, len(m.inputs))
				for i := 0; i <= len(m.inputs)-1; i++ {
					if i == m.focusIndex {
						cmdList[i] = m.inputs[i].Focus()
						continue
					}
					m.inputs[i].Blur()
				}

				return m, tea.Batch(cmdList...)
			}
		}
	}

	cmdList := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmdList[i] = m.inputs[i].Update(msg)
	}

	return m, tea.Batch(cmdList...)
}

func (m ChatModel) View() (s string) {
	mainGrid := m.flexBox.GetRow(0).GetCell(1)

	if WebSocketClient == nil {
		var b strings.Builder
		for i := range m.inputs {
			b.WriteString(m.inputs[i].View())
			if i < len(m.inputs)-1 {
				b.WriteRune('\n')
			}
		}
		titleStyle := lipgloss.NewStyle().Background(mainColor).Foreground(lipgloss.Color("#ffffff")).Render
		// 新建一个
		mainGrid.SetContent(titleStyle(" 连接到 OneBot ") + "\n\n" + b.String())
		mainGrid.SetStyle(lipgloss.NewStyle().Align(lipgloss.Center).AlignVertical(lipgloss.Center))
		SetControlBar(m.flexBox, map[string]string{
			"ctrl+c": "退出",
			"enter":  "连接",
		}, m.errorStr)
	}
	return m.flexBox.Render()
}
