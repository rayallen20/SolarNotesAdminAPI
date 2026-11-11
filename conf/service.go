package conf

import (
	"errors"
	"time"
)

// Service 服务相关配置
type Service struct {
	Port                     string   `yaml:"port"`                        // Port 监听端口
	TrustedProxies           []string `yaml:"trusted_proxies"`             // TrustedProxies 可信代理 IP 或 CIDR 列表,空列表表示不信任代理
	ReadHeaderTimeoutSeconds int64    `yaml:"read_header_timeout_seconds"` // ReadHeaderTimeoutSeconds 读取请求头的最长时间.单位:秒
	ReadTimeoutSeconds       int64    `yaml:"read_timeout_seconds"`        // ReadTimeoutSeconds 读取整个请求的最长时间.单位:秒
	IdleTimeoutSeconds       int64    `yaml:"idle_timeout_seconds"`        // IdleTimeoutSeconds 等待下一个请求的最长空闲时间.单位:秒
	MaxRequestBodyBytes      int64    `yaml:"max_request_body_bytes"`      // MaxRequestBodyBytes API 请求体大小上限,单位:字节
}

// maxServiceTimeoutSeconds 本常量用于表示转换为 time.Duration 时,不会溢出的最大秒数
const maxServiceTimeoutSeconds int64 = (1<<63 - 1) / int64(time.Second)

// validate 本方法校验 HTTP 服务超时配置
func (s *Service) validate() error {
	if s.ReadHeaderTimeoutSeconds <= 0 || s.ReadHeaderTimeoutSeconds > maxServiceTimeoutSeconds {
		return errors.New("service.read_header_timeout_seconds 必须是有效的正秒数")
	}

	if s.ReadTimeoutSeconds <= 0 || s.ReadTimeoutSeconds > maxServiceTimeoutSeconds {
		return errors.New("service.read_timeout_seconds 必须是有效的正秒数")
	}

	if s.IdleTimeoutSeconds <= 0 || s.IdleTimeoutSeconds > maxServiceTimeoutSeconds {
		return errors.New("service.idle_timeout_seconds 必须是有效的正秒数")
	}

	if s.ReadTimeoutSeconds < s.ReadHeaderTimeoutSeconds {
		return errors.New("service.read_timeout_seconds 不得小于 read_header_timeout_seconds")
	}

	if s.MaxRequestBodyBytes <= 0 {
		return errors.New("service.max_request_body_bytes 必须大于 0")
	}
	return nil
}

// ReadHeaderTimeout 本方法用于返回读取请求头的最长时间
func (s *Service) ReadHeaderTimeout() time.Duration {
	return time.Duration(s.ReadHeaderTimeoutSeconds) * time.Second
}

// ReadTimeout 本方法用于返回读取整个请求的最长时间
func (s *Service) ReadTimeout() time.Duration {
	return time.Duration(s.ReadTimeoutSeconds) * time.Second
}

// IdleTimeout 本方法用于返回等待下一个请求的最长空闲时间
func (s *Service) IdleTimeout() time.Duration {
	return time.Duration(s.IdleTimeoutSeconds) * time.Second
}
