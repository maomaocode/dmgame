package notiservices

import (
	"testing"
	"time"
)

func TestRoom_NewRoom(t *testing.T) {
	_, err := NewRoom(2208319)
	if err != nil {
		panic(t)
	}

	time.Sleep(10*time.Second)
}