package service_gin

import (
	"fmt"

	"github.com/360EntSecGroup-Skylar/excelize"

	"github.com/gin-gonic/gin"
)

type UploadExcelReq struct {
	Username string `json:"username" uri:"username" form:"username" binding:"required,min=1"`
	Password string `json:"password" uri:"password" form:"password" binding:"required,min=1"`
	// Notice binding标签为required的话,传来的age不能为空值：0 / 空字符串
	Age int `json:"age" uri:"age" form:"age" binding:"min=1"`
}

// 简单的helloword
func (g *GinUseCase) helloKratosGin(c *gin.Context) {

	/*
		// Notice 虽然可以这样做，但是还是不建议，既然集成了gin框架，请求与响应就都用gin来处理吧～
		// 可以使用gin解析http请求中的参数，然后调用GreeterService中的方法去处理～
		in := &v1.HelloRequest{}
		// Notice context可以直接用gin的
		g.srv.SayHello(c, in)
	*/

	// 获取参数（GET请求中querystring的参数、POST请求中如果Content-Type是application/json那么 json/form表单的参数都可以用这个方法获取）
	userName := c.DefaultQuery("username", "")

	c.JSON(200, gin.H{
		"msg": "success",
		"data": gin.H{
			"greetings": fmt.Sprintf("hello %v", userName),
		},
	})

}

// 使用gin上传excel
func (g *GinUseCase) uploadExcelUsers(c *gin.Context) {

	// Notice 线上的话，做一下文件大小的限制，最好在前端限制下～必要的话在后台也读一下文件流的大小做限制
	// 或者限制一下用户数量：比如在后面的逻辑中发现用户数超过了1w，直接返回错误
	file, _, err := c.Request.FormFile("file")
	if err != nil {

		c.JSON(450, gin.H{
			"msg": "failed",
			"err": fmt.Sprintf("上传文件失败err: %v", err),
		})
		return
	}
	//读excel流
	xlsx, err := excelize.OpenReader(file)
	if err != nil {
		c.JSON(450, gin.H{
			"msg":   "failed",
			"error": fmt.Sprintf("读取excel文件失败err: %v", err),
		})
		c.JSON(200, gin.H{
			"msg":  "success",
			"data": "上传成功!",
		})
		return
	}

	// Notice 可以使用gin的context了！
	err = g.srv.Gc.SaveExcelData(c, xlsx)
	if err != nil {
		//c.JSON(ErrCodeCommon, gin.H{
		//	"msg":   "failed",
		//	"error": fmt.Sprintf("保存excel数据失败err: %v", err),
		//})
		c.JSON(500, gin.H{
			"msg":   "failed",
			"error": fmt.Sprintf("保存excel数据失败err: %v", err),
		})
		return
	}

	//c.JSON(ErrCodeCommon, gin.H{
	//	"msg":  "success",
	//	"data": "上传成功!",
	//})
	c.JSON(500, gin.H{
		"msg":  "success",
		"data": "上传成功!",
	})

}

// // 1、使用gin框架下载
// func (g *GinUseCase) DownloadFileGin(c *gin.Context) {

// 	// Notice 到时候换成 服务器 或 项目部署所在的docker容器(先把文件COPY进去) 里面文件的绝对路径
// 	filePath := "/Users/wanghongwei/Documents/my-kratos-demos/kratosGin/IMG_1984.jpeg"

// 	file, errFile := os.Open(filePath)
// 	if errFile != nil {
// 		c.JSON(500, Response{
// 			Code:    ErrCodeCommon,
// 			Reason:  fmt.Sprintf("读取文件出错:%v", errFile),
// 			Message: nil,
// 		})
// 		return
// 	}
// 	defer file.Close()

// 	fileContentDisposition := fmt.Sprintf("attachment;filename=%v", "蓝天.jpeg")
// 	c.Header("Content-Disposition", fileContentDisposition)

// 	// Notice 1、本地文件的写法
// 	//c.File(filePath)

// 	// Notice 2、如果是从oss中读取的二进制流数据，这样写
// 	body, _ := io.ReadAll(file) // 模拟二进制流数据
// 	_, errW := c.Writer.Write(body)
// 	if errW != nil {
// 		c.JSON(500, Response{
// 			Code:    ErrCodeCommon,
// 			Reason:  fmt.Sprintf("下载文件出错:%v", errFile),
// 			Message: nil,
// 		})
// 		return
// 	}

//		return
//	}
//

// 使用websocket
// func (g *GinUseCase) Ws(c *gin.Context) {
// 	// Notice 1、升级get请求为webSocket协议
// 	ws, err := g.upgrader.Upgrade(c.Writer, c.Request, nil)
// 	if err != nil {
// 		fmt.Println("升级get请求为webSocket协议失败err: ", err)
// 		return
// 	}
// 	defer ws.Close()

// 	// Notice 2、读取ws中的数据
// 	for {
// 		// 读取ws中的数据
// 		_, msg, err := ws.ReadMessage()
// 		if err != nil {
// 			fmt.Println("读取ws中的数据失败err: ", err)
// 			return
// 		}
// 		fmt.Println("读取ws中的数据msg: ", string(msg))

// 		// Notice 3、写入ws中的数据
// 		err = ws.WriteMessage(websocket.TextMessage, []byte("hello client!"))
// 		if err != nil {
// 			fmt.Println("写入ws中的数据失败err: ", err)
// 			return
// 		}
// 	}

// }
