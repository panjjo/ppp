package ppp

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"io/ioutil"
	"log"
	"path/filepath"
)

var aliPayPrivateKey *rsa.PrivateKey //应用私钥
var aliPayPublicKey *rsa.PublicKey   //支付宝公钥

//AliPay使用私钥做验签
//用于同步接口请求
func AliPaySigner(data map[string]string) (signer []byte) {
	message := mapSortAndJoin(data, "=", "&")
	rng := rand.Reader
	hashed := sha256.Sum256([]byte(message))
	signer, _ = rsa.SignPKCS1v15(rng, aliPayPrivateKey, crypto.SHA256, hashed[:])
	return
}

//AliPay使用支付宝公钥做验证
//异步callback
func AliPayRSAVerify(message map[string]string, sign string) (err error) {
	digest := sha256.Sum256([]byte(mapSortAndJoin(message, "=", "&")))
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
		log.Fatal("Load AliPay PublicKey Error:", err)
	}
	aliPayPublicKey, _ = pub.(*rsa.PublicKey)
}
