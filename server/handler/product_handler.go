package handler

import (
	"data-list/server/database"
	"data-list/server/model"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm/clause"
)

// GetProducts 从数据库中检索所有商品。
func GetProducts(c *gin.Context) {
	var products []model.Product
	if err := database.DB.Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取商品列表失败"})
		return
	}
	c.JSON(http.StatusOK, products)
}

// DeleteProduct 根据ID删除一个商品。
func DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的商品ID"})
		return
	}

	if err := database.DB.Delete(&model.Product{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除商品失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "商品删除成功"})
}

// UploadProducts 处理上传Excel文件以新增或更新商品。
// 它基于SKU使用"Upsert"机制来避免重复。
func UploadProducts(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件上传失败"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法打开上传的文件"})
		return
	}
	defer src.Close()

	f, err := excelize.OpenReader(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法读取Excel文件"})
		return
	}

	// 从第一个工作表中获取所有行。
	rows, err := f.GetRows(f.GetSheetName(0))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法从工作表获取行"})
		return
	}

	var productsToUpsert []model.Product
	log.Printf("开始处理Excel文件，总行数: %d", len(rows))

	// 跳过表头行 (i=0)
	for i, row := range rows {
		if i == 0 {
			log.Println("跳过表头行:", row)
			continue
		}

		// --- 新增日志 ---
		// log.Printf("正在处理第 %d 行, 列数: %d, 内容: %v", i+1, len(row), row)

		// 基本验证：确保行有足够的列。
		// 根据您的Excel文件结构调整此数字。
		if len(row) < 12 {
			log.Printf("第 %d 行因列数不足 (%d < 12) 而被跳过", i+1, len(row))
			continue
		}

		shopID, _ := strconv.ParseInt(row[0], 10, 64)

		product := model.Product{
			ShopID:   shopID,
			ShopCode: row[1],
			SPU:      row[2],
			SKC:      row[3],
			SKU:      row[4],
			SkcCode:  row[5],
			SkuCode:  row[6],
			ColorCN:  row[7],
			ColorEN:  row[8],
			Size:     row[9],
			ImageURL: row[10],
			BarCode:  row[11],
		}
		productsToUpsert = append(productsToUpsert, product)
	}

	if len(productsToUpsert) == 0 {
		log.Println("文件中未找到有效的商品数据")
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件中未找到有效的商品数据"})
		return
	}

	// 使用GORM的OnConflict (Upsert)功能。
	// 如果存在相同SKU的商品，则更新所有其他字段。
	// 否则，插入一个新商品。
	err = database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "sku"}},
		DoUpdates: clause.AssignmentColumns([]string{"shop_id", "shop_code", "spu", "skc", "skc_code", "sku_code", "color_cn", "color_en", "size", "image_url", "bar_code"}),
	}).Create(&productsToUpsert).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("数据库操作失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("成功处理 %d 件商品。", len(productsToUpsert)),
	})
}
