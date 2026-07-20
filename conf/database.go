package conf

// Database 数据库连接相关配置
type Database struct {
	Domain   string `yaml:"domain"`   // Domain 服务端地址
	Port     string `yaml:"port"`     // Port 端口
	User     string `yaml:"user"`     // User 用户名
	Password string `yaml:"password"` // Password 密码
	Name     string `yaml:"name"`     // Name 库名
}
