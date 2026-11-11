package sys_error

import "errors"

var (
	// InvalidLoginRequestErr 本错误表示请求对象缺失或基本登录字段为空
	InvalidLoginRequestErr = errors.New("请求参数不合法")
	// InvalidCredentialsErr 本错误表示账号不存在/密码不匹配
	InvalidCredentialsErr = errors.New("账号或密码错误")
	// LoginStateChangedErr 本错误表示密码验证后,写入令牌的条件更新未命中
	LoginStateChangedErr = errors.New("账号状态已变化,请重新登录")
)
