# ppp
微信支付，支付宝支付服务商后端golang版本
支持多个服务商，多个app/服务商同时收款，具体看配置文件

## 启动
###证书配置
```
cert
    alipay
        private.key  // 应用私钥
        public.key   // 支付宝公钥
    wxpay
        cert.p12     //微信的p12证书
```

 - cp config.tmp.yml config.yml
 - go run rpc/hprose.go --path=configfilepath

### hprose
使用hprose的rpc模式启动
hprose 项目地址 github.com/hprose/hprose-golang

### Config
参考config.tmp.yml

## 支付宝支付
### 单商户模式/服务商模式
### 扫码支付
### 网站支付
### 手机网站支付
### app支付

## 微信支付
### 企业付款到零钱包
### 单服务商模式/服务商模式
### 扫码支付（主扫）
### 条码支付（被扫）
### 公众号支付
### app支付
### h5支付
