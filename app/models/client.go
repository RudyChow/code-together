package models

import (
	"log"
	"sync"
	"time"

	"github.com/RudyChow/code-together/app/utils"
	"github.com/RudyChow/code-together/app/ws/messages"
	"github.com/gorilla/websocket"
)

// Client : 客户端
type Client struct {
	conn      *websocket.Conn
	RoomID    string `json:"room_id"`
	Username  string `json:"username"`
	language  string
	version   string
	code      string
	sendCh    chan interface{}
	onceClose *sync.Once
}

const (
	// 写数据超时时间
	writeWait = 10 * time.Second
	// 读取pong信息的时间
	pongWait = 60 * time.Second
	// ping客户端的时间，一定要小于pongWait
	pingPeriod = (pongWait * 9) / 10
	// read最大的size
	maxMessageSize = 512
)

// NewClient : 生成一个客户端
func NewClient(conn *websocket.Conn, roomID string) *Client {
	//如果没有房号，则随便建一个
	if roomID == "" {
		roomID = utils.RandStringBytes(6)
	}

	return &Client{
		conn:      conn,
		RoomID:    roomID,
		Username:  utils.RandStringBytes(8),
		sendCh:    make(chan interface{}, 100),
		onceClose: &sync.Once{},
	}
}

// Read : 从websocket中读取消息
func (c *Client) Read() {
	defer c.close()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("client %s read error: %v", c.Username, err)
			break
		}

		//获取请求消息
		message, err := messages.GetRequest(data)
		if err != nil {
			log.Println(err)
			continue
		}

		switch message.Type {
		//提交代码
		case "push":
			codeRequest, ok := (message.Data).(map[string]interface{})
			if !ok {
				continue
			}

			if _, ok = codeRequest["code"]; !ok {
				continue
			}
			if _, ok = codeRequest["language"]; !ok {
				continue
			}
			if _, ok = codeRequest["version"]; !ok {
				continue
			}

			c.language = codeRequest["language"].(string)
			c.version = codeRequest["version"].(string)
			c.code = codeRequest["code"].(string)

			RoomManager.SyncCh <- c
		//拉取代码
		case "pull":
			c.sendCh <- messages.CodeResponse(RoomManager.rooms[c.RoomID])
		}

	}
}

// Send : 从websocket中发送消息
func (c *Client) Send() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.close()
	}()
	for {
		select {
		// 发送信息
		case message, ok := <-c.sendCh:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteJSON(message)

		// ping客户端
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("client %s send ping err: %v", c.Username, err)
				return
			}
		}
	}
}

// SendInfo : 发送个人信息
func (c *Client) SendInfo() {
	c.sendCh <- messages.InfoResponse(c)
}

// 关闭socket
func (c *Client) close() {
	c.onceClose.Do(func() {
		c.conn.Close()
		RoomManager.LeaveCh <- c
	})
}
