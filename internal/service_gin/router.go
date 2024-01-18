package service_gin

import (
	"github.com/gin-gonic/gin"
)

func RegisterHTTPServer(us *GinUseCase) *gin.Engine {
	router := gin.New()

	rootGrp := router.Group("/api")

	// 用户相关API
	userGrp := rootGrp.Group("/user")
	// helloWorld
	userGrp.GET("/sayhi", us.helloKratosGin)
	// 上传excel文件存入DB与Redis中
	userGrp.POST("/uploadUser", us.uploadExcelUsers)
	// 使用gin下载
	// userGrp.GET("/downloadFileGin", us.DownloadFileGin)
	//ws
	// userGrp.GET("/ws", us.Ws)

	return router
}
