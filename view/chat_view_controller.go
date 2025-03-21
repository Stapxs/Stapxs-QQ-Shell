package view

import (
	"github.com/Stapxs/Stapxs-QQ-Shell/utils/runtime"
	"github.com/Stapxs/Stapxs-QQ-Shell/utils/websocket"
	"github.com/charmbracelet/bubbles/list"
)

// connect
// @Description: 连接 OneBot 操作
// @receiver model *ChatModel
// @param key string
// @return []tea.Cmd
func (model *ChatModel) connect() {
	// 如果是最后一个输入框，按下 enter 键则连接到 OneBot
	if model.tags.focusIndex == len(model.inputs)-1 {
		if webSocketClient != nil {
			webSocketClient.Close()
		}
		address := model.inputs[0].Value() + "?access_token=" + model.inputs[1].Value()
		webSocketClient = websocket.NewClient(address)
		if err := webSocketClient.Connect(); err != nil {
			webSocketClient = nil
			model.tags.tipStr = "连接失败: " + err.Error()
			model.tags.focusIndex = 0
		} else {
			runtime.LoginStatus["address"] = model.inputs[0].Value()
			runtime.LoginStatus["token"] = model.inputs[1].Value()
			runtime.LoginStatus["statue"] = false
			model.tags.tipStr = "已连接"
			// 初始化获取 Bot 信息
			webSocketClient.SendMessage("get_version_info", nil, "GetVersionInfo")

			model.tags.pointStatue = "list"
		}
	}
}

// loadChat
// @Description: 加载聊天记录
// @receiver model *ChatModel
// @param item userItem
func (model *ChatModel) loadChat(item userItem) {
	// 从 userList 中获取完整信息
	selectedUserInfo := map[string]interface{}{}
	for _, v := range runtime.Data["userList"].([]map[string]interface{}) {
		id := v["user_id"]
		if id == nil {
			id = v["group_id"]
		}
		if id.(float64) == item.Id() {
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
		webSocketClient.SendMessage("get_friend_msg_history", map[string]interface{}{
			"user_id":    item.Id(),
			"message_id": 0,
			"count":      20,
		}, "GetChatHistoryFist")
	} else {
		webSocketClient.SendMessage("get_group_msg_history", map[string]interface{}{
			"group_id":   item.Id(),
			"message_id": 0,
			"count":      20,
		}, "GetChatHistoryFist")
		// 获取群成员列表
		webSocketClient.SendMessage("get_group_member_list", map[string]interface{}{
			"group_id": item.Id(),
			"no_cache": true,
		}, "GetGroupMemberList")
	}

	runtime.Data["chatInfo"] = map[string]interface{}{
		"title": item.Title(),
		"id":    item.Id(),
		"type":  userType,
	}
}

// sendMessage
// @Description: 发送消息
// @receiver model *ChatModel
func (model *ChatModel) sendMessage() {
	if runtime.Data["chatInfo"] != nil {
		userType := runtime.Data["chatInfo"].(map[string]interface{})["type"].(string)
		var sendData map[string]interface{}
		// 构建消息体（map[string]interface{} 数组）
		var message = []map[string]interface{}{
			{
				"type": "text",
				"data": map[string]interface{}{
					"text": model.sendInput.Value(),
				},
			},
		}
		if model.tags.replayMsg != -1 {
			message = append(message, map[string]interface{}{
				"type": "reply",
				"data": map[string]interface{}{
					"id": model.tags.replayMsg,
				},
			})
		}
		// 构建消息
		if userType == "private" {
			sendData = map[string]interface{}{
				"user_id": runtime.Data["chatInfo"].(map[string]interface{})["id"].(float64),
				"message": message,
			}
		} else {
			sendData = map[string]interface{}{
				"group_id": runtime.Data["chatInfo"].(map[string]interface{})["id"].(float64),
				"message":  message,
			}
		}
		webSocketClient.SendMessage("send_msg", sendData, "SendMsgBack")
		// 清空输入框
		model.sendInput.SetValue("")
		model.sendInput.Blur()
		// 清空其他标志
		model.tags.replayMsg = -1
		// 滚动到底部
		model.data.msgViewList.Select(len(model.data.msgViewList.Items()))
	}
}

// getFriendList
// @Description: 整理好友列表
// @receiver model *ChatModel
// @return []list.Item
func (model *ChatModel) getFriendList() []list.Item {
	// 刷新显示的列表
	var items []list.Item
	for _, data := range runtime.Data["userList"].([]map[string]interface{}) {
		if data["user_id"] != nil {
			showName := data["nickname"].(string)
			if data["remark"].(string) != "" {
				showName = data["remark"].(string) + "(" + data["nickname"].(string) + ")"
			}
			var desc = ""
			if data["raw_message"] != nil {
				desc = data["raw_message"].(string)
			}
			items = append(items, userItem{
				title:       showName,
				filterTitle: data["py_filter"].(string),
				desc:        desc,
				id:          data["user_id"].(float64),
			})
		} else {
			var desc = ""
			if data["raw_message"] != nil {
				desc = data["raw_message"].(string)
			}
			items = append(items, userItem{
				title:       data["group_name"].(string),
				filterTitle: data["py_filter"].(string),
				desc:        desc,
				id:          data["group_id"].(float64),
			})
		}
	}
	return items
}
