package util

import (
	"github.com/go-kratos/kratos/v2/log"
	"net/http"
)

type HttpCb func(w http.ResponseWriter, r *http.Request) error

type HttpInterceptor struct {
}

func (i *HttpInterceptor) Handle (cb HttpCb) http.Handler{
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			route := r.RequestURI
			err := cb(w, r)
			if err != nil {
				log.Errorf("uri: %s err: %v\n", route, err)
			}
		},
		)
}