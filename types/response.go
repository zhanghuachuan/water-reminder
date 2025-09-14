package types

// 统一API响应格式
type ApiResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ApiError 用于算子返回的错误信息
type ApiError struct {
	Message    string `json:"message"`
	ErrorMsg   string `json:"errorMsg"`
	StatusCode int    `json:"statusCode"`
}

// Error 实现error接口
func (e *ApiError) Error() string {
	return e.Message + ": " + e.ErrorMsg
}

// 创建API错误
func NewApiError(message, errorMsg string, statusCode int) *ApiError {
	return &ApiError{
		Message:    message,
		ErrorMsg:   errorMsg,
		StatusCode: statusCode,
	}
}

// 登录成功响应数据
type LoginResponseData struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

// 创建成功响应
func NewSuccessResponse(message string, data interface{}) *ApiResponse {
	return &ApiResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// 创建错误响应
func NewErrorResponse(message string, errorMsg string) *ApiResponse {
	return &ApiResponse{
		Success: false,
		Message: message,
		Error:   errorMsg,
	}
}
