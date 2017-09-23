package ppp

//对账单主体
type Statement struct {
}

func (S *Statement) List(request *ListRequest, resp *TradeListResult) error {
	trades, err := listTrade(request)
	if err != nil {
		resp.Code = TradeQueryErr
		resp.SourceData = err.Error()
	} else {
		resp.Data = trades
	}
	return nil
}
func (S *Statement) Count(request *ListRequest, resp *CountResult) error {
	n, err := countTrade(request)
	if err != nil {
		resp.Code = TradeQueryErr
		resp.SourceData = err.Error()
	} else {
		resp.Data = n
	}
	return nil
}
