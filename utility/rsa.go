package utility

import (
	"config"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"io/ioutil"
)

var (
	publickey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
)

func init() {
	var err error
	publickey, err = loadPublicKeyFile(*config.EncryptRSAPublicKey)
	if err != nil {
		panic(err)
	}
}

func parsePublicKey(keybuffer []byte) (*rsa.PublicKey, error) {
	//解密pem格式的公钥
	block, _ := pem.Decode(keybuffer)
	if block == nil {
		return nil, errors.New("public key error")
	}
	//解析公钥
	pubkeyinterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	publickey := pubkeyinterface.(*rsa.PublicKey)
	return publickey, nil
}

func loadPublicKeyFile(keyfile string) (*rsa.PublicKey, error) {
	keybuffer, err := ioutil.ReadFile(keyfile)
	if err != nil {
		return nil, err
	}
	return parsePublicKey(keybuffer)
}

func parsePrivateKey(keybuffer []byte) (*rsa.PrivateKey, error) {
	//解密pem格式的私钥
	block, _ := pem.Decode([]byte(keybuffer))
	if block == nil {
		return nil, errors.New("private key error!")
	}
	//解析PKCS1格式的私钥
	privatekey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.New("parse private key error!")
	}

	return privatekey, nil
}

// Load private key from private key file
func loadPrivateKeyFile(keyfile string) (*rsa.PrivateKey, error) {
	keybuffer, err := ioutil.ReadFile(keyfile)
	if err != nil {
		return nil, err
	}
	return parsePrivateKey(keybuffer)
}

//
func SetNewPublicKey(keybuffer []byte) error {
	pub, err := parsePublicKey(keybuffer)
	if err == nil {
		publickey = pub
	}
	return err
}

func SetNewPrivateKey(keybuffer []byte) error {
	pri, err := parsePrivateKey(keybuffer)
	if err == nil {
		privateKey = pri
	}
	return err
}

// 加密
func RSAEncrypt(origData []byte) ([]byte, error) {
	if *config.EuhtDataEncryptStategy == false {
		return origData, nil
	}
	return rsa.EncryptPKCS1v15(rand.Reader, publickey, origData)
}

// 加密,返回string
func RSAEncryptAndBase64(origData []byte) (string, error) {
	if *config.EuhtDataEncryptStategy == false {
		return string(origData), nil
	}
	data, err := rsa.EncryptPKCS1v15(rand.Reader, publickey, origData)
	if err != nil {
		return "", err
	} else {
		return base64.StdEncoding.EncodeToString(data), nil
	}
}

// 解密
func RSADecrypt(ciphertext []byte) ([]byte, error) {
	if *config.EuhtDataEncryptStategy == false {
		return ciphertext, nil
	}
	return rsa.DecryptPKCS1v15(rand.Reader, privateKey, ciphertext)
}

// 解密,返回string
func Base64AndRSADecrypt(ciphertext string) ([]byte, error) {
	if *config.EuhtDataEncryptStategy == false {
		return []byte(ciphertext), nil
	}
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	} else {
		return rsa.DecryptPKCS1v15(rand.Reader, privateKey, data)
	}
}
