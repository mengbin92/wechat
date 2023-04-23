package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
)

var xmlContentType = []string{"application/xml; charset=utf-8"}
var plainContentType = []string{"text/plain; charset=utf-8"}

func aesEncrypt(plainText []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	plainText = pkcs7Padding(plainText, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	cipherText := make([]byte, len(plainText))
	blockMode.CryptBlocks(cipherText, plainText)
	return cipherText, nil
}

func aesDecrypt(cipherText []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	plainText := make([]byte, len(cipherText))
	blockMode.CryptBlocks(plainText, cipherText)
	plainText = pkcs7UnPadding(plainText)
	return plainText, nil
}

func pkcs7UnPadding(data []byte) []byte {
	length := len(data)
	unpadding := int(data[length-1])
	return data[:(length - unpadding)]
}

func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

func decodeAESKey(key string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(key + "=")
}

func decodeNetworkByteOrder(orderBytes []byte) (n uint32) {
	return uint32(orderBytes[0])<<24 |
		uint32(orderBytes[1])<<16 |
		uint32(orderBytes[2])<<8 |
		uint32(orderBytes[3])
}

func encodeNetworkByteOrder(orderBytes []byte, n uint32) {
	orderBytes[0] = byte(n >> 24)
	orderBytes[1] = byte(n >> 16)
	orderBytes[2] = byte(n >> 8)
	orderBytes[3] = byte(n)
}

func setContentType(w http.ResponseWriter, value []string) {
	header := w.Header()

	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = value
	}
}

func responseOK(ctx *gin.Context, str string) {

}
