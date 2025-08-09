package handler

import (
	"data-list/server/database"
	"data-list/server/model"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm/clause"
)

// GetProducts retrieves all products from the database.
func GetProducts(c *gin.Context) {
	var products []model.Product
	if err := database.DB.Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve products"})
		return
	}
	c.JSON(http.StatusOK, products)
}

// DeleteProduct deletes a product by its ID.
func DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	if err := database.DB.Delete(&model.Product{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

// UploadProducts handles uploading an Excel file to add or update products.
// It uses an "Upsert" mechanism based on the SKU to avoid duplicates.
func UploadProducts(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File upload failed"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
		return
	}
	defer src.Close()

	f, err := excelize.OpenReader(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read Excel file"})
		return
	}

	// Get all rows from the first sheet.
	rows, err := f.GetRows(f.GetSheetName(0))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get rows from sheet"})
		return
	}

	var productsToUpsert []model.Product
	// Skip header row (i=0)
	for i, row := range rows {
		if i == 0 {
			continue
		}

		// Basic validation: ensure the row has enough columns.
		// Adjust the number based on your Excel file's structure.
		if len(row) < 13 {
			continue // Or handle error
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "No valid product data found in the file"})
		return
	}

	// Use GORM's OnConflict (Upsert) feature.
	// If a product with the same SKU exists, update all other fields.
	// Otherwise, insert a new product.
	err = database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "sku"}},
		DoUpdates: clause.AssignmentColumns([]string{"shop_id", "shop_code", "spu", "skc", "skc_code", "sku_code", "color_cn", "color_en", "size", "image_url", "bar_code"}),
	}).Create(&productsToUpsert).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to upsert products: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Successfully processed %d products.", len(productsToUpsert)),
	})
}
