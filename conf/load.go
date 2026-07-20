package conf

import (
	"fmt"
	"os"
	"path/filepath"
)

// env 用于区分生产/研发环境的变量 默认为开发环境
// 生产环境加载配置文件: config.prod.yaml
// 研发环境加载配置文件: config.dev.yaml
var env = "dev"

// configFileDir 配置文件所在路径
var configFileDir string

// configFileName 配置文件名称
var configFileName string

// resolveConfigFileName 本函数用于根据环境变量 SOLAR_NOTES_ENV 的值确认当前环境,
// 并根据确认结果设置 env 变量的值和 configFileName 的值(研发环境保持 env 变量的值为"dev"不变)
func resolveConfigFileName() error {
	envVariable := os.Getenv("SOLAR_NOTES_ENV")
	if envVariable == "prod" {
		env = "prod"
		configFileName = "config." + env + ".yaml"
		return nil
	}

	if envVariable == "dev" || envVariable == "" {
		configFileName = "config." + env + ".yaml"
		return nil
	}

	return fmt.Errorf("无法识别的 SOLAR_NOTES_ENV 值: %q (仅接受dev/prod)", envVariable)
}

// findConfigFileDir 本函数用于查找配置文件所在路径
// 优先根据可执行文件所在路径查找配置文件路径,若找不到则使用工作路径查找配置文件路径
func findConfigFileDir() error {
	configFilePath, exeErr := findDirByExecutable()
	if exeErr == nil {
		configFileDir = configFilePath
		return nil
	}

	configFilePath, wdErr := findDirByWorkingDir()
	if wdErr == nil {
		configFileDir = configFilePath
		return nil
	}

	return fmt.Errorf("未找到配置文件 %s:\n  按可执行文件目录查找: %v\n  按工作目录查找: %v", configFileName, exeErr, wdErr)
}

// findDirByExecutable 本函数用于根据可执行文件的路径查找配置文件所在路径
func findDirByExecutable() (string, error) {
	executablePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	executableDir := filepath.Dir(executablePath)

	// 检测配置文件是否存在
	configFilePath := filepath.Join(executableDir, configFileName)
	_, err = os.Stat(configFilePath)
	if err != nil {
		return "", err
	}

	return executableDir, nil
}

// findDirByWorkingDir 本函数用于根据当前进程所在的工作目录查找配置文件所在路径
func findDirByWorkingDir() (string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// 检测配置文件是否存在
	configFilePath := filepath.Join(workingDir, configFileName)
	_, err = os.Stat(configFilePath)
	if err != nil {
		return "", err
	}

	return workingDir, nil
}
