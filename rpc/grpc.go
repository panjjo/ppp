package main

import (
	"context"
	"github.com/panjjo/ppp"
	proto "github.com/panjjo/ppp/proto"
	"github.com/sirupsen/logrus"
	"gogs.yunss.com/go/k8s"
	"google.golang.org/grpc"
	"net"
)

func startGRPCServer() {
	lis, err := net.Listen("tcp", config.Sys.ADDR)
	if err != nil {
		logrus.Fatal("start grpc server fail:", err)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(k8s.UnaryTraceInterceptor))
	proto.RegisterPPPServer(s, &GRPCHandler{})
	logrus.Infoln("Start:" + config.Sys.ADDR)
	if err := s.Serve(lis); err != nil {
		logrus.Fatalln("grpc server error:", err)
	}
}

type GRPCHandler struct{}

func error2Status(err ppp.Error) *proto.Result {
	return &proto.Result{Code: err.Code, Msg: err.Msg}
}

func (grpc *GRPCHandler) AliPayParams(ctx context.Context, req *proto.Params) (*proto.ParamsResult, error) {
	res, err := alipay.PayParams(ppp.NewContextWithCfg(proto.ALIPAY, req.AppID), req)
	return &proto.ParamsResult{Data: res, Status: error2Status(err)}, nil
}

func (grpc *GRPCHandler) AliBarPay(ctx context.Context, req *proto.Barpay) (*proto.TradeResult, error) {
	res, err := alipay.BarPay(ppp.NewContextWithCfg(proto.ALIPAY, req.AppID), req)
	return &proto.TradeResult{Data: res, Status: error2Status(err)}, nil
}

func (grpc *GRPCHandler) AliRefund(ctx context.Context, req *proto.Refund) (*proto.RefundResult, error) {
	res, err := alipay.Refund(ppp.NewContextWithCfg(proto.ALIPAY, req.AppID), req)
	return &proto.RefundResult{Data: res, Status: error2Status(err)}, nil
}

func (grpc *GRPCHandler) AliCancel(ctx context.Context, req *proto.Trade) (*proto.TradeResult, error) {
	res, err := alipay.Cancel(ppp.NewContextWithCfg(proto.ALIPAY, req.AppID), req)
	return &proto.TradeResult{Data: res, Status: error2Status(err)}, nil
}

func (grpc *GRPCHandler) AliTradeInfo(ctx context.Context, req *proto.Trade) (*proto.TradeResult, error) {
	res, err := alipay.TradeInfo(ppp.NewContextWithCfg(proto.ALIPAY, req.AppID), req)
	return &proto.TradeResult{Data: res, Status: error2Status(err)}, nil
}

func (grpc *GRPCHandler) AliAuthSigned(ctx context.Context, req *proto.Account) (*proto.AccountResult, error) {
	res, err := alipay.AuthSigned(ppp.NewContextWithCfg(proto.ALIPAY, req.AppID), req)
	return &proto.AccountResult{Data: res, Status: error2Status(err)}, nil
}

func (grpc *GRPCHandler) AliAuth(ctx context.Context, req *proto.AuthReq) (*proto.AccountResult, error) {
	res, err := alipay.Auth(ppp.NewContextWithCfg(proto.ALIPAY, req.AppID), req.Code)
	return &proto.AccountResult{Data: res, Status: error2Status(err)}, nil
}

func (grpc *GRPCHandler) AliBindUser(ctx context.Context, req *proto.User) (*proto.UserResult, error) {
	res, err := alipay.BindUser(ppp.NewContextWithCfg(proto.ALIPAY, req.AppID), req)
	return &proto.UserResult{Data: res, Status: error2Status(err)}, nil
}

func (grpc *GRPCHandler) AliMchPay(ctx context.Context, req *proto.Transfer) (*proto.TransferResult, error) {
	res, err := alipay.MchPay(ppp.NewContextWithCfg(proto.ALIPAY, req.AppID), req)
	return &proto.TransferResult{Data: res, Status: error2Status(err)}, nil
}

