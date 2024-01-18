package biz

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"errors"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/spf13/cast"
)

const (
	UserTableName = "user"

	// 批量创建时的batchSize,一次批量create的数量
	BatchCreateUserSize = 2
)

type User struct {
	Id       uint32    `gorm:"primaryKey;column:id" json:"id" structs:"id"`
	Name     string    `gorm:"column:name" json:"name" structs:"name" validate:"required,min=1"`
	Age      int8      `gorm:"column:age" json:"age" structs:"age" validate:"required,min=1"`
	Desc     string    `gorm:"column:desc" json:"desc" structs:"desc"`
	CreateAt time.Time `gorm:"column:create_at;default:null" json:"createAt" structs:"create_at"`
	UpdateAt time.Time `gorm:"column:update_at;default:null" json:"updateAt" structs:"update_at"`
}

type GinUsecase struct {
	log *log.Helper

	ginRepo GinRepo
}

type GinRepo interface {
	BatchCreateExcelDataToDBAndRedis(ctx context.Context, users []*User) error

	// ListFruits(ctx context.Context) (any, error)
}

func NewGinUseCase(logger log.Logger, ginRepo GinRepo) *GinUsecase {
	return &GinUsecase{
		log:     log.NewHelper(logger),
		ginRepo: ginRepo,
	}
}

// GinUsecase的方法
func (gc *GinUsecase) SaveExcelData(ctx context.Context, xlsxFile *excelize.File) error {

	// 测试数据
	//userLst := gc.genTestUserLst()

	//解析excel的数据
	userLst, lxRrr := gc.readExcel(xlsxFile)
	if lxRrr != nil {
		return errors.New(fmt.Sprintf("解析excel文件失败err: %v", lxRrr))
	}

	// 写入数据库与redis
	errBatchCreate := gc.ginRepo.BatchCreateExcelDataToDBAndRedis(ctx, userLst)
	if errBatchCreate != nil {
		return errBatchCreate
	}

	return nil
}

// readExcel 读取excel
func (gc *GinUsecase) readExcel(xlsx *excelize.File) ([]*User, error) {
	// 根据名字获取cells的内容，返回的是一个[][]string
	rows := xlsx.GetRows(xlsx.GetSheetName(xlsx.GetActiveSheetIndex()))

	var users []*User
	for i, row := range rows {
		// Notice 去掉第一行是excel表头部分
		if i == 0 {
			continue
		}

		var data User

		name := row[0]
		age := row[1]
		desc := row[2]

		if name == "" || age == "" {
			continue
		}

		data.Name = name
		data.Age = cast.ToInt8(age)
		data.Desc = desc

		//for k, v := range row {
		//	// 第一列是name
		//	if k == 0 {
		//		data.Name = v
		//	}
		//	// 第二列是age
		//	if k == 1 {
		//		data.Age = cast.ToInt8(v)
		//	}
		//	// 第三列是desc
		//	if k == 2 {
		//		data.Desc = v
		//	}
		//}
		//将数据追加到集合中
		users = append(users, &data)
	}
	return users, nil
}

func (gc *GinUsecase) genTestUserLst() []*User {
	u1 := &User{
		Name: "whw1",
		Age:  21,
		Desc: "xxx",
	}
	u2 := &User{
		Name: "whw2",
		Age:  22,
		Desc: "xxx123",
	}
	u3 := &User{
		Name: "whw3",
		Age:  23,
		Desc: "xxx666",
	}
	u4 := &User{
		Name: "whw4",
		Age:  24,
		Desc: "xxxwrqwer",
	}
	u5 := &User{
		Name: "whw5",
		Age:  25,
		Desc: "asdfasdf",
	}
	userLst := []*User{u1, u2, u3, u4, u5}

	return userLst
}
