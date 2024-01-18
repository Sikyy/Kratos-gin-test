package data

import (
	"Gin-server-text/internal/biz"
	"Gin-server-text/internal/utils"
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm/clause"
)

const (
	// user:(name):(age), value固定存一个"OK"
	userKeyFormat = "user:%v:%v"
	userKeyTTL    = time.Hour * 24
)

type ginRepo struct {
	data *Data
	log  *log.Helper
}

func NewGinRepo(data *Data, logger log.Logger) biz.GinRepo {
	return &ginRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (g *ginRepo) BatchCreateExcelDataToDBAndRedis(ctx context.Context, users []*biz.User) error {

	var err error

	db := g.data.Mysql.Table(biz.UserTableName).WithContext(ctx)
	// Notice 批量创建的size
	db.CreateBatchSize = biz.BatchCreateUserSize

	// 1、写入数据库
	// 1-1、方案一: 如果数据中混入了一些唯一索引冲突的数据，会报错，这样不太友好！
	//err := db.Create(&users).Error

	// Notice 1-2、方案二: 遇到联合索引冲突不做任何操作，仅插入"新"数据 不改动"老"数据
	// Notice 注意数据库必须设置 联合唯一索引 否则这个语句不会生效
	// 这种方案下,如果插入的一批数据中遇到联合索引冲突的不会修改原来的数据也不会插入,新数据可以写入到数据库
	// 这种方案也有一个好处: 如果下面写入redis出错了，再导入一遍并不会影响原来导入到数据库中的数据
	// 另外需要注意一个点：ON DUPLICATE KEY 这个操作也是"事务"操作，如果单独使用该语句的话不用再把它放到事务中执行了
	err = db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}, {Name: "age"}},
		DoNothing: true,

		// Notice 拓展:
		// 1、既给age+1 又修改desc字段
		// 如果遇到唯一索引冲突就给相应的age加1，并且desc设置成传来的desc，注意VALUES方法中是`desc`
		//DoUpdates: clause.Assignments(map[string]interface{}{
		//	"age":  gorm.Expr("age + ?", 1),
		//	"desc": gorm.Expr("VALUES(`desc`)"),
		//}),
		// 2、 只修改字段
		// DoUpdates: clause.AssignmentColumns([]string{"age", "desc"}),
	}).Create(&users).Error

	// Notice 批量创建后,不能返回新增(或者不变)用户的id！
	//for _, user := range users {
	//	fmt.Println(user.Id, user.Name, user.Age, user.Desc)
	//}

	// 2、写入redis
	// Notice 如果一次写入的数据特别多的话，建议分批次存储！pipeline一次存100条数据
	pipe := g.data.Redis.Pipeline()
	for i := 0; i < len(users); i += 100 {
		// Notice 防止切片越界：panic 或 获取到零值
		currUsers := users[i:utils.GetMinInt(i+100, len(users))]
		for _, userModel := range currUsers {
			cacheKey := fmt.Sprintf(userKeyFormat, userModel.Name, userModel.Age)
			pipe.Set(ctx, cacheKey, "OK", userKeyTTL)
		}
		_, errExec := pipe.Exec(ctx)
		// Notice redis操作直接返回错误还是continue 看实际的情况吧！
		if errExec != nil {
			return errExec
		}
	}

	return err
}
