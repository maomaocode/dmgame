package notiservices

import (
	"fmt"

	"github.com/bitly/go-simplejson"
)

func DanMuHandler(service *NotiService, data []byte) error {
	jsonData, err := simplejson.NewJson(data)
	if err != nil {
		return err
	}
	info := jsonData.Get("info")

	var (
		msg, _  = info.GetIndex(1).String()
		//id, _   = info.GetIndex(2).GetIndex(0).Int()
		name, _ = info.GetIndex(2).GetIndex(1).String()
		lv, _   = info.GetIndex(4).GetIndex(0).Int()
	)
	fmt.Printf("[%s] lv%-3d: %s\n", name, lv, msg)

	return nil
}

func GiftHandler(service *NotiService, data []byte) error {
	return nil
}

func BuyHandler(service *NotiService, data []byte) error {
	return nil
}

func ChatMsgHandler(service *NotiService, data []byte) error {
	return nil
}

func ChadMsgDelHandler(service *NotiService, data []byte) error {
	return nil
}
