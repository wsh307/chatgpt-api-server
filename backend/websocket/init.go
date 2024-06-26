package websocket

import (
	"context"
	"net/http"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gorilla/websocket"
)

var (
	clientManager = NewClientManager() // 管理者
)
var upGrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func StartWebSocket(ctx context.Context) {
	g.Log().Info(ctx, "启动：WebSocket")
	go clientManager.start()
	go clientManager.ping(ctx)
}

func WsPage(r *ghttp.Request) {
	isAdmin := r.Session.MustGet("isAdmin").Bool()
	if !isAdmin {

		r.Response.Status = 401
		r.Response.WriteJson(g.Map{
			"detail": "Authentication credentials were not provided.",
		})
		return

	}
	conn, err := upGrader.Upgrade(r.Response.ResponseWriter, r.Request, nil)
	if err != nil {
		return
	}
	currentTime := uint64(gtime.Now().Unix())
	client := NewClient(conn.RemoteAddr().String(), conn, currentTime)
	go client.read()
	go client.write()
	// 用户连接事件
	clientManager.Register <- client
}

func init() {
	ctx := gctx.GetInitCtx()
	StartWebSocket(ctx)
	s := g.Server()
	s.BindHandler("/socket", WsPage)
	// go func() {
	// 	// 每秒向所有客户端发送一次 hello
	// 	for {
	// 		clientManager.Broadcast <- &WResponse{
	// 			Event: "hello",
	// 			Data:  gtime.Now().String(),
	// 		}
	// 		time.Sleep(time.Second)
	// 	}
	// }()
}
