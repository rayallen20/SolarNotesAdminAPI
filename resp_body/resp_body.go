package resp_body

import "SolarNotesAdminAPI/sys_error"

// 响应码规则:
// 200: 成功响应
// 500: 服务端内部错误
//
// 10XXX: 系统级别错误
//
//	101XX: 参数校验错误
//	102XX: 数据库错误
//
// 20XXX: 具体业务错误
const (
	SuccessCode              = 200 // SuccessCode 成功响应状态码
	ServiceInternalErrorCode = 500 // ServiceInternalErrorCode 服务端内部错误状态码

	ValidateErrorCode = 10101 // ValidateErrorCode 参数校验错误状态码
	DBErrorCode       = 10201 // DBErrorCode 数据库错误状态码

	LoginCredentialErrorCode = 20001 // LoginCredentialErrorCode 账号或密码错误状态码
	LoginStateChangedCode    = 20002 // LoginStateChangedCode 登录期间凭据状态变化,需重新登录状态码
	TokenExpiredCode         = 20003 // TokenExpiredCode token过期状态码
	InvalidTokenCode         = 20004 // InvalidTokenCode token无效状态码
	LoginRateLimitCode       = 20005 // LoginRateLimitCode 请求登录接口触发访问频率限制状态码

	PlanetExistCode = 20101 // PlanetExistCode 请求创建的天体名称已经存在状态码

	BookPlanetNotFoundCode = 20102 // BookPlanetNotFoundCode 天体引用失效状态码
	BookNotFoundCode       = 20202 // BookNotFoundCode 表示书籍已不存在, 前端可重新获取书籍选项
	BookEditChangedCode    = 20204 // BookEditChangedCode 表示书籍的编辑快照已失效.
)

type RespBody struct {
	Code    int            `json:"code"`    // Code 响应体中的状态码部分
	Message string         `json:"message"` // Message 响应体中的消息部分
	Data    map[string]any `json:"data"`    // Data 响应体中的数据部分 即有效载荷
}

// Success 本方法用于设置响应成功响应体
func (r *RespBody) Success(data map[string]any) {
	r.Code = SuccessCode
	r.Message = "success"
	r.Data = data
}

// failure 本方法用于统一设置错误业务码/消息/空对象 Data ,避免输出 null 或遗留成功数据
func (r *RespBody) failure(code int, message string) {
	r.Code = code
	r.Message = message
	r.Data = map[string]any{}
}

// ValidateError 本方法用于设置参数校验错误时的响应体
func (r *RespBody) ValidateError(err error) {
	message := "请求参数不合法"
	if err != nil {
		message = err.Error()
	}
	r.failure(ValidateErrorCode, message)
}

// DBError 本方法用于设置数据库错误时的响应体
func (r *RespBody) DBError(err error) {
	r.failure(DBErrorCode, err.Error())
	// TODO: 开发完成后将DB错误换成这句
	// TODO: 形参名也要换成 _
	// r.failure(DBErrorCode, "数据库操作失败,请稍后重试")
}

// InternalError 本方法用于生成统一内部错误响应体,不向前端输出底层错误
func (r *RespBody) InternalError(err error) {
	r.failure(ServiceInternalErrorCode, err.Error())

	// TODO: 开发完成后将DB错误换成这句
	// TODO: 形参名也要换成 _
	// r.failure(ServiceInternalErrorCode, "服务暂时不可用,请稍后重试")
}

// LoginRateLimitError 本方法用于设置请求登录接口触发访问频率限制时的响应体
func (r *RespBody) LoginRateLimitError() {
	r.failure(LoginRateLimitCode, "请求过于频繁,请稍后再试")
}

// LoginCredentialError 本方法用于设置账号不存在/密码不正确时的响应体,不向前端输出具体错误
func (r *RespBody) LoginCredentialError() {
	r.failure(LoginCredentialErrorCode, "账号或密码错误")
}

// LoginStateChangedError 本方法用于设置验证用户名密码后,管理员记录的条件更新未命中时的响应体,不向前端输出具体错误
func (r *RespBody) LoginStateChangedError() {
	r.failure(LoginStateChangedCode, "账号状态已变化,请重新登录")
}

// TokenExpiredError 本方法用于设置token过期时的响应体
func (r *RespBody) TokenExpiredError() {
	r.failure(TokenExpiredCode, "登录已过期,请重新登录")
}

// InvalidToken 本方法用于生成令牌无效或已撤销响应时的响应体
func (r *RespBody) InvalidToken() {
	r.failure(InvalidTokenCode, "登录状态无效,请重新登录")
}

// PlanetExistError 本方法用于生成天体名称重复时的响应体
func (r *RespBody) PlanetExistError() {
	r.failure(PlanetExistCode, "请求创建的天体已存在,无法创建")
}

// BookPlanetNotFoundError 本方法用于生成天体引用失效时的响应体
func (r *RespBody) BookPlanetNotFoundError() {
	r.failure(BookPlanetNotFoundCode, sys_error.BookPlanetNotFoundErr.Error())
}

// BookNotFoundError 将书籍不存在映射为统一业务响应.
func (r *RespBody) BookNotFoundError() {
	r.failure(BookNotFoundCode, sys_error.BookNotFoundErr.Error())
}

// BookEditChangedError 返回书籍编辑冲突以及空的 data 对象.
func (r *RespBody) BookEditChangedError() {
	r.failure(BookEditChangedCode, sys_error.BookEditChangedErr.Error())
}

const (
	// CatalogueNotFoundCode 表示父目录已不存在.
	CatalogueNotFoundCode = 20301
	// CatalogueOrderChangedCode 表示本层排序快照已失效.
	CatalogueOrderChangedCode = 20302
	// CatalogueSortRequiredCode 表示用户请求显示待排序的节点.
	CatalogueSortRequiredCode = 20303
)

// CatalogueNotFoundError 返回父目录不存在的统一响应.
func (r *RespBody) CatalogueNotFoundError() {
	r.failure(CatalogueNotFoundCode, sys_error.CatalogueNotFoundErr.Error())
}

// CatalogueOrderChangedError 提示前端刷新后重新核对目标顺序.
func (r *RespBody) CatalogueOrderChangedError() {
	r.failure(CatalogueOrderChangedCode, sys_error.CatalogueOrderChangedErr.Error())
}

// CatalogueSortRequiredError 提示管理员先保存该节点所在层的阅读顺序.
func (r *RespBody) CatalogueSortRequiredError() {
	r.failure(CatalogueSortRequiredCode, sys_error.CatalogueSortRequiredErr.Error())
}

// ArticlePreviewUnavailableCode 表示已存在的节点暂时无法预览.
const ArticlePreviewUnavailableCode = 20401

// ArticlePreviewUnavailableError 返回稳定的业务码和消息.
func (r *RespBody) ArticlePreviewUnavailableError() {
	r.failure(ArticlePreviewUnavailableCode, sys_error.ArticlePreviewUnavailableErr.Error())
}
