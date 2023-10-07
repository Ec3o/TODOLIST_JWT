package main

// TODOError 此文件用于预定义错误类型
type TODOError struct {
	Message string `json:"error"`
}

type USERError struct {
	Message string `json:"error"`
}

type JWTError struct {
	Message string `json:"error"`
}

var (
	ErrInvalidTODOFormat = TODOError{"抱歉，您提供的TODO数据格式不正确"}
	ErrInvalidDeadline   = TODOError{"无效的截止时间"}
	ErrReadTODOData      = TODOError{"无法读取TODO数据"}
	ErrSaveTODOData      = TODOError{"无法保存TODO数据"}
	ErrTODOIndexNotExist = TODOError{"抱歉，您访问的ToDo目前不存在，请先创建"}
	ErrTODONotFound      = TODOError{"抱歉，您要删除的ToDo目前不存在，请先创建"}
	ErrInvalidUSERFormat = USERError{"抱歉，您提供的用户数据格式不正确"}
	ErrInvalidPassword   = USERError{"密码不能为空或密码长度过短"}
	ErrReadUserData      = USERError{"无法读取用户数据"}
	ErrSaveUserData      = USERError{"无法保存保存数据"}
	ErrUserlogin         = USERError{"用户未注册或密码错误"}
	ErrNoToken           = JWTError{"未提供令牌"}
	ErrInvalidToken      = JWTError{"无效令牌"}
	ErrGenerateToken     = JWTError{"令牌生成错误"}
)
