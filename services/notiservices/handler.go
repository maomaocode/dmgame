package notiservices

import (
	"fmt"
)

func DanMuHandler(service *NotiService, data []byte) error {
	//jsonData,err := simplejson.NewJson(data)
	//if err != nil {
	//	return err
	//}

	fmt.Println("[DanMu]",string(data))
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
