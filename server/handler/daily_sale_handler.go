package handler

import (
	"log"
	"net/http"
	"data-list/server/database"
	"data-list/server/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

// CreateDailySaleSku 接收并处理批量上报的SKU日销量数据
// 已优化为支持 "Upsert" (Insert on duplicate Update)
func CreateDailySaleSku(c *gin.Context) {
	var sales []model.DailySaleSku

	if err := c.ShouldBindJSON(&sales); err != nil {
		log.Printf("!!! 数据绑定失败: %v", err)  // 新增日志
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data: " + err.Error(),
		})
		return
	}

	if len(sales) == 0 {
		log.Printf("!!! 收到空的销售数据数组") // 新增日志
		// 如果没有数据，直接返回错误
		// 这可以防止不必要的数据库操作
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Received an empty sales array.",
		})
		return
	}

	// 使用GORM的 "Upsert" 功能进行批量插入或更新
	// OnConflict 子句是这里的核心：
	// - Columns: 指定冲突判断的列，即我们设置的唯一索引 (sales_date, sku)
	// - DoUpdates: 指定当冲突发生时，需要更新哪些列。
	//   clause.AssignmentColumns([]string{"sales_number"}) 表示只更新 sales_number 字段，
	//   其他字段（如 created_at）将保持不变。
	result := database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "sales_date"}, {Name: "sku"}},
		DoUpdates: clause.AssignmentColumns([]string{"sales_number"}),
	}).Create(&sales)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to save or update data: " + result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Successfully saved or updated sales data.",
		"records_processed": result.RowsAffected,
	})
}
