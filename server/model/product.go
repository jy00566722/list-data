package model

import "time"

// Product represents the products table, aligning with the detailed structure in README.md
type Product struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ShopID      int64     `gorm:"comment:店铺ID" json:"shop_id"`
	ShopCode    string    `gorm:"type:varchar(255);comment:店铺代码" json:"shop_code"`
	SPU         string    `gorm:"type:varchar(255);comment:商品SPU" json:"spu"`
	SKC         string    `gorm:"type:varchar(255);comment:商品SKC" json:"skc"`
	SKU         string    `gorm:"type:varchar(255);uniqueIndex;comment:商品SKU" json:"sku"`
	SkcCode     string    `gorm:"type:varchar(255);comment:商品SKC货号" json:"skc_code"`
	SkuCode     string    `gorm:"type:varchar(255);comment:店铺SKU货号" json:"sku_code"`
	ColorCN     string    `gorm:"type:varchar(255);comment:商品颜色.中文" json:"color_cn"`
	ColorEN     string    `gorm:"type:varchar(255);comment:商品颜色.英文" json:"color_en"`
	Size        string    `gorm:"type:varchar(255);comment:商品尺码" json:"size"`
	ImageURL    string    `gorm:"type:varchar(1024);comment:商品缩略图url" json:"image_url"`
	BarCode     string    `gorm:"type:varchar(255);comment:商品条码编码" json:"bar_code"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
