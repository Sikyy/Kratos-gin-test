//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"Gin-server-text/internal/biz"
	"Gin-server-text/internal/conf"
	"Gin-server-text/internal/data"
	"Gin-server-text/internal/server"
	"Gin-server-text/internal/service"
	"Gin-server-text/internal/service_gin"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, service_gin.ProviderSet, newApp))
}
