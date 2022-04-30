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
	OpHeartBeat = 2
	OpAuth      = 7
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

		data = make([]byte, 0, dataLen)
		_, err = conn.Read(data)
		if err != nil {
			return err
		}

		if err := handler(code, data); err != nil {
			fmt.Println(conn.RemoteAddr(), err)
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

	var dataLen int32
	var headLen int16
	var version int16
	var code int32
	var seq int32

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

	return code, dataLen, nil
}
