package ppp

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"

	"github.com/mitchellh/packer/builder/azure/pkcs12"
)

// LoadPrivateKeyFromFile 从文件中加载私钥
func LoadPrivateKeyFromFile(file string) (key *rsa.PrivateKey, err error) {
	private, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	key, err = x509.ParsePKCS1PrivateKey(base64Decode(string(private)))
	return
}

// LoadPublicKeyFromFile 从文件中加载公钥
func LoadPublicKeyFromFile(file string) (key *rsa.PublicKey, err error) {
	public, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(public)
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	key, _ = pub.(*rsa.PublicKey)
	return
}

// LoadCertFromP12 从p12中加载证书
func LoadCertFromP12(file, pwd string) (cert tls.Certificate, err error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return cert, err
	}

	blocks, err := pkcs12.ToPEM(b, pwd)
	if err != nil {
		return cert, err
	}

	var pemData []byte
	for _, b := range blocks {
		pemData = append(pemData, pem.EncodeToMemory(b)...)
	}

	cert, err = tls.X509KeyPair(pemData, pemData)
	return
}
