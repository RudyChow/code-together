package models

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/RudyChow/code-together/app/ws/messages"
	"github.com/RudyChow/code-together/conf"
)

var RoomManager *roomManager

type roomManager struct {
	rooms           map[string]*Room
	SyncCh          chan *Client
	LeaveCh         chan *Client
	RoomBroadcastCh chan string
}

func init() {
	RoomManager = &roomManager{
		rooms:           make(map[string]*Room),
		SyncCh:          make(chan *Client, 100),
		LeaveCh:         make(chan *Client, 100),
		RoomBroadcastCh: make(chan string, 100),
	}
}

// Run 开始运行房间管理者
func (manager *roomManager) Run() {
	ticker := time.NewTicker(time.Second)

	for {
		select {
		// 同步代码
		case client := <-manager.SyncCh:
			manager.sync(client)
		// 离开房间
		case client := <-manager.LeaveCh:
			manager.leave(client)
		// 广播房间信息
		case roomID := <-manager.RoomBroadcastCh:
			manager.roomBroadcast(roomID)
		case <-ticker.C:
			manager.gc()
		}
	}
}

// Join : 新加一个客户端
func (manager *roomManager) Join(c *Client) error {
	// 房间判断
	_, ok := manager.rooms[c.RoomID]
	if len(manager.rooms) >= conf.Cfg.Manager.RoomCount && !ok {
		return errors.New("没有空闲房间了惹")
	}
	if !ok {
		manager.rooms[c.RoomID] = NewRoom()
		manager.rooms[c.RoomID].LastModUser = c.Username
	}

	// 房间人数判断
	if len(manager.rooms[c.RoomID].Clients) >= conf.Cfg.Manager.ClientCount {
		return errors.New("满人了惹")
	}
	manager.rooms[c.RoomID].Clients[c.Username] = c

	// 广播	房间信息
	manager.RoomBroadcastCh <- c.RoomID
	return nil
}

// 同步代码
func (manager *roomManager) sync(c *Client) {
	now := time.Now().Unix()
	// 如果最后修改人不是自己 并且 具体最后修改时间x秒内 则不可进行修改
	if manager.rooms[c.RoomID].LastModUser != c.Username && now-manager.rooms[c.RoomID].LastModTime <= conf.Cfg.Manager.SyncGap {
		c.SendData(messages.MessageResponse(fmt.Sprintf("%s is coding..", manager.rooms[c.RoomID].LastModUser)))
		return
	}
	manager.rooms[c.RoomID].Language = c.language
	manager.rooms[c.RoomID].Version = c.version
	manager.rooms[c.RoomID].Code = c.code
	manager.rooms[c.RoomID].LastModUser = c.Username
	manager.rooms[c.RoomID].LastModTime = now
	manager.RoomBroadcastCh <- c.RoomID
}

// 客户端离开
func (manager *roomManager) leave(c *Client) {
	log.Printf("client %v is leaving", c.Username)
	_, ok := manager.rooms[c.RoomID]
	//如果房间存在，则删除房间内该用户
	if ok {
		delete(manager.rooms[c.RoomID].Clients, c.Username)
		//如果房间没人了，则删除房间
		if len(manager.rooms[c.RoomID].Clients) == 0 {
			delete(manager.rooms, c.RoomID)
		}
	}
	manager.RoomBroadcastCh <- c.RoomID
}

// 广播房间信息
func (manager *roomManager) roomBroadcast(roomID string) {
	if _, ok := manager.rooms[roomID]; !ok {
		return
	}
	for _, client := range manager.rooms[roomID].Clients {
		client.sendCh <- messages.RoomResponse(manager.rooms[roomID])
	}
}

// 定期清理过期房间
func (manager *roomManager) gc() {
	for roomID, room := range manager.rooms {
		// 如果没过期，则不处理
		if time.Now().Unix()-room.LastModTime <= conf.Cfg.Manager.RoomExpire {
			continue
		}

		for _, client := range room.Clients {
			client.conn.Close()
			delete(room.Clients, client.Username)
		}

		delete(manager.rooms, roomID)
	}
}

func (manager *roomManager) GetRooms() map[string]*Room {
	return manager.rooms
}
