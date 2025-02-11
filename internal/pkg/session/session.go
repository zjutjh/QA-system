package session

import (
	"github.com/gin-gonic/gin"
	"github.com/zjutjh/WeJH-SDK/sessionHelper"
	"go.uber.org/zap"
)

// Init 初始化Session会话管理
func Init(r *gin.Engine) {
	config := getConfig()
	err := sessionHelper.Init(&config, r)
	if err != nil {
		zap.L().Fatal(err.Error())
	}
}
