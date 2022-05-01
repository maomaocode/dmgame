package pkg

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"testing"
	"time"

	"golang.org/x/net/websocket"
)

func TestWebSocket(t *testing.T) {
	params := map[string]interface{}{
		"uid":      0,
		"roomid":  8722013,
		"platform": "web",
		"protover": 3,
		"type":     2,
		"key":      "KewJcyTmwFO0Is8ED373uc71AmqPRJqrmY-PKUaG8OPCZ_EZ_03UtmNKVx1XMdKYKRxIvT-a5mHI2-1-XJzYR2YqEtE8YltEnq1FldUYvMTi94KPDmMkhlXb9_14xNKL6vFn30A3txPme5UHYxE3vg==",
	}

	loc, _ := url.Parse(fmt.Sprintf("wss://%s:%d/sub", "broadcastlv.chat.bilibili.com", 443))
	ori, _ := url.Parse("https://www.bilibili.com/")

	conn, err := websocket.DialConfig(&websocket.Config{
		Location:  loc,
		Origin:    ori,
		Version: websocket.ProtocolVersionHybi13,
		TlsConfig: &tls.Config{
			InsecureSkipVerify:          true,
		},
		Dialer:    &net.Dialer{
			Timeout:       10*time.Second,
		},
	})

	//conn, err := websocket.Dial(, "", "https://www.bilibili.com/")
	if err != nil {
		t.Fatal(err)
	}

	if err := SendSocketMsg(conn, OpAuth, params); err != nil {
		t.Fatal(err)
	}

	if err := ListenConn(context.Background(), conn, func(code int32, data []byte) error {
		fmt.Println(string(data))
		return nil
	}); err != nil {
		t.Fatal(err)
	}

}
