package pkg

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"golang.org/x/net/websocket"
)

const (
	HANDSHAKE         = 0
	HANDSHAKE_REPLY   = 1
	OpHeartBeat       = 2
	OpHeartBeatReply  = 3
	OpSendMsg         = 4
	OpSendMsgReply    = 5
	DISCONNECT_REPLY  = 6
	OpAuth            = 7
	OpAuthReply       = 8
	RAW               = 9
	PROTO_READY       = 10
	PROTO_FINISH      = 11
	CHANGE_ROOM       = 12
	CHANGE_ROOM_REPLY = 13
	REGISTER          = 14
	REGISTER_REPLY    = 15
	UNREGISTER        = 16
	UNREGISTER_REPLY  = 17
	//# B站业务自定义OP
	//# MinBusinessOp = 1000
	//# MaxBusinessOp = 10000
)

func DialSocket(url string) (*websocket.Conn, error) {
	conn, err := websocket.Dial(url, "", "http://localhost")
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func SendSocketMsg(conn *websocket.Conn, code int32, msg interface{}) error {
	fmt.Println("[SEND] ", conn.RemoteAddr(), code, msg)
	_, err := conn.Write(packMsg(code, msg))
	return err
}

func ListenConn(ctx context.Context, conn *websocket.Conn, handler func(code int32, data []byte) error) error {
	defer conn.Close()

	fmt.Println("listen conn: ", conn.RemoteAddr())

	for {

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:

		}

		data := make([]byte, 16)
		_, err := conn.Read(data)
		if err != nil {
			return err
		}
		code, dataLen, err := unPackData(data)
		if err != nil {
			return err
		}

		if code == OpHeartBeatReply {
			dd := make([]byte, 4)
			_, err = conn.Read(dd)
			if err != nil {
				return err
			}
			continue
		}

		data = make([]byte, dataLen)
		_, err = conn.Read(data)
		if err != nil {
			return err
		}

		if err := handler(code, data); err != nil {
			fmt.Println("[ERR] ", conn.RemoteAddr(), err)
		}
	}
}

func packMsg(code int32, msg interface{}) []byte {
	buf := bytes.NewBuffer(nil)

	data, _ := json.Marshal(msg)

	// 4bytes len + 2bytes header len + 2bytes version + 4bytes code + 4bytes seq
	_ = binary.Write(buf, binary.BigEndian, int32(16+len(data)))
	_ = binary.Write(buf, binary.BigEndian, int16(16))
	_ = binary.Write(buf, binary.BigEndian, int16(1))
	_ = binary.Write(buf, binary.BigEndian, code)
	_ = binary.Write(buf, binary.BigEndian, int32(1))

	if len(data) != 0 && msg != nil {
		_ = binary.Write(buf, binary.BigEndian, data)
	}

	return buf.Bytes()
}

func unPackData(data []byte) (int32, int32, error) {
	buf := bytes.NewBuffer(data)

	var (
		dataLen int32
		headLen int16
		version int16
		code    int32
		seq     int32
	)

	if err := binary.Read(buf, binary.BigEndian, &dataLen); err != nil {
		return 0, 0, err
	}
	if err := binary.Read(buf, binary.BigEndian, &headLen); err != nil {
		return 0, 0, err
	}
	if err := binary.Read(buf, binary.BigEndian, &version); err != nil {
		return 0, 0, err
	}
	if err := binary.Read(buf, binary.BigEndian, &code); err != nil {
		return 0, 0, err
	}
	if err := binary.Read(buf, binary.BigEndian, &seq); err != nil {
		return 0, 0, err
	}

	return code, dataLen - int32(headLen), nil
}
