package session

import (
	"github.com/SituChengxiang/WeJH-SDK/sessionHelper"
	"github.com/gin-gonic/gin"
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
