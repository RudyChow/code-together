package ws

import (
	"log"
	"net/http"

	"github.com/RudyChow/code-together/app/models"
	"github.com/RudyChow/code-together/app/ws/messages"
	"github.com/RudyChow/code-together/conf"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// StartWebsocket 开启websocket
func StartWebsocket() {

	gin.SetMode(conf.Cfg.HTTP.Mode)
	r := gin.Default()

	r.GET("/ws/:room", serveWs)
	r.GET("/ws", serveWs)
	r.GET("/api/rooms", func(r *gin.Context) {
		r.JSON(http.StatusOK, models.RoomManager.GetRooms())
	})

	go models.RoomManager.Run()

	r.Run(conf.Cfg.HTTP.Addr)
}

func serveWs(r *gin.Context) {
	conn, err := upgrader.Upgrade(r.Writer, r.Request, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}

	// 为每一个socket创建一个客户端
	client := models.NewClient(conn, r.Param("room"))

	// 加入房间
	err = models.RoomManager.Join(client)
	if err != nil {
		conn.WriteJSON(messages.MessageResponse(err.Error()))
		conn.Close()
		return
	}

	// 成功加入房间
	client.SendInfo()
	go client.Read()
	go client.Send()
}
