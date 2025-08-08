package main

import (
	"data-list/server/database"
	"data-list/server/router"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 初始化数据库连接
	database.Init()

	// 2. 创建 Gin 引擎
	r := gin.Default()

	// 3. 添加CORS中间件
	// 这是解决浏览器扩展与本地服务器通信的关键。
	// 它允许所有来源(AllowAllOrigins)和所有必要的头(Headers)，
	// 确保浏览器不会阻止前端的请求。
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(config))


	// 4. 注册所有在 router 包中定义的路由
	router.SetupRouter(r)

	// 5. 启动服务
	// 默认监听并在 0.0.0.0:8080 上启动
	r.Run()
}
