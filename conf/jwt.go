package conf

import (
	"encoding/base64"
	"errors"
	"strings"
	"time"
)

type JWT struct {
	Issuer        string `yaml:"issuer"`         // Issuer JWT的ISS,标识允许签发令牌的后端
	Audience      string `yaml:"audience"`       // Audience JWT的AUD,标识令牌预期服务的前端
	ExpireSeconds int64  `yaml:"expire_seconds"` // ExpireSeconds JWT的过期时间,单位秒
	Secret        string `yaml:"secret"`         // Secret JWT的签名密钥
	signingKey    []byte // signingKey 从 Secret 解码出的原始字节(私有字段不参与YAML绑定)
}

// validate 本方法用于验证JWT配置的合法性,若合法则将秘钥保存至 signingKey 字段中,若不合法则返回错误并阻止配置加载
func (j *JWT) validate() error {
	j.Issuer = strings.TrimSpace(j.Issuer)
	j.Audience = strings.TrimSpace(j.Audience)

	if j.Issuer == "" || j.Audience == "" {
		return errors.New("JWT配置错误: Issuer和Audience不能为空字符串")
	}

	const maxExpireSeconds int64 = (1<<63 - 1) / int64(time.Second)
	if j.ExpireSeconds <= 0 || j.ExpireSeconds > maxExpireSeconds {
		return errors.New("JWT配置错误: expire_seconds 必须是有效的正秒数")
	}

	encoded := strings.TrimSpace(j.Secret)
	if encoded == "" {
		return errors.New("JWT配置错误: Secret不能为空字符串")
	}

	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return errors.New("JWT配置错误: Secret必须是Base64编码的字符串")
	}

	if len(key) < 32 {
		return errors.New("JWT配置错误: Secret解码后的字节长度必须至少为32字节")
	}

	// 复制字节切片以避免外部修改
	j.signingKey = append([]byte(nil), key...)

	return nil
}

// SigningKey 本方法用于获取JWT的签名密钥,为防止调用方修改签名密钥,返回的字节切片为签名密钥的副本
func (j *JWT) SigningKey() []byte {
	return append([]byte(nil), j.signingKey...)
}

// Lifetime 本方法用于获取JWT的有效期,单位为time.Duration
func (j *JWT) Lifetime() time.Duration {
	return time.Duration(j.ExpireSeconds) * time.Second
}
