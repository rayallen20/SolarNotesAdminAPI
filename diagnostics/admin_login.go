package diagnostics

import (
	"SolarNotesAdminAPI/auth"
	"SolarNotesAdminAPI/sys_error"
	"context"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/go-sql-driver/mysql"
)

// LogAdminLoginFailure 本函数用于记录管理员登录失败的日志,并根据错误类型分类记录
func LogAdminLoginFailure(err error) {
	var dbErr *sys_error.DatabaseError
	var internalErr *sys_error.InternalError

	switch {
	case errors.As(err, &dbErr):
		log.Printf("管理员登录失败: category = database %s", describeLoginCause(dbErr.Cause))
	case errors.As(err, &internalErr):
		log.Printf("管理员登录失败: category = internal %s", describeLoginCause(internalErr.Cause))
	default:
		log.Printf("管理员登录失败: category = unknown %s", describeLoginCause(err))
	}
}

// describeLoginCause 本函数用于根据错误类型描述管理员登录失败的原因,以便在日志中记录
func describeLoginCause(cause error) string {
	switch {
	case cause == nil:
		return "reason = none"
	case errors.Is(cause, context.Canceled):
		return "reason = context canceled"
	case errors.Is(cause, context.DeadlineExceeded):
		return "reason = context deadline exceeded"
	case errors.Is(cause, auth.InvalidHashErr):
		return "reason = invalid password hash"
	}

	var mysqlErr *mysql.MySQLError
	if errors.As(cause, &mysqlErr) && mysqlErr != nil {
		return fmt.Sprintf("reason = mysql, mysql error number = %d", mysqlErr.Number)

		// Tips: 若需要查看更详细信息,请使用以下代码
		//return fmt.Sprintf("reason = mysql, mysql error number = %d, mysql error message = %s", mysqlErr.Number, mysqlErr.Message)
	}

	var networkErr *net.OpError
	if errors.As(cause, &networkErr) && networkErr != nil {
		return fmt.Sprintf("reason = network")

		// Tips: 若需要查看更详细信息,请使用以下代码
		// return fmt.Sprintf("reason = network, network operation = %s, network error = %s", networkErr.Op, networkErr.Err)
	}

	return fmt.Sprintf("reason = unclassified error, error type = %T", cause)

	// Tips: 若需要查看更详细信息,请使用以下代码
	// return fmt.Sprintf("reason = unclassified error, error type = %T, error message = %s", cause, cause.Error())
}
