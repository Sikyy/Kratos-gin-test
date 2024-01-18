package data

import (
	"Gin-server-text/internal/conf"
	"Gin-server-text/pkg"
	"time"

	"github.com/duke-git/lancet/convertor"
	"github.com/go-redis/redis/extra/redisotel/v8"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo, NewGinRepo, NewMysql, NewRedis)

// Data .
type Data struct {
	Mysql *gorm.DB
	Redis *redis.Client
}

func NewData(c *conf.Data, logger log.Logger, redisCli *redis.Client, mysqlCli *gorm.DB) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{
		Mysql: mysqlCli,
		Redis: redisCli,
	}, cleanup, nil
}

func NewMysql(c *conf.Data, logger log.Logger) *gorm.DB {
	defer func() {
		if err := recover(); err != nil {
			log.NewHelper(logger).Errorw("kind", "mysql", "error", err)
		}
	}()
	zapLog := pkg.NewZapGormLogV2(log.NewHelper(logger), "mysql")
	gormConfig := &gorm.Config{
		Logger: zapLog,
	}
	// mysql数据库连接
	db, err := gorm.Open(mysql.Open(c.Database.Source), gormConfig)
	if err != nil {
		panic(err)
	}
	db.Use(&pkg.TracePlugin{})
	sqlDb, err := db.DB()
	if err != nil {
		panic(err)
	}
	sqlDb.SetMaxIdleConns(int(c.Database.MinIdleConns))
	sqlDb.SetMaxOpenConns(int(c.Database.MaxOpenConns))
	sqlDb.SetConnMaxLifetime(time.Hour * time.Duration(c.Database.ConMaxLeftTime))
	log.NewHelper(logger).Infow("kind", "mysql", "status", "enable")
	return db
}

func NewRedis(c *conf.Data, logger log.Logger) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:         c.Redis.Addr,
		Password:     c.Redis.Password,
		DB:           int(c.Redis.Db),
		PoolSize:     int(c.Redis.PoolSize),     // 连接池数量
		MinIdleConns: int(c.Redis.MinIdleConns), //好比最小连接数
		MaxRetries:   int(c.Redis.MaxRetries),   // 命令执行失败时，最多重试多少次，默认为0即不重试

	})
	rdb.AddHook(redisotel.NewTracingHook(
		redisotel.WithAttributes(
			attribute.String("db.type", "redis"),
			attribute.String("db.ip", c.Redis.Addr),
			attribute.String("db.instance", c.Redis.Addr+"/"+convertor.ToString(c.Redis.Db)),
		),
	))
	log.NewHelper(logger).Infow("kind", "redis", "status", "enable")
	return rdb
}
