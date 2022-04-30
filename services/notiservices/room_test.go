package notiservices

import (
	"testing"
)

func TestRoom_NewRoom(t *testing.T) {
	_, err := NewRoom(&NotiService{}, 8722013)
	if err != nil {
		panic(t)
	}

	//time.Sleep(10*time.Second)
	select {}
}
