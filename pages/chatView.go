package pages

import (
	"bytes"
	"github.com/Stapxs/Stapxs-QQ-Shell/utils"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/mdp/qrterminal/v3"
	"github.com/sahilm/fuzzy"
	"sort"
	"strings"

	"github.com/76creates/stickers/flexbox"
	"github.com/Stapxs/Stapxs-QQ-Shell/utils/websocket"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type userItem struct {
	title       string
	filterTitle string
	desc        string
	id          float64
}

func (i userItem) Title() string       { return i.title }
func (i userItem) FilterTitle() string { return i.filterTitle }
func (i userItem) Description() string { return i.desc }
func (i userItem) Id() float64         { return i.id }
func (i userItem) FilterValue() string { return i.filterTitle }

type msgItem struct {
	title      string
	desc       string
	id         float64
	rawMessage []interface{}
}

func (i msgItem) Title() string             { return i.title }
func (i msgItem) Description() string       { return i.desc }
func (i msgItem) Id() float64               { return i.id }
func (i msgItem) RawMessage() []interface{} { return i.rawMessage }
func (i msgItem) FilterValue() string       { return i.desc }

type ChatModel struct {
	tipStr            string            // 提示信息
	flexBox           *flexbox.FlexBox  // 基础布局
	inputs            []textinput.Model // 登录输入框
	focusIndex        int
	list              list.Model // 好友列表
	msgViewList       list.Model // 消息列表
	appendControlList map[string]string
	viewImage         string
	sendInput         textinput.Model // 发送消息输入框
}

var WebSocketClient *websocket.Client

var (
	listStyle            = lipgloss.NewStyle().Margin(1, 2)
	pointStatue          = "login"
	delegateItemList     = list.NewDefaultDelegate()
	delegateItemListDark = list.NewDefaultDelegate() // 失去焦点
	delegateItemMsg      = list.NewDefaultDelegate()
	delegateItemMsgDark  = list.NewDefaultDelegate() // 消息列表（失去焦点）
)

func InitialChatModel() ChatModel {
	// 基础布局
	flexBox := flexbox.New(30, 10)
	rows := []*flexbox.Row{
		flexBox.NewRow().AddCells(
			flexbox.NewCell(17, 10),
			flexbox.NewCell(30, 10),
		),
		flexBox.NewRow().AddCells(
			flexbox.NewCell(1, 1),
			flexbox.NewCell(19, 1),
			flexbox.NewCell(5, 1).SetStyle(tipStyle),
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
			t.EchoMode = textinput.EchoPassword
		}
		t.Prompt = ""
		inputs[i] = t
	}
	var tMsg textinput.Model
	tMsg = textinput.New()
	tMsg.Placeholder = "发送……"
	tMsg.Prompt = ""
	// 好友列表
	selectTitleStyle := lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(mainColor).Foreground(mainColor).Padding(0, 0, 0, 1)
	selectTitleStyleDark := lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(mainFontColor).Foreground(mainFontColor).Padding(0, 0, 0, 1)
	selectTitleStyleMsg := lipgloss.NewStyle().BorderForeground(mainColor).Foreground(mainColor).Padding(0, 0, 0, 2)
	selectTitleStyleMsgDark := lipgloss.NewStyle().BorderForeground(mainFontColor).Foreground(mainFontColor).Padding(0, 0, 0, 1)
	delegateItemList.Styles.SelectedTitle = selectTitleStyle
	delegateItemList.Styles.SelectedDesc = selectTitleStyle.Foreground(mainColorDark)
	delegateItemListDark.Styles.SelectedTitle = selectTitleStyleDark
	delegateItemListDark.Styles.SelectedDesc = selectTitleStyleDark.Foreground(mainFontColor)
	delegateItemMsg.Styles.SelectedTitle = selectTitleStyleMsg
	delegateItemMsg.Styles.SelectedDesc = selectTitleStyleMsg.Foreground(mainColorDark)
	delegateItemMsgDark.Styles.SelectedTitle = selectTitleStyleMsgDark
	delegateItemMsgDark.Styles.SelectedDesc = selectTitleStyleMsgDark.Foreground(mainFontColor)
	userList := list.New([]list.Item{userItem{title: "", desc: ""}}, delegateItemList, 0, 0)
	userList.Title = "用户列表"
	userList.FilterInput.Prompt = "搜索："
	userList.Filter = UserFilter
	userList.Styles.Title = titleStyle
	userList.SetShowStatusBar(false)
	userList.SetShowHelp(false)
	setupListKey(&userList)
	// 消息列表
	msgList := list.New([]list.Item{msgItem{title: "", desc: ""}}, delegateItemMsg, 0, 0)
	msgList.Title = "消息"
	msgList.Styles.Title = titleStyle
	msgList.SetShowStatusBar(false)
	msgList.SetShowHelp(false)
	msgList.SetShowFilter(false)
	msgList.SetShowPagination(false)
	cleanListKey(&msgList)
	return ChatModel{
		tipStr:      "未连接",
		flexBox:     flexBox,
		inputs:      inputs,
		list:        userList,
		msgViewList: msgList,
		sendInput:   tMsg,
		viewImage:   "",
	}
}

func (m ChatModel) Init() tea.Cmd {
	m.flexBox.SetWidth(WindowWidth)
	m.flexBox.SetHeight(WindowHeight)
	return textinput.Blink
}

func (m ChatModel) Update(msg tea.Msg) (View, tea.Cmd) {
	m.appendControlList = make(map[string]string)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.flexBox.SetWidth(msg.Width)
		m.flexBox.SetHeight(msg.Height)
		return m, nil
	case tea.KeyMsg:
		m.viewImage = ""

		s := msg.String()
		if WebSocketClient == nil {
			// 未连接
			if utils.InArray([]string{"tab", "enter", "up", "down"}, s) {
				// 连接输入框焦点控制
				if s == "up" {
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
						m.tipStr = "连接失败: " + err.Error()
						m.focusIndex = 0
					} else {
						utils.LoginStatus["address"] = m.inputs[0].Value()
						utils.LoginStatus["token"] = m.inputs[1].Value()
						utils.LoginStatus["statue"] = false
						m.tipStr = "已连接"
						// 初始化获取 Bot 信息
						WebSocketClient.SendMessage("get_version_info", nil, "GetVersionInfo")

						pointStatue = "list"
					}
				}
				// 刷新输入框
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
		} else if utils.LoginStatus["statue"] == true {
			// 已连接
			if s == "enter" {
				if pointStatue == "list" || pointStatue == "list-search" {
					selectedItem := m.list.SelectedItem().(userItem)
					// 从 userList 中获取完整信息
					selectedUserInfo := map[string]interface{}{}
					for _, v := range utils.RuntimeData["userList"].([]map[string]interface{}) {
						id := v["user_id"]
						if id == nil {
							id = v["group_id"]
						}
						if id.(float64) == selectedItem.Id() {
							selectedUserInfo = v
							break
						}
					}
					//如果有 group_name 则为群聊
					userType := "private"
					if selectedUserInfo["group_name"] != nil {
						userType = "group"
					}
					// 获取聊天记录（首次）
					if userType == "private" {
						WebSocketClient.SendMessage("get_friend_msg_history", map[string]interface{}{
							"user_id":    selectedItem.Id(),
							"message_id": 0,
							"count":      20,
						}, "GetChatHistoryFist")
					} else {
						WebSocketClient.SendMessage("get_group_msg_history", map[string]interface{}{
							"group_id":   selectedItem.Id(),
							"message_id": 0,
							"count":      20,
						}, "GetChatHistoryFist")
					}

					changeChatView(&m, "chat")
					utils.RuntimeData["chatInfo"] = map[string]interface{}{
						"title": selectedItem.Title(),
						"id":    selectedItem.Id(),
						"type":  userType,
					}
				} else if pointStatue == "chat" {
					if m.sendInput.Focused() {
						// 发送消息
						if utils.RuntimeData["chatInfo"] != nil {
							userType := utils.RuntimeData["chatInfo"].(map[string]interface{})["type"].(string)
							var sendData map[string]interface{}
							if userType == "private" {
								sendData = map[string]interface{}{
									"user_id": utils.RuntimeData["chatInfo"].(map[string]interface{})["id"].(float64),
									"message": m.sendInput.Value(),
								}
							} else {
								sendData = map[string]interface{}{
									"group_id": utils.RuntimeData["chatInfo"].(map[string]interface{})["id"].(float64),
									"message":  m.sendInput.Value(),
								}
							}
							WebSocketClient.SendMessage("send_msg", sendData, "SendMsgBack")
							// 清空输入框
							m.sendInput.SetValue("")
							m.sendInput.Focus()
						}
					} else {
						m.sendInput.Focus()
					}
				}
			}
		}

		if s == "esc" {
			if pointStatue == "chat" {
				m.sendInput.Blur()
			}
		}
		if utils.InArray([]string{"left", "right"}, s) {
			if utils.LoginStatus["statue"] == true {
				if s == "left" && pointStatue == "chat" {
					changeChatView(&m, "list")
				} else if s == "right" && pointStatue == "list" {
					changeChatView(&m, "chat")
				}
			}
		}

		if s == "v" && !m.sendInput.Focused() {
			if pointStatue == "chat" {
				selectedMsg := m.msgViewList.SelectedItem().(msgItem)
				rawMessage := selectedMsg.RawMessage()
				msgTypes := websocket.GetTypesInMessage(rawMessage)
				if utils.InArray(msgTypes, "image") {
					// 找出 rawMessage 中的第一张图片
					for _, v := range rawMessage {
						if v.(map[string]interface{})["type"].(string) == "image" {
							data := v.(map[string]interface{})["data"].(map[string]interface{})
							m.viewImage = data["url"].(string)
						}
					}
				}
			}
		}
	}

	// 始终更新内容，这些由任意交互操作或者外部触发的 500ms 被动刷新进行刷新
	if WebSocketClient == nil {
		cmdList := make([]tea.Cmd, len(m.inputs))
		for i := range m.inputs {
			m.inputs[i], cmdList[i] = m.inputs[i].Update(msg)
		}

		return m, tea.Batch(cmdList...)
	} else if utils.LoginStatus["statue"] == true {
		if utils.RuntimeData["userList"] != nil {
			nowItemSize := len(m.list.Items())
			// 刷新显示的列表
			var items []list.Item
			for _, v := range utils.RuntimeData["userList"].([]map[string]interface{}) {
				if v["user_id"] != nil {
					showName := v["nickname"].(string)
					if v["remark"].(string) != "" {
						showName = v["remark"].(string) + "(" + v["nickname"].(string) + ")"
					}
					longNick := v["longNick"].(string)
					if longNick == "" {
						longNick = "[这个人很懒什么都没写]"
					}
					items = append(items, userItem{
						title:       showName,
						filterTitle: v["py_filter"].(string),
						desc:        longNick,
						id:          v["user_id"].(float64),
					})
				} else {
					items = append(items, userItem{
						title:       v["group_name"].(string),
						filterTitle: v["py_filter"].(string),
						desc:        "[群聊]",
						id:          v["group_id"].(float64),
					})
				}
			}
			// 如果长度变了（todo 或者强制更新），则更新
			// PS：如果每次都更新会导致搜索的时候列表闪烁
			if len(items) != nowItemSize {
				m.list.SetItems(items)
			}

			if pointStatue == "list" || pointStatue == "list-search" {
				if m.list.FilterState() == list.Filtering ||
					m.list.FilterState() == list.FilterApplied {
					pointStatue = "list-search"
				} else {
					pointStatue = "list"
				}
			}
			if m.list.FilterState() == list.FilterApplied {
				m.list.Title = "用户列表（筛选）"
			} else {
				m.list.Title = "用户列表"
			}
		}
		if utils.RuntimeData["messageList"] != nil {
			var items []list.Item
			for _, v := range utils.RuntimeData["messageList"].([]map[string]interface{}) {
				items = append(items, msgItem{
					title:      v["sender"].(map[string]interface{})["nickname"].(string),
					desc:       v["rawMessage"].(string),
					id:         v["messageId"].(float64),
					rawMessage: v["message"].([]interface{}),
				})
			}
			m.msgViewList.SetItems(items)
			m.msgViewList.Title = utils.RuntimeData["chatInfo"].(map[string]interface{})["title"].(string)
		}

		// 检查选中的消息提供附加功能菜单
		if pointStatue == "chat" {
			selectedMsg := m.msgViewList.SelectedItem().(msgItem)
			rawMessage := selectedMsg.RawMessage()
			msgTypes := websocket.GetTypesInMessage(rawMessage)
			if utils.InArray(msgTypes, "image") {
				m.appendControlList["V"] = "查看图片"
			}
		}

		var listCmd tea.Cmd
		if pointStatue == "chat" {
			m.msgViewList, listCmd = m.msgViewList.Update(msg)
		} else {
			m.list, listCmd = m.list.Update(msg)
		}

		var cmdSi tea.Cmd
		m.sendInput, cmdSi = m.sendInput.Update(msg)
		return m, tea.Batch(cmdSi, listCmd)
	}

	return m, nil
}

