package view

import (
	"bytes"
	"github.com/Stapxs/Stapxs-QQ-Shell/utils"
	"github.com/Stapxs/Stapxs-QQ-Shell/utils/runtime"
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

var (
	webSocketClient *websocket.Client
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
	flexBox   *flexbox.FlexBox  // 基础布局
	inputs    []textinput.Model // 登录输入框
	sendInput textinput.Model   // 发送消息输入框

	tags Tags // 标签
	data Data // 数据
}

type Tags struct {
	tipStr      string // 提示信息
	focusIndex  int    // 登录输入框焦点位置
	viewImage   string // 查看的图片
	pointStatue string // 当前状态
}

type Data struct {
	list              list.Model        // 好友列表
	msgViewList       list.Model        // 消息列表
	appendControlList map[string]string // 附加控制列表
}

var (
	listStyle            = lipgloss.NewStyle().Margin(1, 2)
	delegateItemList     = list.NewDefaultDelegate()
	delegateItemListBlur = list.NewDefaultDelegate() // 失去焦点
	delegateItemMsg      = list.NewDefaultDelegate()
	delegateItemMsgDark  = list.NewDefaultDelegate() // 消息列表（失去焦点）

	selectTitleStyle        = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(mainColor).Foreground(mainColor).Padding(0, 0, 0, 1)
	selectTitleStyleBlur    = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(mainFontColor).Foreground(mainFontColor).Padding(0, 0, 0, 1)
	selectTitleStyleMsg     = lipgloss.NewStyle().BorderForeground(mainColor).Foreground(mainColor).Padding(0, 0, 0, 2)
	selectTitleStyleMsgBlur = lipgloss.NewStyle().BorderForeground(mainFontColor).Foreground(mainFontColor).Padding(0, 0, 0, 1)
)

// InitialChatModel 初始化聊天视图
func InitialChatModel() *ChatModel {
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
	delegateItemList.Styles.SelectedTitle = selectTitleStyle
	delegateItemList.Styles.SelectedDesc = selectTitleStyle.Foreground(mainColorDark)
	delegateItemListBlur.Styles.SelectedTitle = selectTitleStyleBlur
	delegateItemListBlur.Styles.SelectedDesc = selectTitleStyleBlur.Foreground(mainFontColor)
	userList := list.New([]list.Item{userItem{title: "", desc: ""}}, delegateItemList, 0, 0)
	userList.Title = "用户列表"             // 标题
	userList.FilterInput.Prompt = "搜索：" // 搜索提示
	userList.Filter = userFilter        // 搜索方法
	userList.Styles.Title = titleStyle  // 标题样式
	userList.SetShowStatusBar(false)    // 隐藏状态栏
	userList.SetShowHelp(false)         // 隐藏帮助
	setupListKey(&userList)             // 设置快捷键

	// 消息列表
	delegateItemMsg.Styles.SelectedTitle = selectTitleStyleMsg
	delegateItemMsg.Styles.SelectedDesc = selectTitleStyleMsg.Foreground(mainColorDark)
	delegateItemMsgDark.Styles.SelectedTitle = selectTitleStyleMsgBlur
	delegateItemMsgDark.Styles.SelectedDesc = selectTitleStyleMsgBlur.Foreground(mainFontColor)
	msgList := list.New([]list.Item{msgItem{title: "", desc: ""}}, delegateItemMsg, 0, 0)
	msgList.Title = "消息"              // 标题
	msgList.Styles.Title = titleStyle // 标题样式
	msgList.SetShowStatusBar(false)   // 隐藏状态栏
	msgList.SetShowHelp(false)        // 隐藏帮助
	msgList.SetShowFilter(false)      // 隐藏搜索框
	msgList.SetShowPagination(false)  // 隐藏分页
	cleanListKey(&msgList)            // 清除快捷键

	return &ChatModel{
		flexBox:   flexBox,
		inputs:    inputs,
		sendInput: tMsg,
		tags: Tags{
			tipStr:      "未连接",
			focusIndex:  0,
			viewImage:   "",
			pointStatue: "login",
		},
		data: Data{
			list:              userList,
			msgViewList:       msgList,
			appendControlList: map[string]string{},
		},
	}
}

// 视图 ========================================

func (model *ChatModel) Init() tea.Cmd {
	model.flexBox.SetWidth(WindowWidth)
	model.flexBox.SetHeight(WindowHeight)
	return textinput.Blink
}

func (model *ChatModel) Update(msg tea.Msg) (View, tea.Cmd) {
	// 每次更新清空附加控制列表
	model.data.appendControlList = make(map[string]string)

	switch msg := msg.(type) {
	// 窗口大小变更事件 ========================================
	case tea.WindowSizeMsg:
		model.flexBox.SetWidth(msg.Width)
		model.flexBox.SetHeight(msg.Height)
		return model, nil
	// 按键行为控制事件 ========================================
	case tea.KeyMsg:
		// 任意操作清空图片查看（退出图片查看状态）
		model.tags.viewImage = ""
		inputKey := msg.String()

		// >> 连接页面输入框控制
		if webSocketClient == nil {
			if utils.InArray([]string{"tab", "up", "down"}, inputKey) {
				// 未连接状态，对连接输入框进行操作控制和维护
				if inputKey == "up" {
					model.tags.focusIndex--
				} else {
					model.tags.focusIndex++
				}
				if model.tags.focusIndex > len(model.inputs) {
					model.tags.focusIndex = 0
				} else if model.tags.focusIndex < 0 {
					model.tags.focusIndex = len(model.inputs)
				}
				// 刷新输入框
				cmdList := make([]tea.Cmd, len(model.inputs))
				for i := 0; i <= len(model.inputs)-1; i++ {
					if i == model.tags.focusIndex {
						cmdList[i] = model.inputs[i].Focus()
						continue
					}
					model.inputs[i].Blur()
				}
				// 在连接页面不需要处理其他事件，直接返回
				return model, tea.Batch(cmdList...)
			}
		}

		// >> 回车功能控制
		if inputKey == "enter" {
			if runtime.LoginStatus["statue"] == true {
				// 已连接，对页面上的所有回车操作进行处理
				if model.tags.pointStatue == "list" || model.tags.pointStatue == "list-search" {
					// 列表视图，进入聊天视图
					selectedItem := model.data.list.SelectedItem().(userItem)
					model.loadChat(selectedItem)
					changeChatView(model, "chat")
				} else if model.tags.pointStatue == "chat" {
					// 聊天视图，发送消息
					if model.sendInput.Focused() {
						model.sendMessage()
					} else {
						model.sendInput.Focus()
					}
				}
			} else {
				model.connect()
			}
		}

		// >> 取消输入
		if inputKey == "esc" {
			if model.tags.pointStatue == "chat" {
				model.sendInput.Blur()
			}
		}

		// >> 视图切换
		if utils.InArray([]string{"left", "right"}, inputKey) {
			if runtime.LoginStatus["statue"] == true {
				if inputKey == "left" && model.tags.pointStatue == "chat" {
					changeChatView(model, "list")
				} else if inputKey == "right" && model.tags.pointStatue == "list" {
					changeChatView(model, "chat")
				}
			}
		}

		// >> 图片查看器
		if inputKey == "v" && !model.sendInput.Focused() {
			if model.tags.pointStatue == "chat" {
				selectedMsg := model.data.msgViewList.SelectedItem().(msgItem)
				rawMessage := selectedMsg.RawMessage()
				msgTypes := websocket.GetTypesInMessage(rawMessage)
				if utils.InArray(msgTypes, "image") {
					// 找出 rawMessage 中的第一张图片
					for _, v := range rawMessage {
						if v.(map[string]interface{})["type"].(string) == "image" {
							data := v.(map[string]interface{})["data"].(map[string]interface{})
							model.tags.viewImage = data["url"].(string)
						}
					}
				}
			}
		}

		// >> 刷新连接输入框（需要放在最后处理）
		if webSocketClient == nil {
			cmdList := make([]tea.Cmd, len(model.inputs))
			for i := range model.inputs {
				model.inputs[i], cmdList[i] = model.inputs[i].Update(msg)
			}
			return model, tea.Batch(cmdList...)
		}

		// 刷新列表
		var listCmd tea.Cmd
		if model.tags.pointStatue == "chat" {
			model.data.msgViewList, listCmd = model.data.msgViewList.Update(msg)
		} else {
			model.data.list, listCmd = model.data.list.Update(msg)
		}

		var cmdSi tea.Cmd
		model.sendInput, cmdSi = model.sendInput.Update(msg)
		return model, tea.Batch(cmdSi, listCmd)
	// 周期触发器事件 ========================================
	case utils.UpdateMsg:
		// 刷新 UI 数据
		if runtime.LoginStatus["statue"] == true {
			// >> 好友列表
			if runtime.Data["userList"] != nil {
				// 刷新好友列表
				listSize := len(model.data.list.Items())
				items := model.getFriendList()
				if len(items) != listSize {
					model.data.list.SetItems(items)
				}
				// 刷新搜索
				if model.tags.pointStatue == "list" || model.tags.pointStatue == "list-search" {
					if model.data.list.FilterState() == list.Filtering ||
						model.data.list.FilterState() == list.FilterApplied {
						model.tags.pointStatue = "list-search"
					} else {
						model.tags.pointStatue = "list"
					}
				}
				if model.data.list.FilterState() == list.FilterApplied {
					model.data.list.Title = "用户列表（筛选）"
				} else {
					model.data.list.Title = "用户列表"
				}
			}
			// >> 消息列表
			if runtime.Data["messageList"] != nil {
				var items []list.Item
				for _, v := range runtime.Data["messageList"].([]map[string]interface{}) {
					items = append(items, msgItem{
						title:      v["sender"].(map[string]interface{})["nickname"].(string),
						desc:       v["rawMessage"].(string),
						id:         v["messageId"].(float64),
						rawMessage: v["message"].([]interface{}),
					})
				}
				model.data.msgViewList.SetItems(items)
				model.data.msgViewList.Title =
					runtime.Data["chatInfo"].(map[string]interface{})["title"].(string)
			}

			// >> 刷新功能指示器
			if model.tags.pointStatue == "chat" {
				selectedMsg := model.data.msgViewList.SelectedItem().(msgItem)
				rawMessage := selectedMsg.RawMessage()
				msgTypes := websocket.GetTypesInMessage(rawMessage)
				if utils.InArray(msgTypes, "image") {
					model.data.appendControlList["V"] = "查看图片"
				}
			}
		}
	}

	// 兜底
	return model, nil
}

func (model *ChatModel) View() (s string) {
	sendInputStyle := lipgloss.NewStyle().Padding(0, 1).Border(lipgloss.NormalBorder(), false, false, true, false).BorderForeground(lipgloss.Color("241")).Foreground(mainFontColor).MarginLeft(3)
	centerStyle := lipgloss.NewStyle().Align(lipgloss.Center).AlignVertical(lipgloss.Center)

	listGrid := model.flexBox.GetRow(0).GetCell(0)
	mainGrid := model.flexBox.GetRow(0).GetCell(1)

	w, h := listStyle.GetFrameSize()

	// 刷新控制指示器内容
	controlList := updateListKey(model)
	if !model.sendInput.Focused() {
		for k, v := range model.data.appendControlList {
			controlList[k] = v
		}
	}

	// 绘制视图 ========================================
	if webSocketClient == nil {
		// 绘制登录输入区
		var b strings.Builder
		for i := range model.inputs {
			b.WriteString(model.inputs[i].View())
			if i < len(model.inputs)-1 {
				b.WriteRune('\n')
			}
		}
		mainGrid.SetStyle(centerStyle)
		mainGrid.SetContent(titleStyle.Render(" 连接到 OneBot ") + "\n\n" + b.String())
	} else if runtime.LoginStatus["statue"] == true {
		// 绘制聊天面板
		// >> 状态指示器
		mainGrid.SetContent("")
		mainGrid.SetStyle(lipgloss.NewStyle())
		model.tags.tipStr = "已连接"
		// >> 好友列表
		model.data.list.SetSize(listGrid.GetWidth()-w, listGrid.GetHeight()-h)
		listGrid.SetContent(listStyle.Render(model.data.list.View()))
		model.data.msgViewList.SetSize(mainGrid.GetWidth()-w, mainGrid.GetHeight()-h-2)
		// >> 消息列表
		if model.tags.viewImage == "" {
			// 消息列表
			mainGrid.SetContent(
				listStyle.Render(model.data.msgViewList.View()) + "\n " +
					sendInputStyle.Render(model.sendInput.View()),
			)
		} else {
			// 图片预览器
			mainGrid.SetStyle(centerStyle)
			// 将 URL 生成二维码
			var buf bytes.Buffer
			// 生成二维码到缓冲区
			qrterminal.GenerateWithConfig(model.tags.viewImage, qrterminal.Config{
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
			urlLine := len(model.tags.viewImage) / widthGet
			// 将 URL 按宽度插入 \n
			linedUrl := ""
			for i := 0; i < urlLine; i++ {
				linedUrl += model.tags.viewImage[i*widthGet:(i+1)*widthGet] + "\n"
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
		// >> 消息输入框
		model.sendInput.Width = mainGrid.GetWidth() - w - 5
	}

	SetControlBar(model.flexBox, controlList, model.tags.tipStr)
	return model.flexBox.Render()
}

// ======================================================

func userFilter(term string, targets []string) []list.Rank {
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
