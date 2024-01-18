package pkg

import (
	"strings"

	"github.com/go-kratos/kratos/v2/transport"
	mysql2 "github.com/go-sql-driver/mysql"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	gormSpanKey        = "Gorm"
	callBackBeforeName = "tracing:before"
	callBackAfterName  = "tracing:after"
)

func before(db *gorm.DB) {
	ctx := db.Statement.Context
	if _, ok := transport.FromServerContext(ctx); ok {
		_, span := otel.Tracer(gormSpanKey).Start(ctx, ""+db.Statement.Table, trace.WithSpanKind(trace.SpanKindInternal))
		db.InstanceSet(gormSpanKey, span)
	}
}

func after(db *gorm.DB) {
	// 从GORM的DB实例中取出span
	_span, isExist := db.InstanceGet(gormSpanKey)
	if !isExist {
		return
	}
	// 断言进行类型转换
	span, ok := _span.(trace.Span)
	if !ok {
		return
	}
	ctx := db.Statement.Context
	if _, ok := transport.FromServerContext(ctx); !ok {
		return
	}
	defer span.End()
	// Error
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		span.RecordError(db.Error)
		span.SetStatus(codes.Error, db.Error.Error())
	}
	dsn, _ := mysql2.ParseDSN(db.Dialector.(*mysql.Dialector).DSN)
	adds := strings.Split(dsn.Addr, ":")
	// sql
	span.SetAttributes(attribute.String("db.type", "mysql"))
	span.SetAttributes(attribute.String("db.name", dsn.DBName))
	span.SetAttributes(attribute.String("db.user", dsn.User))
	span.SetAttributes(attribute.String("db.ip", adds[0]))
	span.SetAttributes(attribute.String("db.port", adds[1]))
	span.SetAttributes(attribute.String("db.instance", dsn.DBName))
	span.SetAttributes(attribute.String("db.statement", db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)))
	span.SetAttributes(attribute.Int64("rowsAffected", db.Statement.RowsAffected))
}

type TracePlugin struct{}

func (op *TracePlugin) Name() string {
	return "tracingPlugin"
}

func (op *TracePlugin) Initialize(db *gorm.DB) (err error) {
	// 开始前
	db.Callback().Create().Before("gorm_plugin:before_create").Register(callBackBeforeName, before)
	db.Callback().Query().Before("gorm_plugin:query").Register(callBackBeforeName, before)
	db.Callback().Delete().Before("gorm_plugin:before_delete").Register(callBackBeforeName, before)
	db.Callback().Update().Before("gorm_plugin:setup_reflect_value").Register(callBackBeforeName, before)
	db.Callback().Row().Before("gorm_plugin:row").Register(callBackBeforeName, before)
	db.Callback().Raw().Before("gorm_plugin:raw").Register(callBackBeforeName, before)

	// 结束后
	db.Callback().Create().After("gorm_plugin:after_create").Register(callBackAfterName, after)
	db.Callback().Query().After("gorm_plugin:after_query").Register(callBackAfterName, after)
	db.Callback().Delete().After("gorm_plugin:after_delete").Register(callBackAfterName, after)
	db.Callback().Update().After("gorm_plugin:after_update").Register(callBackAfterName, after)
	db.Callback().Row().After("gorm_plugin:row").Register(callBackAfterName, after)
	db.Callback().Raw().After("gorm_plugin:raw").Register(callBackAfterName, after)
	return
}

var _ gorm.Plugin = &TracePlugin{}
