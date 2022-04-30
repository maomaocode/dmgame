// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	biz2 "dmgame/services/notiservices/internal/biz"
	conf2 "dmgame/services/notiservices/internal/conf"
	data2 "dmgame/services/notiservices/internal/data"
	server2 "dmgame/services/notiservices/internal/server"
	service2 "dmgame/services/notiservices/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf2.Server, *conf2.Data, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server2.ProviderSet, data2.ProviderSet, biz2.ProviderSet, service2.ProviderSet, newApp))
}
