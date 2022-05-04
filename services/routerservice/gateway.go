package routerservice

import (
	"dmgame/util"
	"github.com/go-kratos/kratos/v2/log"
	"net/http"
)

type gateway struct {
	logger *log.Logger
	mux *http.ServeMux
}

func NewGateway() *gateway{
	g := &gateway{
		mux:  http.NewServeMux(),
	}

	httpInterceptor := &util.HttpInterceptor{}
	g.mux.Handle("/GetNotiAddr", httpInterceptor.Handle(g.getNotiAddr))
	return g
}

func (g *gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mux.ServeHTTP(w, r)
}

func run() {
	gateway := NewGateway()
	server := &http.Server{
		Addr:              listenAddr,
		Handler:           gateway,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Errorf("listen and server failed on err: %v", err)
	}
}
