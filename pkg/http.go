package pkg

import (
	"fmt"
	"io/ioutil"
	"net/http"

	json "github.com/bitly/go-simplejson"
)

var (
	header = map[string]string{
		"Accept":       "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Content-Type": "application/json; charset=utf-8",
		"User-Agent":   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.51 Safari/537.36",
	}
)

func HTTPRequest(method, url string) (*json.Json, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range header {
		req.Header.Set(k, v)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.NewJson(data)
	if err != nil {
		return nil, err
	}

	code, _ := jsonData.Get("code").Int()
	if code != 0 {
		return nil, fmt.Errorf("request err, msg:%s", jsonData.Get("message"))
	}

	return jsonData.Get("data"), nil
}
