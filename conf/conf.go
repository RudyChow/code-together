package conf

import (
	"log"

	"github.com/BurntSushi/toml"
)

// Cfg : 配置
var Cfg *config

type config struct {
	HTTP    *http
	Manager *roomManager
}

type http struct {
	Addr string
	Mode string
}

type roomManager struct {
	RoomCount   int
	RoomExpire  int64
	ClientCount int
	ChargeGap   int
}

func init() {
	//读取配置文件
	if _, err := toml.DecodeFile("./config.toml", &Cfg); err != nil {
		log.Panic(err)
	}
}
