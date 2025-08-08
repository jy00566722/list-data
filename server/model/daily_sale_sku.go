package model

import "time"

// DailySaleSku 对应于 daily_sales_sku 表，用于收集原始上报的SKU日销量数据。
// 这个模型严格遵循“收集器”原则，只记录最原始的数据，不做任何聚合。
type DailySaleSku struct {
	ID          uint      `gorm:"primarykey"`
	// 为 SalesDate 和 Sku 添加了复合唯一索引 (uk_date_sku)，确保了同一SKU在同一天只有一条记录。
	SalesDate   string    `json:"sales_date" gorm:"uniqueIndex:uk_date_sku;type:varchar(20);comment:销售日期"`
	Sku         string    `json:"sku" gorm:"uniqueIndex:uk_date_sku;type:varchar(255);comment:商品SKU"`
	SalesNumber int       `json:"sales_number" gorm:"comment:销售数量"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"` // GORM会自动在创建和更新时更新该字段
}

// TableName 自定义 GORM 使用的表名
func (DailySaleSku) TableName() string {
	return "daily_sales_sku"
}