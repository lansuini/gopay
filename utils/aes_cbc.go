package utils

import (
	"crypto/aes"
	"github.com/lgbya/go-dump"
	"luckypay/pkg/config"
)
import "crypto/cipher"
import "bytes"
import "encoding/base64"
import "log"

var AesCbc = newAesCbc()

func newAesCbc() *aes_cbc {
	return &aes_cbc{}
}

type aes_cbc struct {
	key []byte
	iv  []byte
}

func (t *aes_cbc) init() {
	t.key = []byte(config.Instance.DataSalt)
	t.iv = []byte(config.Instance.DataSaltIV)
}
func (t *aes_cbc) Encrypt(text []byte) (string, error) {
	//生成cipher.Block 数据块
	block, err := aes.NewCipher(t.key)
	if err != nil {
		log.Println("错误 -" + err.Error())
		return "", err
	}
	//填充内容，如果不足16位字符
	blockSize := block.BlockSize()
	originData := t.pad(text, blockSize)
	//加密方式
	blockMode := cipher.NewCBCEncrypter(block, t.iv)
	//加密，输出到[]byte数组
	crypted := make([]byte, len(originData))
	blockMode.CryptBlocks(crypted, originData)
	return base64.StdEncoding.EncodeToString(crypted), nil
}

func (t *aes_cbc) pad(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func (t *aes_cbc) Decrypt(text string) (string, error) {
	dump.Printf(string(t.key))
	decode_data, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return "", nil
	}
	//生成密码数据块cipher.Block
	block, err := aes.NewCipher(t.key)
	if err != nil {
		return "", err
	}
	//解密模式
	blockMode := cipher.NewCBCDecrypter(block, t.iv)
	//输出到[]byte数组
	origin_data := make([]byte, len(decode_data))
	blockMode.CryptBlocks(origin_data, decode_data)
	//去除填充,并返回
	return string(t.unpad(origin_data)), nil
}

func (t *aes_cbc) unpad(ciphertext []byte) []byte {
	length := len(ciphertext)
	//去掉最后一次的padding
	unpadding := int(ciphertext[length-1])
	return ciphertext[:(length - unpadding)]
}
