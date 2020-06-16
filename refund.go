package ppp

const (
	refundTable = "refunds"

	// RefundStatusSucc 退款成功
	RefundStatusSucc Status = 1
)

// Refund 退款交易结构
type Refund struct {
	RefundID    string
	ID          string
	OutRefundID string
	SourceID    string // 使用Trade.OutTradeID
	Amount      int64
	Status      Status
	MchID       string
	UserID      string
	UpTime      int64
	RefundTime  int64
	Create      int64
	AppID       string
	From        string // 单据来源 alipay wxpay
	Memo        string
}

func saveRefund(refund *Refund) error {
	return DBClient.Insert(refundTable, refund)
}
