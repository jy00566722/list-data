package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// JSONB allows storing JSON data in a database
type JSONB json.RawMessage

// Value implements the driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.RawMessage(j).MarshalJSON()
}

// Scan implements the sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, j)
}

// DailySale represents the daily_sales table
type DailySale struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProductID uint      `gorm:"not null;uniqueIndex:uk_product_date" json:"product_id"`
	Product   Product   `gorm:"foreignKey:ProductID"`
	SalesDate time.Time `gorm:"type:date;not null;uniqueIndex:uk_product_date" json:"sales_date"`
	SalesCount int      `gorm:"not null" json:"sales_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	RawData   JSONB     `gorm:"type:json" json:"raw_data"`
}