//    // wxpay 微信服务商模式
func (grpc *GRPCHandler) WXPayParam(ctx context.Context, req *proto.Params) (*proto.ParamsResult, error) {
	res, err := wxpay.PayParams(ppp.NewContextWithCfg(proto.WXPAY, req.AppID), req)
	return &proto.ParamsResult{Data: res, Status: error2Status(err)}, nil
}
func (grpc *GRPCHandler) WXBarPay(ctx context.Context, req *proto.Barpay) (*proto.TradeResult, error) {
	res, err := wxpay.BarPay(ppp.NewContextWithCfg(proto.WXPAY, req.AppID), req)
	return &proto.TradeResult{Data: res, Status: error2Status(err)}, nil
}
func (grpc *GRPCHandler) WXRefund(ctx context.Context, req *proto.Refund) (*proto.RefundResult, error) {
	res, err := wxpay.Refund(ppp.NewContextWithCfg(proto.WXPAY, req.AppID), req)
	return &proto.RefundResult{Data: res, Status: error2Status(err)}, nil
}
func (grpc *GRPCHandler) WXTradeInfo(ctx context.Context, req *proto.Trade) (*proto.TradeResult, error) {
	res, err := wxpay.TradeInfo(ppp.NewContextWithCfg(proto.WXPAY, req.AppID), req)
	return &proto.TradeResult{Data: res, Status: error2Status(err)}, nil
}
func (grpc *GRPCHandler) WXCancel(ctx context.Context, req *proto.Trade) (*proto.TradeResult, error) {
	res, err := wxpay.Cancel(ppp.NewContextWithCfg(proto.WXPAY, req.AppID), req)
	return &proto.TradeResult{Data: res, Status: error2Status(err)}, nil
}

func (grpc *GRPCHandler) WXBindUser(ctx context.Context, req *proto.User) (*proto.UserResult, error) {
	res, err := wxpay.BindUser(ppp.NewContextWithCfg(proto.WXPAY, req.AppID), req)
	return &proto.UserResult{Data: res, Status: error2Status(err)}, nil
}
func (grpc *GRPCHandler) WXAuthSigned(ctx context.Context, req *proto.Account) (*proto.AccountResult, error) {
	res, err := wxpay.AuthSigned(ppp.NewContextWithCfg(proto.WXPAY, req.AppID), req)
	return &proto.AccountResult{Data: res, Status: error2Status(err)}, nil
}

func (grpc *GRPCHandler) WXSinglePayParams(ctx context.Context, req *proto.Params) (*proto.ParamsResult, error) {
	res, err := wxpaySingle.PayParams(ppp.NewContextWithCfg(proto.WXPAYSINGLE, req.AppID), req)
	return &proto.ParamsResult{Data: res, Status: error2Status(err)}, nil
}
func (grpc *GRPCHandler) WXSingleBarPay(ctx context.Context, req *proto.Barpay) (*proto.TradeResult, error) {
	res, err := wxpaySingle.BarPay(ppp.NewContextWithCfg(proto.WXPAYSINGLE, req.AppID), req)
	return &proto.TradeResult{Data: res, Status: error2Status(err)}, nil
}
func (grpc *GRPCHandler) WXSingleRefund(ctx context.Context, req *proto.Refund) (*proto.RefundResult, error) {
	res, err := wxpaySingle.Refund(ppp.NewContextWithCfg(proto.WXPAYSINGLE, req.AppID), req)
	return &proto.RefundResult{Data: res, Status: error2Status(err)}, nil
}
func (grpc *GRPCHandler) WXSingleTradeInfo(ctx context.Context, req *proto.Trade) (*proto.TradeResult, error) {
	res, err := wxpaySingle.TradeInfo(ppp.NewContextWithCfg(proto.WXPAYSINGLE, req.AppID), req)
	return &proto.TradeResult{Data: res, Status: error2Status(err)}, nil
}
func (grpc *GRPCHandler) WXSingleCancel(ctx context.Context, req *proto.Trade) (*proto.TradeResult, error) {
	res, err := wxpaySingle.Cancel(ppp.NewContextWithCfg(proto.WXPAYSINGLE, req.AppID), req)
	return &proto.TradeResult{Data: res, Status: error2Status(err)}, nil
}
func (grpc *GRPCHandler) WXSingleMchPay(ctx context.Context, req *proto.Transfer) (*proto.TransferResult, error) {
	res, err := wxpaySingle.MchPay(ppp.NewContextWithCfg(proto.WXPAYSINGLE, req.AppID), req)
	return &proto.TransferResult{Data: res, Status: error2Status(err)}, nil
}
