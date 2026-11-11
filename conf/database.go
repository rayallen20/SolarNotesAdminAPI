package conf

import "fmt"

// Database 数据库连接相关配置
type Database struct {
	Host     string `yaml:"host"`     // Host 服务端地址
	Port     string `yaml:"port"`     // Port 端口
	User     string `yaml:"user"`     // User 用户名
	Password string `yaml:"password"` // Password 密码
	Name     string `yaml:"name"`     // Name 库名
}

// validate 本方法用于检查配置文件的 database 配置段是否存在空项
func (d *Database) validate() error {
	var emptyKeys []string

	if d.Host == "" {
		emptyKeys = append(emptyKeys, "host")
	}

	if d.Port == "" {
		emptyKeys = append(emptyKeys, "port")
	}

	if d.User == "" {
		emptyKeys = append(emptyKeys, "user")
	}

	if d.Password == "" {
		emptyKeys = append(emptyKeys, "password")
	}

	if d.Name == "" {
		emptyKeys = append(emptyKeys, "name")
	}

	if len(emptyKeys) > 0 {
		return fmt.Errorf("database 配置段存在空项: %v", emptyKeys)
	}

	return nil
}