package routerservice

import "time"

const (
	listenAddr = "127.0.0.1:30000"
	readTimeout = 300 * time.Second
	writeTimeout = 300 * time.Second
	idleTimeout = 600 * time.Second
)
