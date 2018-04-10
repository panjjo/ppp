package ppp

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/pkcs12"
)

var aliPayPrivateKey *rsa.PrivateKey //应用私钥
var aliPayPublicKey *rsa.PublicKey   //支付宝公钥
var wxPaySecretKey string            //微信支付
var wxPayCertTlsConfig *tls.Config   //微信支付证书tls

//WXPay使用私钥做验签
//用于同步接口请求
func WXPaySigner(data map[string]string) (signer string) {
	message := mapSortAndJoin(data, "=", "&", true)
	message += "&key=" + wxPaySecretKey
	return strings.ToUpper(makeMd5(message))
}

//AliPay使用私钥做验签
//用于同步接口请求
func AliPaySigner(data map[string]string) (signer []byte) {
	message := mapSortAndJoin(data, "=", "&", false)
	rng := rand.Reader
	hashed := sha256.Sum256([]byte(message))
	signer, _ = rsa.SignPKCS1v15(rng, aliPayPrivateKey, crypto.SHA256, hashed[:])
	return
}

//AliPay使用支付宝公钥做验证
//异步callback
func AliPayRSAVerify(message map[string]string, sign string) (err error) {
	digest := sha256.Sum256([]byte(mapSortAndJoin(message, "=", "&", false)))
	data, _ := base64.StdEncoding.DecodeString(sign)
	err = rsa.VerifyPKCS1v15(aliPayPublicKey, crypto.SHA256, digest[:], data)
	return
}

//加载支付宝相关证书信息
func loadAliPayCertKey(path string) {
	private, err := ioutil.ReadFile(filepath.Join(path, "cert/alipay/private.key"))
	if err != nil {
		log.Fatal("Load AliPay PrivateKey Error:", err)
	}
	aliPayPrivateKey, err = x509.ParsePKCS1PrivateKey(base64Decode(string(private)))
	if err != nil {
		log.Fatal("Load AliPay PrivateKey Error:", err)
	}
	public, err := ioutil.ReadFile(filepath.Join(path, "cert/alipay/public.key"))
	if err != nil {
		log.Fatal("Load AliPay PublicKey Error:", err)
	}
	block, _ := pem.Decode(public)
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		log.Fatal("Load AliPay Parse Error:", err)
	}
	aliPayPublicKey, _ = pub.(*rsa.PublicKey)
}

//加载微信支付相关证书信息
func loadWXPayCertKey(path string) {
	b, err := ioutil.ReadFile(filepath.Join(path, "cert/wxpay/cert.p12"))
	if err != nil {
		log.Fatal("Load WXPay Cert Error:", err)
	}

	blocks, err := pkcs12.ToPEM(b, wxPayMchId)
	if err != nil {
		log.Fatal("Load WXPay Cert Error:", err)
	}

	var pemData []byte
	for _, b := range blocks {
		pemData = append(pemData, pem.EncodeToMemory(b)...)
	}

	cert, err := tls.X509KeyPair(pemData, pemData)
	if err != nil {
		log.Fatal("Load WXPay Cert Error:", err)
	}

	wxPayCertTlsConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

}
