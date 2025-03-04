package websocket

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Stapxs/Stapxs-QQ-Shell/utils"
	"github.com/Stapxs/Stapxs-QQ-Shell/utils/runtime"
	"github.com/mozillazg/go-pinyin"
)

type MsgFunc struct{}

// GetVersionInfo 获取版本信息
func (m MsgFunc) GetVersionInfo(c *Client, head string, msg map[string]interface{}, echoList []string) {
	// 如果 runtime 存在（即不是第一次连接），且 app_name 不同，重置 runtime
	if len(runtime.Data) != 0 && runtime.Data["botInfo"] != nil {
		nowBotName := runtime.Data["botInfo"].(map[string]interface{})["app_name"].(string)
		if nowBotName != msg["data"].(map[string]interface{})["app_name"].(string) {
			runtime.Data = make(map[string]interface{})
		}
	}
	runtime.Data["botInfo"] = msg["data"]
	runtime.LoginStatus["statue"] = true
	c.SendMessage("get_login_info", nil, "GetLoginInfo")
}

// GetLoginInfo 获取登录信息
func (m MsgFunc) GetLoginInfo(c *Client, head string, msg map[string]interface{}, echoList []string) {
	runtime.Data["loginInfo"] = msg["data"]

	userId := msg["data"].(map[string]interface{})["user_id"].(float64)
	// 如果 runtime 存在（即不是第一次连接），且 userId 不同，重置 runtime
	if len(runtime.Data) != 0 && runtime.Data["loginInfo"] != nil {
		if userId != runtime.Data["loginInfo"].(map[string]interface{})["user_id"].(float64) {
			runtime.Data = make(map[string]interface{})
		}
	}

	reloadUser(c)
}

// GetFriendList 获取好友列表
func (m MsgFunc) GetFriendList(c *Client, head string, msg map[string]interface{}, echoList []string) {
	data := msg["data"].([]interface{})
	backList := make([]map[string]interface{}, 0)
	for _, buddy := range data {
		classId := buddy.(map[string]interface{})["categoryId"].(float64)
		// 9999 是特别关心，会存在于其他分组中；直接跳过
		if classId != 9999 {
			userList := buddy.(map[string]interface{})["buddyList"].([]interface{})
			for _, userInfo := range userList {
				user := make(map[string]interface{})
				user["longNick"] = userInfo.(map[string]interface{})["longNick"]
				user["nickname"] = userInfo.(map[string]interface{})["nickname"]
				user["remark"] = userInfo.(map[string]interface{})["remark"]
				user["user_id"] = userInfo.(map[string]interface{})["user_id"]
				p := pinyin.NewArgs()
				pyNameList := pinyin.Pinyin(user["nickname"].(string)+user["remark"].(string), p)
				pyName := ""
				for _, v := range pyNameList {
					pyName += v[0]
				}
				if pyName == "" {
					user["py_start"] = user["nickname"].(string)[:1]
					user["py_filter"] = user["nickname"].(string)
				} else {
					user["py_start"] = fmt.Sprintf("%c", pyName[0]-32)
					user["py_filter"] = user["nickname"].(string) + user["remark"].(string) + pyName
				}
				backList = append(backList, user)
			}
		}
	}
	// 如果 runtime.Data["userList"] 不为空则追加
	if runtime.Data["userList"] != nil {
		runtime.Data["userList"] = append(runtime.Data["userList"].([]map[string]interface{}), backList...)
	} else {
		runtime.Data["userList"] = backList
	}
	// 将 userList 根据 py_start 排序
	runtime.Data["userList"] = sortUserList(runtime.Data["userList"].([]map[string]interface{}))
}

