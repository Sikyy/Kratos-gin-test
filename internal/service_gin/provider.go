package service_gin

import (
	"Gin-server-text/internal/service"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewGinUseCase)

type GinUseCase struct {
	srv *service.GreeterService

	log *log.Helper
}

func NewGinUseCase(srv *service.GreeterService, logger log.Logger) *GinUseCase {
	return &GinUseCase{
		srv: srv,
		log: log.NewHelper(logger),
	}
}
