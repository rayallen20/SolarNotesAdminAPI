package auth

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTOptions 本结构体用于表示JWT解析时的所需的可信参数
type JWTOptions struct {
	Issuer     string        // Issuer 期望的iss,签发时写入,解析时要求完全匹配
	Audience   string        // Audience 期望的aud,签发时写入,解析时要求完全匹配
	Lifetime   time.Duration // Lifetime 签发有效期,要求至少1s且为整数秒
	SigningKey []byte        // SigningKey 至少32字节的原始密钥
}

// validate 本方法用于验证JWTOptions的合法性,若合法则返回nil,若不合法则返回错误
func (o *JWTOptions) validate() error {
	if strings.TrimSpace(o.Issuer) == "" || strings.TrimSpace(o.Audience) == "" {
		return errors.New("JWT issuer and audience are required")
	}

	if o.Lifetime < time.Second || o.Lifetime%time.Second != 0 {
		return errors.New("JWT lifetime must be positive whole seconds")
	}

	if len(o.SigningKey) < 32 {
		return errors.New("JWT signing key must contain at least 32 bytes")
	}

	return nil
}

// verificationKey 本方法用于在JWT解析时提供签名密钥,若签名算法不为HS256则返回InvalidTokenErr错误
func (o *JWTOptions) verificationKey(token *jwt.Token) (any, error) {
	if token.Method != jwt.SigningMethodHS256 {
		return nil, InvalidTokenErr
	}

	return o.SigningKey, nil
}
