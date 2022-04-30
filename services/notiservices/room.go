package notiservices

import (
	"context"
	"fmt"
	"sync"
	"time"

	"dmgame/pkg"

	"golang.org/x/net/websocket"
)

const (
	RoomInitURL  = "https://api.live.bilibili.com/xlive/web-room/v1/index/getInfoByRoom?room_id=%d"
	DanMuInitURL = "https://api.live.bilibili.com/xlive/web-room/v1/index/getDanmuInfo?id=%d"
	WebSocketURL = "wss://%s:%d/sub"
)



type Room struct {
	Uid        int
	RoomID     int
	LivaStatus int

	token     string
	danMuHost []*DanMuHost

	conn *websocket.Conn
	mux  sync.Mutex
}

type DanMuHost struct {
	Host    string
	Port    int
	WssPort int
	WsPort  int
}

func (r *Room) initRoomInfo(roomID int32) error {
	url := fmt.Sprintf(RoomInitURL, roomID)
	jsonData, err := pkg.HTTPRequest("GET", url)
	if err != nil {
		return err
	}

	roomInfo := jsonData.Get("room_info")

	r.Uid, _ = roomInfo.Get("uid").Int()
	r.RoomID, _ = roomInfo.Get("room_id").Int()
	r.LivaStatus, _ = roomInfo.Get("live_status").Int()

	if r.LivaStatus == 0 {
		return fmt.Errorf("直播间未开播")
	}

	return nil
}

func (r *Room) initDanMuInfo() error {
	url := fmt.Sprintf(DanMuInitURL, r.Uid)
	jsonData, err := pkg.HTTPRequest("GET", url)
	if err != nil {
		return err
	}

	r.token, _ = jsonData.Get("token").String()

	hostList, _ := jsonData.Get("host_list").Array()
	for i := range hostList {
		host := jsonData.Get("host_list").GetIndex(i)

		danMuHost := &DanMuHost{}
		danMuHost.Host, _ = host.Get("host").String()
		danMuHost.Port, _ = host.Get("port").Int()
		danMuHost.WssPort, _ = host.Get("wss_port").Int()
		danMuHost.WsPort, _ = host.Get("ws_port").Int()

		r.danMuHost = append(r.danMuHost, danMuHost)
	}

	return nil
}

func (r *Room) initConnection() error {
	for _, host := range r.danMuHost {
		url := fmt.Sprintf(WebSocketURL, host.Host, host.WssPort)
		conn, err := pkg.DialSocket(url)
		if err != nil {
			return err
		}
		r.mux.Lock()
		r.conn = conn
		r.mux.Unlock()

		if err := r.auth(); err != nil {
			return err
		}

		return nil
	}

	return fmt.Errorf("room=%d 没有可用的socket 链接", r.RoomID)
}

func (r *Room) auth() error {
	params := map[string]interface{}{
		"uid":      0,
		"room_id":  r.RoomID,
		"platform": "web",
		"protover": 3,
		"type":     2,
		"key":      r.token,
	}

	return pkg.SendSocketMsg(r.conn, pkg.OpAuth, params)
}

func (r *Room) heartBeat() error {
	return pkg.SendSocketMsg(r.conn, pkg.OpHeartBeat, nil)
}

func NewRoom(roomID int32) (*Room, error) {
	room := new(Room)
	if err := room.initRoomInfo(roomID); err != nil {
		return nil, err
	}

	if err := room.initDanMuInfo(); err != nil {
		return nil, err
	}

	if err := room.initConnection(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		if err := pkg.ListenConn(ctx, room.conn, func(code int32, data []byte) error {
			fmt.Println("recv: ", code, string(data))
			return nil
		}); err != nil {
			fmt.Printf("%v while listening conn, closed\n", err)
		}
	}()

	go func() {
		defer func() {
			cancel()
			fmt.Println("ticker exit")
		}()

		fmt.Println("ticker start")
		ticker := time.Tick(30 * time.Second)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker:
				_ = room.heartBeat()
			}
		}
	}()
	return room, nil
}