// GetGroupList 获取群列表
func (m MsgFunc) GetGroupList(c *Client, head string, msg map[string]interface{}, echoList []string) {
	data := msg["data"].([]interface{})
	backList := make([]map[string]interface{}, 0)
	for _, group := range data {
		groupInfo := group.(map[string]interface{})
		user := make(map[string]interface{})
		user["group_id"] = groupInfo["group_id"]
		user["group_name"] = groupInfo["group_name"]
		user["member_count"] = groupInfo["member_count"]
		p := pinyin.NewArgs()
		pyNameList := pinyin.Pinyin(user["group_name"].(string), p)
		pyName := ""
		for _, v := range pyNameList {
			pyName += v[0]
		}
		if pyName == "" {
			user["py_start"] = user["group_name"].(string)[:1]
			user["py_filter"] = user["group_name"].(string)
		} else {
			user["py_start"] = fmt.Sprintf("%c", pyName[0]-32)
			user["py_filter"] = user["group_name"].(string) + pyName
		}
		backList = append(backList, user)
	}
	// 如果 runtime.Data["userList"] 不为空则追加
	// PS：它们共用一个列表
	if runtime.Data["userList"] != nil {
		runtime.Data["userList"] = append(runtime.Data["userList"].([]map[string]interface{}), backList...)
	} else {
		runtime.Data["userList"] = backList
	}
	// 将 userList 根据 py_start 排序
	runtime.Data["userList"] = sortUserList(runtime.Data["userList"].([]map[string]interface{}))
}

// GetGroupMemberList 获取群成员列表
func (m MsgFunc) GetGroupMemberList(c *Client, head string, msg map[string]interface{}, echoList []string) {
	data := msg["data"].([]interface{})
	backList := make([]map[string]interface{}, 0)
	for _, member := range data {
		memberInfo := member.(map[string]interface{})
		user := make(map[string]interface{})
		user["user_id"] = memberInfo["user_id"]
		user["nickname"] = memberInfo["nickname"]
		user["card"] = memberInfo["card"]
		user["role"] = memberInfo["role"]
		backList = append(backList, user)
	}
	runtime.Data["chatInfo"].(map[string]interface{})["memberList"] = backList
}

// GetChatHistoryFist 获取聊天记录（首次）
func (m MsgFunc) GetChatHistoryFist(c *Client, head string, msg map[string]interface{}, echoList []string) {
	if msg["data"] != nil {
		data := msg["data"].(map[string]interface{})
		messages := data["messages"].([]interface{})
		//  从 messages 中移除 raw_message 为空的消息
		singleMsgList := make([]map[string]interface{}, 0, len(messages))
		for i := 0; i < len(messages); i++ {
			if messages[i].(map[string]interface{})["raw_message"] != "" {
				singleMsg := parseMessageBody(messages[i].(map[string]interface{}))
				singleMsgList = append(singleMsgList, singleMsg)
			}
		}
		runtime.Data["messageList"] = singleMsgList
	}
}

// Notice 方法 ========================================

// MessageSent Notice 收到自己发送的消息
func (m MsgFunc) MessageSent(c *Client, head string, msg map[string]interface{}, echoList []string) {
	m.Message(c, head, msg, echoList)
}

// Message Notice 收到的消息
func (m MsgFunc) Message(c *Client, head string, msg map[string]interface{}, echoList []string) {
	var senderId float64
	if msg["group_id"] != nil {
		senderId = msg["group_id"].(float64)
	} else {
		senderId = msg["target_id"].(float64)
	}

	if runtime.Data["chatInfo"] != nil && runtime.Data["chatInfo"].(map[string]interface{})["id"].(float64) == senderId {
		singleMsg := parseMessageBody(msg)
		runtime.Data["messageList"] = append(runtime.Data["messageList"].([]map[string]interface{}), singleMsg)
	}

	// 刷新用户列表中的时间字段
	if runtime.Data["userList"] != nil {
		userList := runtime.Data["userList"].([]map[string]interface{})
		for i, user := range userList {
			var itemId float64
			if user["user_id"] != nil {
				itemId = user["user_id"].(float64)
			} else {
				itemId = user["group_id"].(float64)
			}
			if itemId == senderId {
				userList[i]["time"] = msg["time"]
				userList[i]["raw_message"] = getRawMessage(msg["message"].([]interface{}))
				break
			}
		}
		// 对整个列表按 time 排序
		for i := 0; i < len(userList); i++ {
			for j := i + 1; j < len(userList); j++ {
				timeI := float64(0)
				timeJ := float64(0)
				if userList[i]["time"] != nil {
					timeI = userList[i]["time"].(float64)
				}
				if userList[j]["time"] != nil {
					timeJ = userList[j]["time"].(float64)
				}
				if timeI < timeJ {
					userList[i], userList[j] = userList[j], userList[i]
				}
			}
		}

		runtime.Data["userList"] = userList
		runtime.UpdateList = true
	}
}

