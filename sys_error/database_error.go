package sys_error

// DatabaseError 本错误用于包装数据库故障,便于调用方识别错误类别
type DatabaseError struct {
	Cause error // Cause 保留底层数据库错误,供受控排查与错误链检查,不直接发送给前端
}

// Error 本方法用于返回统一的数据库错误消息,不拼接底层 SQL 或连接细节
func (e *DatabaseError) Error() string {
	return "数据库操作失败,请稍后重试"
}

// Unwrap 本方法用于暴露底层错误,支持 errors.Is 与 errors.As 沿错误链检查原因
func (e *DatabaseError) Unwrap() error {
	return e.Cause
}
