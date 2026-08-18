package security

import "testing"

func TestAESCipherRoundTrip(t *testing.T) {
	cipher, err := NewAESCipher("7546455674406856", "7513956994549178")
	if err != nil {
		t.Fatal(err)
	}
	plain := []byte(`{"title":"中文内容","count":2}`)
	encrypted, err := cipher.Encrypt(plain)
	if err != nil {
		t.Fatal(err)
	}
	decrypted, err := cipher.Decrypt(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if string(decrypted) != string(plain) {
		t.Fatalf("解密结果不一致: %s", decrypted)
	}
}

func TestAESCipherRejectsInvalidPadding(t *testing.T) {
	cipher, err := NewAESCipher("7546455674406856", "7513956994549178")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cipher.Decrypt("aW52YWxpZA=="); err == nil {
		t.Fatal("无效密文应返回错误")
	}
}