// 私有方法 ========================================

func reloadUser(c *Client) {
	c.SendMessage("get_friends_with_category", nil, "GetFriendList")
	c.SendMessage("get_group_list", nil, "GetGroupList")
}

func sortUserList(userList []map[string]interface{}) []map[string]interface{} {
	// 按 py_start 排序
	for i := 0; i < len(userList); i++ {
		for j := i + 1; j < len(userList); j++ {
			if userList[i]["py_start"].(string) > userList[j]["py_start"].(string) {
				userList[i], userList[j] = userList[j], userList[i]
			}
		}
	}
	return userList
}

func parseMessageBody(message map[string]interface{}) map[string]interface{} {
	singleMsg := make(map[string]interface{})
	singleMsg["messageId"] = message["message_id"] // 消息 ID
	singleMsg["sender"] = message["sender"]        // 发送者
	singleMsg["receiver"] = message["real_id"]     // 接收者
	singleMsg["time"] = message["time"]            // 时间
	singleMsg["message"] = message["message"]      // 消息内容
	singleMsg["rawMessage"] = getRawMessage(message["message"].([]interface{}))
	return singleMsg
}

func getRawMessage(messageItem []interface{}) string {
	finalStr := ""
	for _, item := range messageItem {
		msgType := item.(map[string]interface{})["type"].(string)
		data := item.(map[string]interface{})["data"].(map[string]interface{})
		switch msgType {
		case "text":
			text := data["text"].(string)
			text = strings.ReplaceAll(text, "\n", "")
			text = strings.ReplaceAll(text, "\r", "")
			finalStr += text
		case "at":
			// at 消息只针对打开着的群组有效
			var chatInfo = runtime.Data["chatInfo"].(map[string]interface{})
			var memberList = chatInfo["memberList"].([]map[string]interface{})
			var get = false
			for _, member := range memberList {
				// data["qq"].(string)
				var qq, _ = strconv.ParseFloat(data["qq"].(string), 64)
				if member["user_id"] == qq {
					get = true
					if member["card"] != "" {
						finalStr += "@" + member["card"].(string)
					} else {
						finalStr += "@" + member["nickname"].(string)
					}
				}
			}
			if !get {
				finalStr += "[@]"
			}
		case "face":
			finalStr += "[表情]"
		case "bface":
			finalStr += data["text"].(string)
		case "image":
			if data["summary"] != nil && data["summary"] != "" {
				finalStr += data["summary"].(string)
			} else {
				finalStr += "[图片]"
			}
		case "record":
			finalStr += "[语音]"
		case "video":
			finalStr += "[视频]"
		case "file":
			finalStr += "[文件]"
		case "json":
			finalStr += "[卡片消息]"
		case "xml":
			finalStr += "[卡片消息]"
		}
	}
	return finalStr
}

// GetTypesInMessage 获取消息中的类型
func GetTypesInMessage(messageItem []interface{}) []string {
	types := make([]string, 0)
	for _, item := range messageItem {
		msgType := item.(map[string]interface{})["type"].(string)
		if !utils.InArray(types, msgType) {
			types = append(types, msgType)
		}
	}
	return types
}
