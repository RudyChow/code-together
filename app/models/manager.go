package models

import (
	"errors"
	"log"
	"time"

	"github.com/RudyChow/code-together/conf"
)

var RoomManager *roomManager

type roomManager struct {
	rooms   map[string]*Room
	SyncCh  chan *Client
	LeaveCh chan *Client
}

func init() {
	RoomManager = &roomManager{
		rooms:   make(map[string]*Room),
		SyncCh:  make(chan *Client, 128),
		LeaveCh: make(chan *Client, 128),
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
		// gc
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
		manager.rooms[c.RoomID] = NewRoom(c.RoomID)
	}

	// 房间人数判断
	if len(manager.rooms[c.RoomID].Clients) >= conf.Cfg.Manager.ClientCount {
		return errors.New("满人了惹")
	}
	manager.rooms[c.RoomID].Clients[c.Username] = c
	return nil
}

func (manager *roomManager) GetRooms() map[string]*Room {
	return manager.rooms
}

// 同步代码
func (manager *roomManager) sync(c *Client) {
	manager.rooms[c.RoomID].syncCode(c)
	manager.rooms[c.RoomID].boardcastCh <- struct{}{}
}

// 客户端离开
func (manager *roomManager) leave(c *Client) {
	log.Printf("client %v is leaving", c.Username)
	if _, ok := manager.rooms[c.RoomID]; ok {
		//如果房间存在，则删除房间内该用户
		delete(manager.rooms[c.RoomID].Clients, c.Username)
		//如果房间没人了，则删除房间
		if len(manager.rooms[c.RoomID].Clients) == 0 {
			delete(manager.rooms, c.RoomID)
		} else {
			manager.rooms[c.RoomID].boardcastCh <- struct{}{}
		}
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
