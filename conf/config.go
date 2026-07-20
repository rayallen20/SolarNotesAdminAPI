package conf

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"
)

// Config 配置对象
type Config struct {
	Database Database `yaml:"database"`
	Asset    Asset    `yaml:"asset"`
	Storage  Storage  `yaml:"storage"`
}

// Load 本方法用于加载配置文件至 Config 实例
func (c *Config) Load() error {
	confirmName()

	// 获取配置文件所在路径
	err := getConfigFilePath()
	if err != nil {
		return err
	}

	return c.bind()
}

// bind 本方法用于绑定配置文件到 Config 结构体实例上
func (c *Config) bind() error {
	file, err := os.Open(filepath.Join(path, name))
	if err != nil {
		return err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Println("关闭配置文件失败: " + err.Error())
		}
	}()

	decoder := yaml.NewDecoder(file)
	// 开启严格模式 确保yaml中所有的键名在结构体中都存在
	decoder.KnownFields(true)
	err = decoder.Decode(c)
	if err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	return nil
}
