package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
)

// InvalidHashErr 本错误用于表示无效的SHA-256十六进制摘要
var InvalidHashErr = errors.New("invalid SHA-256 hexadecimal digest")

// SHA256Hex 本方法用于计算给定字符串的SHA-256十六进制摘要,返回64个小写16进制字符
func SHA256Hex(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}

// MatchSHA256 本方法用于比较给定的字符串与预期的SHA-256十六进制摘要是否匹配,若预期摘要无效则返回 InvalidHashErr 错误
func MatchSHA256(value, expectedHex string) (bool, error) {
	expected, err := hex.DecodeString(expectedHex)
	if err != nil || len(expected) != sha256.Size {
		return false, InvalidHashErr
	}

	actual := sha256.Sum256([]byte(value))
	return subtle.ConstantTimeCompare(actual[:], expected) == 1, nil
}
