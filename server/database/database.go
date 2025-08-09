package database

import (
	"fmt"
	"log"
	"data-list/server/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// DB 是一个全局的数据库连接实例
var DB *gorm.DB

// Init 初始化数据库连接和自动迁移
func Init() {
	var err error
	// TODO: 建议将数据库配置信息移至配置文件或环境变量中
	// DSN (Data Source Name) 格式: "user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	// 请根据你的本地MySQL环境修改这里的 DSN
	dsn := "root:123456@tcp(127.0.0.1:3306)/gemini_data_list?charset=utf8mb4&parseTime=True&loc=Local"

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("Database connection established successfully.")

	// 自动迁移模式，GORM会自动创建或更新表结构以匹配模型
	fmt.Println("Running auto migration...")
	err = DB.AutoMigrate(&model.DailySaleSku{}, &model.Product{})
	if err != nil {
		log.Fatalf("Failed to auto migrate database: %v", err)
	}
	fmt.Println("Auto migration completed.")
}
