# ppp

## Cert
证书配置
```
cert
    alipay
        private.key  // 应用私钥
        public.key   // 支付宝公钥
    wxpay
        cert.p12     //微信的p12证书
```

## 启动
```
    go run rpc/main.go
```
