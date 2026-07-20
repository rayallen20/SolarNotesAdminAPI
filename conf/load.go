package conf

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

// env 用于区分生产/研发环境的变量 默认为开发环境
// 生产环境加载配置文件: config.prod.yaml
// 研发环境加载配置文件: config.dev.yaml
var env = "dev"

// path 配置文件所在路径
var path string

// name 配置文件名称
var name string

// confirmName 本函数用于确认当前环境,并根据确认结果设置 env 变量的值和 name 的值
// (研发环境保持 env 变量的值为"dev"不变)
func confirmName() {
	envVariable := os.Getenv("SOLAR_NOTES_ENV")
	if envVariable == "prod" {
		env = "prod"
	}

	name = "config." + env + ".yaml"
}

// getConfigFilePath 本函数用于获取配置文件所在路径
func getConfigFilePath() error {
	_, file, _, ok := runtime.Caller(2)
	if !ok {
		err := errors.New("无法读取配置文件所在路径")
		return err
	}

	path = filepath.Dir(file)
	return nil
}
