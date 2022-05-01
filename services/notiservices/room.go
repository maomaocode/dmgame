package notiservices

import (
	"context"
	"fmt"
	"sync"
	"time"

	"dmgame/pkg"

	"github.com/bitly/go-simplejson"
	"golang.org/x/net/websocket"
)

const (
	RoomInitURL  = "https://api.live.bilibili.com/xlive/web-room/v1/index/getInfoByRoom?room_id=%d"
	DanMuInitURL = "https://api.live.bilibili.com/xlive/web-room/v1/index/getDanmuInfo?id=%d"
	WebSocketURL = "wss://%s:%d/sub"
)

type Room struct {
	service *NotiService

	Uid        int
	RoomID     int
	LivaStatus int

	token     string
	danMuHost []*DanMuHost

	conn *websocket.Conn
	mux  sync.Mutex

	cmdHandlerMapping map[string]func(service *NotiService, data []byte) error
}

type DanMuHost struct {
	Host    string
	Port    int
	WssPort int
	WsPort  int
}

func NewRoom(service *NotiService, roomID int32) (*Room, error) {
	room := new(Room)
	room.service = service

	if err := room.initRoomInfo(roomID); err != nil {
		return nil, err
	}

	if err := room.initDanMuInfo(); err != nil {
		return nil, err
	}

	if err := room.initConnection(); err != nil {
		return nil, err
	}

	if err := room.initHandlers(); err != nil {
		return nil, err
	}

	return room, nil
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
		//url := fmt.Sprintf(WebSocketURL, host.Host, host.WssPort)
		url := fmt.Sprintf(WebSocketURL, "broadcastlv.chat.bilibili.com", host.WssPort)
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
		"roomid":  r.RoomID,
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

func (r *Room) handler(code int32, data []byte) error {
	//fmt.Println("[REV] code=", code, " data=", string(data))

	switch code {
	case pkg.OpAuthReply:
		_ = r.heartBeat()
	case pkg.OpSendMsgReply:
		jsonData, err := simplejson.NewJson(data)
		if err != nil {
			return err
		}

		cmd, err := jsonData.Get("cmd").String()
		if err != nil {
			return err
		}

		if handler, ok := r.cmdHandlerMapping[cmd]; ok {
			if err := handler(r.service, data); err != nil {
				return err
			}
		}

	}

	return nil
}

func (r *Room) initHandlers() error {
	if r.cmdHandlerMapping == nil {
		r.cmdHandlerMapping = make(map[string]func(service *NotiService, data []byte) error)
	}

	// *****************************************************************
	// cmd handler

	//收到弹幕
	r.cmdHandlerMapping["DANMU_MSG"] = DanMuHandler
	//有人送礼
	r.cmdHandlerMapping["SEND_GIFT"] = GiftHandler
	//有人上舰
	r.cmdHandlerMapping["GUARD_BUY"] = BuyHandler
	//醒目留言
	r.cmdHandlerMapping["SUPER_CHAT_MESSAGE"] = ChatMsgHandler
	//删除醒目留言
	r.cmdHandlerMapping["SUPER_CHAT_MESSAGE_DELETE"] = ChadMsgDelHandler
	// *****************************************************************

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		if err := pkg.ListenConn(ctx, r.conn, r.handler); err != nil {
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
				_ = r.heartBeat()
			}
		}
	}()

	return nil
}
