package models

import "time"

// Room : 房间
type Room struct {
	Clients     map[string]*Client `json:"clients"`
	Language    string             `json:"language"`
	Version     string             `json:"version"`
	Code        string             `json:"code"`
	LastModUser string             `json:"last_mod_user"`
	LastModTime int64              `json:"last_mod_time"`
}

// NewRoom : 新建一个房间
func NewRoom() *Room {
	clients := make(map[string]*Client)
	room := &Room{
		Clients:     clients,
		LastModTime: time.Now().Unix(),
	}

	return room
}
