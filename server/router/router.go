package router

import (
	"data-list/server/handler"

	"github.com/gin-gonic/gin"
)

// SetupRouter 配置项目的所有路由
func SetupRouter(r *gin.Engine) {
	// 设置一个基础的 /api/v1 路由组
	apiV1 := r.Group("/api/v1")
	{
		// 数据上报相关路由
		salesGroup := apiV1.Group("/sales")
		{
			// POST /api/v1/sales/daily-sku
			// 用于接收前端插件上报的每日SKU销量数据
			salesGroup.POST("/daily-sku", handler.CreateDailySaleSku)
		}

		// 可以在这里继续添加其他路由组, 例如 /api/v1/products 等
	}

	// 添加一个简单的ping路由用于健康检查
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
}
