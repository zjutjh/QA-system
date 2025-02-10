package utils

import (
	"errors"

	"QA-System/internal/global/config"
	"github.com/SituChengxiang/WeJH-SDK/aesHelper"
	"go.uber.org/zap"
)

var encryptKey string

// Init 读入 AES 密钥配置
func Init() error {
	encryptKey = config.Config.GetString("aes.key")
	if len(encryptKey) != 16 && len(encryptKey) != 24 && len(encryptKey) != 32 {
		return errors.New("AES 密钥长度必须为 16、24 或 32 字节")
	}
	err := aesHelper.Init(encryptKey)
	return err
}

// AesEncrypt AES加密
func AesEncrypt(orig string) string {
	e, err := aesHelper.Encrypt(orig)
	if err != nil {
		zap.L().Error(err.Error())
		return ""
	}
	return e
}

// AesDecrypt AES解密
func AesDecrypt(cryted string) string {
	d, err := aesHelper.Decrypt(cryted)
	if err != nil {
		zap.L().Error(err.Error())
		return ""
	}
	return d
}
