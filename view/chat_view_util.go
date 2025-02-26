package view

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
)

// changeChatView
// @Description: 切换聊天视图
// @receiver model *ChatModel
// @param viewName string
func changeChatView(model *ChatModel, viewName string) {
	chatViewlist := &model.data.list
	msgViewList := &model.data.msgViewList
	switch viewName {
	case "list":
		// 输入框
		model.sendInput.Blur()
		// 样式调整
		chatViewlist.SetDelegate(delegateItemList)
		msgViewList.SetDelegate(delegateItemListBlur)
		// 切换案件注册
		setupListKey(chatViewlist)
		cleanListKey(msgViewList)
	case "chat":
		// 样式调整
		chatViewlist.SetDelegate(delegateItemListBlur)
		msgViewList.SetDelegate(delegateItemList)
		// 取消搜索状态
		if model.tags.pointStatue == "list-search" {
			model.data.list.ResetFilter()
		}
		// 切换案件注册
		cleanListKey(chatViewlist)
		setupListKey(msgViewList)
		msgViewList.KeyMap.Filter = key.NewBinding() // 消息列表不提供搜索
	}
	model.tags.pointStatue = viewName
}

func updateListKey(model *ChatModel) map[string]string {
	controlList := map[string]string{}
	switch model.tags.pointStatue {
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
		if model.sendInput.Focused() {
			controlList["ESC"] = "取消"
			controlList["\U000F060C"] = "发送"
		} else {
			controlList["\U000F060C"] = "输入"
		}
	}
	return controlList
}

// setupListKey
// @Description: 设置列表按键
// @param l *list.Model
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

// cleanListKey
// @Description: 清除列表按键
// @param l *list.Model
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
