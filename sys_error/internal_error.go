package sys_error

// InternalError 本错误用于包装配置/摘要格式/token生成等内部处理故障
type InternalError struct {
	Cause error // Cause 保存内部失败原因,对外消息由 Error 统一提供
}

// Error 本方法用于返回统一内部错误消息,避免向外暴露实现细节
func (e *InternalError) Error() string {
	return "服务暂时不可用,请稍后重试"
}

// Unwrap 本方法用于返回原始内部错误,供受控诊断或错误链检查
func (e *InternalError) Unwrap() error {
	return e.Cause
}
