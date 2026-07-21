package db

import (
	"SolarNotesAdminAPI/conf"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// DB 数据库连接句柄
var DB *gorm.DB

// Connect 本函数用于根据数据库配置连接数据库
func Connect(dbConf conf.Database) error {
	dsn := buildDSN(dbConf)
	opts := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 全局禁用表名复数
		},
	}

	// Tips: gorm.Open()函数是急连接,初始化就会执行`SELECT VERSION()`,所以后续不再需要手动ping测试了
	db, err := gorm.Open(mysql.Open(dsn), opts)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	// 同时允许的最大连接数
	sqlDB.SetMaxOpenConns(10)
	// 允许存活的空闲连接数
	sqlDB.SetMaxIdleConns(5)
	// 连接的最长存活时间
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = db

	return nil
}

// buildDSN 本函数用于根据数据库配置生成连接参数
func buildDSN(dbConf conf.Database) string {
	dsn := dbConf.User + ":" + dbConf.Password + "@tcp(" + dbConf.Host +
		":" + dbConf.Port + ")/" + dbConf.Name + "?charset=utf8mb4&parseTime=True&loc=Local"

	return dsn
}
