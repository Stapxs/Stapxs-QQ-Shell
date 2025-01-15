package websocket

import (
	"log"
	"sync"

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

// SendMessage 添加消息到发送队列
func (c *Client) SendMessage(message string) {
	c.sendQueue <- message
}

// Close 关闭 WebSocket 连接
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		_ = c.conn.Close()
		close(c.sendQueue)
	}
}

// readLoop 读取消息的后台协程
func (c *Client) readLoop() {
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("读取消息失败: %v\n", err)
			break
		}
		log.Printf("收到消息: %s\n", string(message))
	}
}

// writeLoop 发送消息的后台协程
func (c *Client) writeLoop() {
	for msg := range c.sendQueue {
		err := c.conn.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			log.Printf("发送消息失败: %v\n", err)
			break
		}
	}
}
