package conf

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

// loaded 包级别配置实例, Load 函数调用成功后持有
var loaded *Config

// Load 本函数用于加载配置文件并持有为包级别实例 服务启动时调用1次即可
func Load() error {
	c := &Config{}

	err := c.load()
	if err != nil {
		return err
	}

	loaded = c
	return nil
}

// Get 本函数用于获取已加载的配置实例 (在 Load 函数调用成功后调用,否则获取到的实例为nil)
func Get() *Config {
	return loaded
}

// Config 配置对象
type Config struct {
	Service   Service   `yaml:"service"`
	Database  Database  `yaml:"database"`
	JWT       JWT       `yaml:"jwt"`
	Asset     Asset     `yaml:"asset"`
	Storage   Storage   `yaml:"storage"`
	Snowflake Snowflake `yaml:"snowflake"`
}

// load 本方法用于加载配置文件至 Config 实例
func (c *Config) load() error {
	// 根据环境确认配置文件名
	err := resolveConfigFileName()
	if err != nil {
		return err
	}

	// 获取配置文件所在路径
	err = findConfigFileDir()
	if err != nil {
		return err
	}

	return c.bind()
}

// bind 本方法用于绑定配置文件到 Config 结构体实例上,然后进行规范化和校验
func (c *Config) bind() error {
	file, err := os.Open(filepath.Join(configFileDir, configFileName))
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

	c.normalize()

	return c.validate()
}

// validate 本方法用于检查配置文件中的配置项值是否合法
func (c *Config) validate() error {
	err := c.Service.validate()
	if err != nil {
		return err
	}

	err = c.Database.validate()
	if err != nil {
		return err
	}

	err = c.JWT.validate()
	if err != nil {
		return err
	}

	err = c.Asset.validate()
	if err != nil {
		return err
	}

	err = c.Storage.validate()
	if err != nil {
		return err
	}

	err = c.Snowflake.validate()
	if err != nil {
		return err
	}

	return nil
}

// normalize 本方法用于对配置文件中的部分配置项进行规范化处理
// 1. 去除 asset.base_url 的末尾斜杠
// 2. 去除 storage.root_dir 的末尾斜杠
func (c *Config) normalize() {
	c.Asset.BaseURL = strings.TrimRight(c.Asset.BaseURL, "/")
	c.Storage.RootDir = strings.TrimRight(c.Storage.RootDir, "/")
}