func (m ChatModel) View() (s string) {
	listGrid := m.flexBox.GetRow(0).GetCell(0)
	mainGrid := m.flexBox.GetRow(0).GetCell(1)

	w, h := listStyle.GetFrameSize()

	// 控制指示器内容
	controlList := map[string]string{}
	switch pointStatue {
	case "login":
		controlList = map[string]string{
			"\U000F060C":    "连接",
			"\uF062|\uF063": "焦点",
		}
		break
	case "list":
		controlList = map[string]string{
			"\uF062|\uF063": "选择",
			"/":             "搜索",
			"\U000F060C":    "选中",
			"\uF061":        "聊天",
		}
		break
	case "list-search":
		controlList = map[string]string{
			"\uF062|\uF063": "选择",
			"\U000F060C":    "选中",
			"ESC":           "取消",
		}
	case "chat":
		controlList = map[string]string{
			"\uF062|\uF063": "滚动",
			"\uF060":        "列表",
		}
		if m.sendInput.Focused() {
			controlList["ESC"] = "取消"
			controlList["\U000F060C"] = "发送"
		} else {
			controlList["\U000F060C"] = "输入"
		}
	}

	// 加上 appendControlList
	if !m.sendInput.Focused() {
		for k, v := range m.appendControlList {
			controlList[k] = v
		}
	}

	if WebSocketClient == nil {
		var b strings.Builder
		for i := range m.inputs {
			b.WriteString(m.inputs[i].View())
			if i < len(m.inputs)-1 {
				b.WriteRune('\n')
			}
		}
		mainGrid.SetStyle(lipgloss.NewStyle().Align(lipgloss.Center).AlignVertical(lipgloss.Center))
		mainGrid.SetContent(titleStyle.Render(" 连接到 OneBot ") + "\n\n" + b.String())
	} else if utils.LoginStatus["statue"] == true {
		// 输入框（占满宽度）
		m.sendInput.Width = mainGrid.GetWidth() - w - 5
		sendInputStyle := lipgloss.NewStyle().Padding(0, 1).Border(lipgloss.NormalBorder(), false, false, true, false).BorderForeground(lipgloss.Color("241")).Foreground(mainFontColor).MarginLeft(3)
		// 更新状态
		mainGrid.SetContent("")
		mainGrid.SetStyle(lipgloss.NewStyle())
		m.tipStr = "已连接"
		// 更新列表
		m.list.SetSize(listGrid.GetWidth()-w, listGrid.GetHeight()-h)
		listGrid.SetContent(listStyle.Render(m.list.View()))
		m.msgViewList.SetSize(mainGrid.GetWidth()-w, mainGrid.GetHeight()-h-2)
		// 绘制聊天区域
		if m.viewImage == "" {
			mainGrid.SetContent(listStyle.Render(m.msgViewList.View()) + "\n " + sendInputStyle.Render(m.sendInput.View()))
		} else {
			mainGrid.SetStyle(lipgloss.NewStyle().Align(lipgloss.Center).AlignVertical(lipgloss.Center))
			// 将 URL 生成二维码
			var buf bytes.Buffer
			// 生成二维码到缓冲区
			qrterminal.GenerateWithConfig(m.viewImage, qrterminal.Config{
				HalfBlocks: true,
				Level:      qrterminal.M,
				Writer:     &buf,

				WhiteChar:      qrterminal.WHITE_WHITE,
				BlackChar:      qrterminal.BLACK_BLACK,
				WhiteBlackChar: qrterminal.WHITE_BLACK,
				BlackWhiteChar: qrterminal.BLACK_WHITE,
			})
			// 如果把 URL 完整输出需要占的行数 viewImage / 宽度
			widthGet := mainGrid.GetWidth() - 2
			urlLine := len(m.viewImage) / widthGet
			// 将 URL 按宽度插入 \n
			linedUrl := ""
			for i := 0; i < urlLine; i++ {
				linedUrl += m.viewImage[i*widthGet:(i+1)*widthGet] + "\n"
			}
			// 从缓冲区获取字符串
			qrString := buf.String()
			line := strings.Split(qrString, "\n")
			if len(line)+urlLine+2 > mainGrid.GetHeight() {
				var buf1 bytes.Buffer
				qrterminal.GenerateWithConfig("显示不下", qrterminal.Config{
					HalfBlocks: true,
					Level:      qrterminal.M,
					Writer:     &buf1,

					WhiteChar:      qrterminal.WHITE_WHITE,
					BlackChar:      qrterminal.BLACK_BLACK,
					WhiteBlackChar: qrterminal.WHITE_BLACK,
					BlackWhiteChar: qrterminal.BLACK_WHITE,
				})
				qrString = buf1.String()
			}
			mainGrid.SetContent(qrString + "\n" + linedUrl + "\n\n" + tipStyle.Render(" • 任意操作退出 • "))
		}
	}

	SetControlBar(m.flexBox, controlList, m.tipStr)
	return m.flexBox.Render()
}

