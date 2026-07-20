package conf

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

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
// 1. 检查 database 部分是否有空项
// 2. 检查 asset.base_url 是否为合法的URL
// 3. 检查 storage.root_dir 是否为绝对路径
func (c *Config) validate() error {
	err := c.checkDatabase()
	if err != nil {
		return err
	}

	err = c.checkAsset()
	if err != nil {
		return err
	}

	err = c.checkStorage()
	if err != nil {
		return err
	}

	return nil
}

// checkDatabase 本方法用于检查配置文件中的 database 部分是否有空项
func (c *Config) checkDatabase() error {
	var emptyKeys []string
	if c.Database.Host == "" {
		emptyKeys = append(emptyKeys, "host")
	}

	if c.Database.Port == "" {
		emptyKeys = append(emptyKeys, "port")
	}

	if c.Database.User == "" {
		emptyKeys = append(emptyKeys, "user")
	}

	if c.Database.Password == "" {
		emptyKeys = append(emptyKeys, "password")
	}

	if c.Database.Name == "" {
		emptyKeys = append(emptyKeys, "configFileName")
	}

	if len(emptyKeys) > 0 {
		return fmt.Errorf("database 配置存在空项: %v", emptyKeys)
	}

	return nil
}

// checkAsset 本方法用于检查配置文件中的 asset.base_url 配置项是否为合法的URL
func (c *Config) checkAsset() error {
	assetBaseURL, err := url.Parse(c.Asset.BaseURL)
	if err != nil {
		return fmt.Errorf("asset.base_url 解析失败: %w", err)
	}

	if assetBaseURL.Scheme == "" || assetBaseURL.Host == "" {
		return fmt.Errorf("asset.base_url 必须为含协议和主机的完整URL, 当前值: %q", c.Asset.BaseURL)
	}

	return nil
}

// checkStorage 本方法用于检查配置文件中的 storage.root_dir 配置项是否为绝对路径
func (c *Config) checkStorage() error {
	isAbs := filepath.IsAbs(c.Storage.RootDir)
	if !isAbs {
		return fmt.Errorf("storage.root_dir 必须为绝对路径, 当前值: %q", c.Storage.RootDir)
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
