package security

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

type AESCipher struct {
	key []byte
	iv  []byte
}

func NewAESCipher(key, iv string) (*AESCipher, error) {
	if len([]byte(key)) != aes.BlockSize || len([]byte(iv)) != aes.BlockSize {
		return nil, fmt.Errorf("AES key 和 IV 必须为 16 字节")
	}
	return &AESCipher{key: []byte(key), iv: []byte(iv)}, nil
}

func (a *AESCipher) Encrypt(plain []byte) (string, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return "", err
	}
	padded := pkcs7Pad(plain, block.BlockSize())
	encrypted := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, a.iv).CryptBlocks(encrypted, padded)
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func (a *AESCipher) Decrypt(encoded string) ([]byte, error) {
	encrypted, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("Base64 解码: %w", err)
	}
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return nil, err
	}
	if len(encrypted) == 0 || len(encrypted)%block.BlockSize() != 0 {
		return nil, fmt.Errorf("密文长度无效")
	}
	plain := make([]byte, len(encrypted))
	cipher.NewCBCDecrypter(block, a.iv).CryptBlocks(plain, encrypted)
	return pkcs7Unpad(plain, block.BlockSize())
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	result := make([]byte, len(data)+padding)
	copy(result, data)
	for i := len(data); i < len(result); i++ {
		result[i] = byte(padding)
	}
	return result
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, fmt.Errorf("PKCS7 数据长度无效")
	}
	padding := int(data[len(data)-1])
	if padding < 1 || padding > blockSize || padding > len(data) {
		return nil, fmt.Errorf("PKCS7 padding 无效")
	}
	for _, value := range data[len(data)-padding:] {
		if int(value) != padding {
			return nil, fmt.Errorf("PKCS7 padding 无效")
		}
	}
	return data[:len(data)-padding], nil
}
