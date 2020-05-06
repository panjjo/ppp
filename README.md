# ppp
----
## 说明
---
### 类型
工具类

### 主要功能

- 集成支付宝微信支付
- 支持扫码支付（主扫/被扫），app支付，小程序支付，公众号支付，网页支付，退款，企业付款
- 集成用户支付宝微信授权管理
- 支持服务商，单商户两种模式
- 支持多个服务商，多个app/服务商同时收款，具体看配置文件



----
## 启动
---
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
 - go run rpc/*.go --path=configfilepath

### GRPC

### Config
参考config.tmp.yml

---
## 调用流程
---
    所有请求都有一个参数叫tag 标识使用那一套商户号进行收款，传空标识使用默认商户进行收款，
    服务商，单商户都是一样的
### 支付宝
    支付宝服务商账号可直接用来收款，所以单商户收款配置就和服务商配置一样，
    可以直接使用服务商配置做单商户收款
+ 服务商模式
  - 网站授权获得app_auth_code [文档地址](https://docs.open.alipay.com/20160728150111277227/intro)
  - 使用app_auth_code 调用 Services.AliPayAuth 进行支付宝授权
  - Services.AliPayAuthSigned 验证授权是否签约（比如当面付等功能是否有权限）
  - 绑定用户（可选）不绑定用户的，支付，退款，查询等直接传入mchid即可
    + Services.AliPayBindUser 将userid与mchid绑定，一个mchid可绑定多个userid，一个userid只能绑定一个mchid
    + Services.AliPayUnBindUser userid与mchid解绑
  - 扫码支付 Services.AliPayBarPay 可传入userid 或 mchid 来确定收款用户
  - 退款 Services.AliPayRefund
  - 查询 Services.AliPayTradeInfo
+ 单商户模式
  - 直接调用对应接口，userid,mchid 都传入空，会对应操作tag所匹配到的商户数据
### 微信
    微信服务商身份不能收款，所以说单商户的需要单独配置，wxpay=服务商 wxpay_single单商户模式，
    微信单商户模式下，同一个商户号可以绑定多个appid进行收款，具体看配置文件配置方法说明
+ 服务商模式
  - Services.WXPayAuthSigned 创建微信账号并验证是否有权限
   - 绑定用户（可选）不绑定用户的，支付，退款，查询等直接传入mchid即可
    + Services.WXPayBindUser 将userid与mchid绑定，一个mchid可绑定多个userid，一个userid只能绑定一个mchid
    + Services.WXPayUnBindUser userid与mchid解绑
  - 扫码支付 Services.WXPayBarPay 可传入userid 或 mchid 来确定收款用户
  - 退款 Services.WXPayRefund
  - 查询 Services.WXPayTradeInfo
+ 单商户模式
  - 直接调用对应接口，userid,mchid 都传入空，会对应操作tag所匹配到的商户数据
  
---
## 支付宝支付
---
- 单商户模式/服务商模式
- 扫码支付
- 网站支付
- 手机网站支付
- app支付
- 企业付款

## 微信支付
- 企业付款到零钱包
- 单服务商模式/服务商模式
- 扫码支付（主扫）
- 条码支付（被扫）
- 公众号支付
- app支付
- h5支付
