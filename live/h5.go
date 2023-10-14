package live

import "net/http"

const (
	H5QueryTimestamp = "Timestamp"
	H5QueryCode      = "Code"
	H5QueryMid       = "Mid"
	H5QueryCaller    = "Caller"
	H5QueryCodeSign  = "CodeSign"

	H5QueryRoomID  = "RoomId"
	H5QueryPlugEnv = "plug_env"
)

type H5SignatureParams struct {
	Timestamp string
	Code      string
	Mid       string
	Caller    string
	CodeSign  string

	RoomID  string
	PlugEnv string
}

// ToSortMap 参与签名的字段转map<string, string>
func (h5sp H5SignatureParams) ToSortMap() map[string]string {
	return map[string]string{
		H5QueryTimestamp: h5sp.Timestamp,
		H5QueryCode:      h5sp.Code,
		H5QueryMid:       h5sp.Mid,
		H5QueryCaller:    h5sp.Caller,
	}
}

// CreateSignature 生成签名
func (h5sp H5SignatureParams) CreateSignature(accessKeySecret string) string {
	signStr := ToSortedString(h5sp.ToSortMap())
	return HmacSHA256(accessKeySecret, signStr)
}

// ValidateSignature 验证签名
func (h5sp H5SignatureParams) ValidateSignature(accessKeySecret string) bool {
	return h5sp.CodeSign == h5sp.CreateSignature(accessKeySecret)
}

// ParseH5SignatureParamsWithRequest 从http.Request中解析出签名参数
func ParseH5SignatureParamsWithRequest(req *http.Request) *H5SignatureParams {
	return &H5SignatureParams{
		Timestamp: req.URL.Query().Get(H5QueryTimestamp),
		Code:      req.URL.Query().Get(H5QueryCode),
		Mid:       req.URL.Query().Get(H5QueryMid),
		Caller:    req.URL.Query().Get(H5QueryCaller),
		CodeSign:  req.URL.Query().Get(H5QueryCodeSign),

		RoomID:  req.URL.Query().Get(H5QueryRoomID),  // 调用页会携带
		PlugEnv: req.URL.Query().Get(H5QueryPlugEnv), // 调用页会携带
	}
}
