package respond

// Error constants
var (
	ErrParameterError = ApiError(404, "Parameter error.")
	ErrServiceError   = ApiError(404, "service exception.")
	ErrNoDataFound    = ApiNullData(100, "no data found.")
	ErrNoPinFound     = ApiError(100, "no pin found.")
	ErrNoChildFound   = ApiError(100, "no child found.")
	ErrNoNodeFound    = ApiError(100, "no node found.")
	ErrNoResultFound  = ApiError(100, "no result found.")
	ErrAddressIsEmpty = ApiError(100, "address is empty.")
)

type ApiResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"message"`
	Data interface{} `json:"data"`
}

func ApiError(code int, msg string) (res *ApiResponse) {
	return &ApiResponse{Code: code, Msg: msg}
}
func ApiNullData(code int, msg string) (res *ApiResponse) {
	return &ApiResponse{Code: code, Msg: msg, Data: []string{}}
}
func ApiSuccess(code int, msg string, data interface{}) (res *ApiResponse) {
	return &ApiResponse{Code: code, Msg: msg, Data: data}
}
