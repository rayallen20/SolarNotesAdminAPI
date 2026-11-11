package db

import (
	"SolarNotesAdminAPI/conf"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// DB 数据库连接句柄
var DB *gorm.DB

// Connect 本函数用于根据数据库配置连接数据库
func Connect(dbConf conf.Database) error {
	dsn := buildDSN(dbConf)

	loggerConf := logger.Config{
		SlowThreshold:             time.Second, // 慢 SQL 阈值
		LogLevel:                  logger.Warn, // 日志级别
		IgnoreRecordNotFoundError: true,        // 忽略ErrRecordNotFound(记录未找到)错误
		ParameterizedQueries:      true,        // 启用参数化查询
		Colorful:                  false,       // 彩色打印
	}
	queryLogger := logger.New(log.New(os.Stdout, "[gorm] ", log.LstdFlags), loggerConf)

	opts := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 全局禁用表名复数
		},
		Logger: queryLogger,
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
