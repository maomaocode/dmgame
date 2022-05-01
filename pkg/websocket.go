package pkg

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/dsnet/compress/brotli"
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
	conn, err := websocket.Dial(url, "", "https://www.bilibili.com/")
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
		msgHeader, err := unPackData(data)
		if err != nil {
			return err
		}

		if msgHeader.code == OpHeartBeatReply {
			dd := make([]byte, 4)
			_, err = conn.Read(dd)
			if err != nil {
				return err
			}
			//var x int32 // 人气
			//binary.Read(bytes.NewBuffer(dd), binary.BigEndian, &x)
			//fmt.Println(x)
			continue
		}

		data = make([]byte, msgHeader.dataLen-int32(msgHeader.headLen))
		_, err = conn.Read(data)
		if err != nil {
			return err
		}

		if msgHeader.code == OpSendMsgReply && msgHeader.version == 3 {
			reader, err := brotli.NewReader(bytes.NewBuffer(data), &brotli.ReaderConfig{})
			if err != nil {
				return err
			}
			data, _ = ioutil.ReadAll(reader)

			for len(data) > 0 {
				msgHeader, err = unPackData(data[:16])
				if err != nil {
					break
				}

				data = data[16:]
				if err := handler(msgHeader.code, data[:msgHeader.dataLen-int32(msgHeader.headLen)]); err != nil {
					fmt.Println("[ERR] ", conn.RemoteAddr(), err)
				}
				data = data[msgHeader.dataLen-int32(msgHeader.headLen):]
			}

			continue
		}

		if err := handler(msgHeader.code, data); err != nil {
			fmt.Println("[ERR] ", conn.RemoteAddr(), err)
		}
	}
}

func packMsg(code int32, msg interface{}) []byte {
	buf := bytes.NewBuffer(nil)

	var data []byte
	if msg != nil {
		data, _ = json.Marshal(msg)
	}

	// 4bytes len + 2bytes header len + 2bytes version + 4bytes code + 4bytes seq
	_ = binary.Write(buf, binary.BigEndian, int32(16+len(data)))
	_ = binary.Write(buf, binary.BigEndian, int16(16))
	_ = binary.Write(buf, binary.BigEndian, int16(1))
	_ = binary.Write(buf, binary.BigEndian, code)
	_ = binary.Write(buf, binary.BigEndian, int32(1))
	_ = binary.Write(buf, binary.BigEndian, data)

	return buf.Bytes()
}

type MsgHeader struct {
	dataLen int32
	headLen int16
	version int16
	code    int32
	seq     int32
}

func unPackData(data []byte) (MsgHeader, error) {
	buf := bytes.NewBuffer(data)

	h := MsgHeader{}

	if err := binary.Read(buf, binary.BigEndian, &h.dataLen); err != nil {
		return h, err
	}
	if err := binary.Read(buf, binary.BigEndian, &h.headLen); err != nil {
		return h, err
	}
	if err := binary.Read(buf, binary.BigEndian, &h.version); err != nil {
		return h, err
	}
	if err := binary.Read(buf, binary.BigEndian, &h.code); err != nil {
		return h, err
	}
	if err := binary.Read(buf, binary.BigEndian, &h.seq); err != nil {
		return h, err
	}

	return h, nil
}
