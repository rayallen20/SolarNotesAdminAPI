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

// MatchSHA256 对原值计算摘要, 再与库中的十六进制摘要比较.
func MatchSHA256(value, expectedHex string) (bool, error) {
	return MatchSHA256Digest(sha256.Sum256([]byte(value)), expectedHex)
}

// MatchSHA256Digest 比较已经算好的摘要与库中摘要, 支持大小写十六进制.
func MatchSHA256Digest(actual [sha256.Size]byte, expectedHex string) (bool, error) {
	expected, err := hex.DecodeString(expectedHex)
	if err != nil || len(expected) != sha256.Size {
		return false, InvalidHashErr
	}
	return subtle.ConstantTimeCompare(actual[:], expected) == 1, nil
}
