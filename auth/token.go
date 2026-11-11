package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// InvalidTokenErr 本错误用于表示无效的token
	InvalidTokenErr = errors.New("invalid administrator token")
	// ExpiredTokenErr 本错误用于表示token已过期
	ExpiredTokenErr = errors.New("administrator token expired")
)

// maxAdminTokenLength 本常量用于定义token的最大长度,超过该长度的token将被视为无效
const maxAdminTokenLength = 4096

// IssueAdminToken 本方法用于为给定的管理员ID签发一个新的JWT令牌,并返回该令牌字符串
func IssueAdminToken(adminID int64, options *JWTOptions) (string, error) {
	err := options.validate()
	if err != nil {
		return "", err
	}

	if adminID <= 0 {
		return "", errors.New("administrator ID must be positive")
	}

	randomID := make([]byte, 16)
	_, err = rand.Read(randomID)
	if err != nil {
		return "", err
	}

	now := time.Now().UTC().Truncate(time.Second)
	claims := &AdminClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:  options.Issuer,
			Subject: strconv.FormatInt(adminID, 10),
			Audience: jwt.ClaimStrings{
				options.Audience,
			},
			ExpiresAt: jwt.NewNumericDate(now.Add(options.Lifetime)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        hex.EncodeToString(randomID),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	raw, err := token.SignedString(options.SigningKey)
	if err != nil {
		return "", err
	}

	if len(raw) > maxAdminTokenLength {
		return "", errors.New("JWT exceeds the supported length")
	}

	return raw, nil
}

// ParseAdminToken 本函数用于解析给定的JWT token字符串,并返回其中包含的管理员ID,若token无效或已过期则返回相应错误
func ParseAdminToken(raw string, options *JWTOptions) (int64, error) {
	err := options.validate()
	if err != nil {
		return 0, err
	}

	if raw == "" || len(raw) > maxAdminTokenLength {
		return 0, InvalidTokenErr
	}

	claims := &AdminClaims{}
	token, err := jwt.ParseWithClaims(
		raw,
		claims,
		options.verificationKey,
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(options.Issuer),
		jwt.WithAudience(options.Audience),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)
	if err != nil {
		if isOnlyExpiredError(err) {
			return 0, ExpiredTokenErr
		}

		return 0, InvalidTokenErr
	}

	if token == nil || !token.Valid {
		return 0, InvalidTokenErr
	}

	id, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || id <= 0 {
		return 0, InvalidTokenErr
	}

	return id, nil
}

// isOnlyExpiredError 本函数用于判断给定的错误是否仅包含 ExpiredTokenErr ,若是则返回true,否则返回false
func isOnlyExpiredError(err error) bool {
	hasExpired, hasOther := classifyTokenError(err)
	return hasExpired && !hasOther
}

// classifyTokenError 本函数用于递归展开错误链,分别报告是否出现 ExpiredTokenErr 和其他具体错误
// hasExpired 表示是否出现 ExpiredTokenErr , hasOther 表示是否出现其他具体错误
func classifyTokenError(err error) (hasExpired, hasOther bool) {
	if err == nil {
		return false, false
	}

	switch wrapped := err.(type) {
	case interface{ Unwrap() []error }:
		for _, child := range wrapped.Unwrap() {
			childExpired, childOther := classifyTokenError(child)
			hasExpired = hasExpired || childExpired
			hasOther = hasOther || childOther
		}

		return hasExpired, hasOther
	case interface{ Unwrap() error }:
		return classifyTokenError(wrapped.Unwrap())
	default:
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return true, false
		case errors.Is(err, jwt.ErrTokenInvalidClaims):
			return false, false
		default:
			return false, true
		}
	}
}
