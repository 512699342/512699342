package utility

import (
	"testing"
)

func TestRSAEncrypt(t *testing.T) {
	name := "张三"
	encryptName, err := RSAEncrypt([]byte(name))
	if err != nil {
		t.Errorf("RSAEncrypt fail: %s", err.Error())
	} else {
		t.Logf("%v", encryptName)
	}
}

func BenchmarkEncrypt(b *testing.B) {
	name := "张三"
	for i := 0; i < b.N; i++ {
		RSAEncrypt([]byte(name))
	}
}
func TestRSAEncryptAndBase64(t *testing.T) {
	name := "张三"
	encryptName, err := RSAEncryptAndBase64([]byte(name))
	if err != nil {
		t.Errorf("RSAEncryptAndBase64 fail: %s", err.Error())
	} else {
		t.Log(encryptName)
	}
}

func BenchmarkRSAEncryptAndBase64(b *testing.B) {
	name := "张三"
	for i := 0; i < b.N; i++ {
		RSAEncryptAndBase64([]byte(name))
	}
}

func TestRSADecrypt(t *testing.T) {
	name := "张三"
	encryptName, err := RSAEncrypt([]byte(name))
	if err != nil {
		t.Errorf("RSAEncrypt fail: %s", err.Error())
	}
	decryptName, err := RSADecrypt([]byte(encryptName))
	if err != nil {
		t.Errorf("RSADecrypt fail: %s", err.Error())
	} else {
		t.Log(string(decryptName))
	}
}

func BenchmarkRSADecrypt(b *testing.B) {
	b.StopTimer()
	name := "张三"
	encryptName, _ := RSAEncrypt([]byte(name))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		RSADecrypt([]byte(encryptName))
	}
}

func TestBase64AndRSADecrypt(t *testing.T) {
	name := "张三"
	encryptName, err := RSAEncryptAndBase64([]byte(name))
	if err != nil {
		t.Errorf("RSAEncryptAndBase64 fail: %s", err.Error())
	}
	decryptName, err := Base64AndRSADecrypt(encryptName)
	if err != nil {
		t.Errorf("Base64AndRSADecrypt fail: %s", err.Error())
	} else {
		t.Log(string(decryptName))
	}
}

func BenchmarkBase64AndRSADecryptt(b *testing.B) {
	b.StopTimer()
	name := "张三"
	encryptName, _ := RSAEncryptAndBase64([]byte(name))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		Base64AndRSADecrypt(encryptName)
	}
}
