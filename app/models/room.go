package models

import (
	"fmt"
	"time"

	"github.com/RudyChow/code-together/conf"

	"github.com/RudyChow/code-together/app/ws/messages"
)

// Room : 房间
type Room struct {
	Clients     map[string]*Client `json:"clients"`
	Language    string             `json:"language"`
	Version     string             `json:"version"`
	Code        string             `json:"code"`
	LastModTime int64              `json:"last_mod_time"`
	InCharge    string             `json:"in_charge"`
	chargeTimer *time.Timer
	ID          string
	inChargeCh  chan string
	boardcastCh chan struct{}
}

// NewRoom : 新建一个房间
func NewRoom(ID string) *Room {
	clients := make(map[string]*Client)
	room := &Room{
		Clients:     clients,
		LastModTime: time.Now().Unix(),
		chargeTimer: time.NewTimer(0),
		ID:          ID,
		inChargeCh:  make(chan string, 64),
		boardcastCh: make(chan struct{}, 64),
	}
	room.chargeTimer.Stop()
	go room.Run()
	return room
}

// 同步代码
func (room *Room) syncCode(c *Client) {
	if room.InCharge != "" && room.InCharge != c.Username {
		c.sendCh <- messages.MessageResponse(fmt.Sprintf("%s is coding..", room.InCharge))
		return
	}

	room.Language = c.language
	room.Version = c.version
	room.Code = c.code
	room.LastModTime = time.Now().Unix()
	room.setCharge(c.Username)

	room.chargeTimer.Reset(time.Duration(conf.Cfg.Manager.ChargeGap) * time.Second)
}

func (room *Room) Run() {
	for {
		select {
		// 控制权计时
		case <-room.chargeTimer.C:
			room.setCharge("")
		// 控制权转让
		case <-room.inChargeCh:
			for _, client := range room.Clients {
				client.sendCh <- messages.ChargeResponse(room.InCharge)
			}
		// 广播房间信息
		case <-room.boardcastCh:
			for _, client := range room.Clients {
				client.sendCh <- messages.RoomResponse(room)
			}
		}
	}
}

func (room *Room) setCharge(inCharge string) {
	//有变动则通知
	if room.InCharge != inCharge {
		room.InCharge = inCharge
		room.inChargeCh <- room.InCharge
	}
}
