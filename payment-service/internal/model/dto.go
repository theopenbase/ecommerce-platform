package model

// 请求
type PayRequest struct {
	OrderNo   string  `json:"order_no" binding:"required"`
	Channel   string  `json:"channel" binding:"required,oneof=alipay wechat bank balance"`
	Amount    float64 `json:"amount" binding:"required,gt=0"`
}

type RefundRequest struct {
	TransNo string  `json:"trans_no" binding:"required"`
	Amount  float64 `json:"amount" binding:"required,gt=0"`
	Reason  string  `json:"reason"`
}

// 响应
type PayResponse struct {
	TransNo        string `json:"trans_no"`
	Channel        string `json:"channel"`
	PayURL         string `json:"pay_url,omitempty"`   // H5/扫码支付 URL
	QRCode         string `json:"qr_code,omitempty"`   // 扫码支付二维码内容
	ExpiresIn      int    `json:"expires_in"`          // 有效期（秒）
}

type RefundResponse struct {
	RefundNo string  `json:"refund_no"`
	Amount   float64 `json:"amount"`
	Status   uint8   `json:"status"`
}
