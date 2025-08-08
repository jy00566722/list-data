package model

import "time"

// Product represents the products table
type Product struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ShopName        string    `gorm:"type:varchar(255);not null;uniqueIndex:uk_shop_skc" json:"shop_name"`
	SKC             string    `gorm:"type:varchar(255);not null;uniqueIndex:uk_shop_skc" json:"skc"`
	SPU             string    `gorm:"type:varchar(255)" json:"spu"`
	ItemCode        string    `gorm:"type:varchar(255)" json:"item_code"`
	ProductName     string    `gorm:"type:varchar(512)" json:"product_name"`
	ProductImageURL string    `gorm:"type:varchar(1024)" json:"product_image_url"`
	Category        string    `gorm:"type:varchar(255)" json:"category"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