// ======================================================

func changeChatView(m *ChatModel, viewName string) {
	chatViewlist := &m.list
	msgViewList := &m.msgViewList
	switch viewName {
	case "list":
		// 输入框
		m.sendInput.Blur()
		// 样式调整
		chatViewlist.SetDelegate(delegateItemList)
		// 切换案件注册
		setupListKey(chatViewlist)
		cleanListKey(msgViewList)
	case "chat":
		// 样式调整
		chatViewlist.SetDelegate(delegateItemListDark)
		// 取消搜索状态
		if pointStatue == "list-search" {
			m.list.ResetFilter()
		}
		// 切换案件注册
		cleanListKey(chatViewlist)
		setupListKey(msgViewList)
		msgViewList.KeyMap.Filter = key.NewBinding() // 消息列表不提供搜索
	}
	pointStatue = viewName
}

func setupListKey(l *list.Model) {
	l.KeyMap.CursorUp = key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "上"),
	)
	l.KeyMap.CursorDown = key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "下"),
	)
	l.KeyMap.PrevPage = key.NewBinding()
	l.KeyMap.NextPage = key.NewBinding()
	l.KeyMap.GoToStart = key.NewBinding(
		key.WithKeys("home"),
		key.WithHelp("home", "顶部"),
	)
	l.KeyMap.GoToEnd = key.NewBinding(
		key.WithKeys("end"),
		key.WithHelp("end", "底部"),
	)
	l.KeyMap.ShowFullHelp = key.NewBinding()
	l.KeyMap.CloseFullHelp = key.NewBinding()
	l.KeyMap.Quit = key.NewBinding()
}

func cleanListKey(l *list.Model) {
	l.KeyMap.CursorUp = key.NewBinding()
	l.KeyMap.CursorDown = key.NewBinding()
	l.KeyMap.PrevPage = key.NewBinding()
	l.KeyMap.NextPage = key.NewBinding()
	l.KeyMap.GoToStart = key.NewBinding()
	l.KeyMap.GoToEnd = key.NewBinding()
	l.KeyMap.ShowFullHelp = key.NewBinding()
	l.KeyMap.CloseFullHelp = key.NewBinding()
	l.KeyMap.Quit = key.NewBinding()
}

func UserFilter(term string, targets []string) []list.Rank {
	ranks := fuzzy.Find(term, targets)
	sort.Stable(ranks)
	result := make([]list.Rank, len(ranks))
	for i, r := range ranks {
		result[i] = list.Rank{
			Index:          r.Index,
			MatchedIndexes: r.MatchedIndexes,
		}
	}
	return result
}
