package websocket

import (
	"encoding/json"
	"reflect"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/Stapxs/Stapxs-QQ-Shell/utils"
	"github.com/Stapxs/Stapxs-QQ-Shell/utils/runtime"

	"github.com/gorilla/websocket"
)

// Client 定义 WebSocket 客户端结构
type Client struct {
	conn      *websocket.Conn
	url       string
	sendQueue chan string
	mu        sync.Mutex
}

// NewClient 创建新的 WebSocket 客户端
func NewClient(url string) *Client {
	return &Client{
		url:       url,
		sendQueue: make(chan string, 10), // 消息队列
	}
}

// IsConnected 判断是否已经连接
func (c *Client) IsConnected() bool {
	return c.conn != nil
}

// Connect 连接到 WebSocket 服务器
func (c *Client) Connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
	if err != nil {
		return err
	}
	c.conn = conn

	// 启动写和读协程
	go c.readLoop()
	go c.writeLoop()

	return nil
}

// SendRawMessage 添加消息到发送队列
func (c *Client) SendRawMessage(message string) {
	c.sendQueue <- message
}

// SendMessage 发送 OneBot 消息
func (c *Client) SendMessage(name string, value map[string]interface{}, echo string) {
	if value == nil {
		value = make(map[string]interface{})
	}
	data := map[string]interface{}{
		"action": name,
		"params": value,
		"echo":   echo,
	}
	message, _ := json.Marshal(data)
	messageStr := string(message)
	c.SendRawMessage(messageStr)
}

// Close 关闭 WebSocket 连接
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		_ = c.conn.Close()
		close(c.sendQueue)
	}

	c.conn = nil
}

// readLoop 读取消息的后台协程
func (c *Client) readLoop() {
	for {
		_, message, err := c.conn.ReadMessage()
		if err == nil {
			if string(message) != "" {
				msg := string(message)
				parseMessage(c, msg)
			}
		} else if websocket.IsUnexpectedCloseError(err) {
			// 清理一些数据
			runtime.LoginStatus = make(map[string]interface{})
			runtime.Data = make(map[string]interface{})
			// 结束连接
			c.Close()
			return
		}
	}
}

// writeLoop 发送消息的后台协程
func (c *Client) writeLoop() {
	for msg := range c.sendQueue {
		err := c.conn.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			break
		}
	}
}

// ========================================

func parseMessage(c *Client, message string) {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(message), &data)
	if err != nil {
		return
	}

	if data["status"] != nil && data["status"] == "failed" {
		return
	}

	methodName := ""
	var echoList []string
	if _, ok := data["echo"]; ok {
		echoList := strings.Split(data["echo"].(string), "_")
		methodName = echoList[0]
	} else {
		noticeType := data["post_type"].(string)
		if noticeType == "notice" {
			if data["sub_type"] != nil {
				noticeType = data["sub_type"].(string)
			} else {
				noticeType = data["notice_type"].(string)
			}
		}
		// noticeType 是小写下划线的，需要转换为大驼峰
		nameList := strings.Split(noticeType, "_")
		for _, name := range nameList {
			firstChar := name[0]
			methodName += string(firstChar-32) + name[1:]
		}
	}

	v := reflect.ValueOf(MsgFunc{})
	method := v.MethodByName(methodName)
	if method.IsValid() {
		in := make([]reflect.Value, 4)
		in[0] = reflect.ValueOf(c)
		in[1] = reflect.ValueOf(methodName)
		in[2] = reflect.ValueOf(data)
		in[3] = reflect.ValueOf(echoList)
		defer func() string {
			if r := recover(); r != nil {
				runtime.CurrentView = "main"
				runtime.ErrorMsg = "处理消息 " + methodName + " 异常"
				filteredStack := utils.FilterStack(debug.Stack(), "github.com/Stapxs/Stapxs-QQ-Shell")
				runtime.ErrorFullTrace = filteredStack
			}
			return ""
		}()
		method.Call(in)
	}
}
