package conf

import (
	"fmt"
	"net/url"
)

// Asset 资产相关配置
type Asset struct {
	BaseURL string `yaml:"base_url"` // BaseURL 资产URL基地址
}

// validate 本方法用于检查配置文件中的 asset 配置段中,base_url是否为合法URL
func (a *Asset) validate() error {
	baseURL, err := url.Parse(a.BaseURL)
	if err != nil {
		return fmt.Errorf("asset.base_url 解析失败: %w", err)
	}

	if baseURL.Scheme == "" || baseURL.Host == "" {
		return fmt.Errorf("asset.base_url 必须为包含协议和主机的完整URL, 当前值: %q", a.BaseURL)
	}

	return nil
}