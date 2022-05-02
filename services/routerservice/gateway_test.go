package routerservice

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestGateway_ServeHTTP(t *testing.T) {
	go run()
	uri := "GetNotiAddr"
	client := &http.Client{}
	res, err := client.Post(fmt.Sprintf("http://localhost:30000/%s", uri), "", nil)
	if err != nil {
		t.Fatalf("client res err: %v", err)
	}

	t.Logf("res: %v", res)

	time.Sleep(10 * time.Second)
}

func TestGateway_SendHTTPReq(t *testing.T) {

}