package service

// 专门负责处理发送到stream的信息的一些函数
import (
	"context"
	"time"

	pkg "QA-System/internal/pkg/redis"
)

// FromSurveyIDToStream 通过问卷ID将问卷信息发送到Redis Stream
func FromSurveyIDToStream(surveyID int) error {
	// 获取问卷信息
	survey, err := GetSurveyByID(surveyID)
	if err != nil {
		return err
	}

	creatorEmail, err1 := GetUserEmailByID(survey.UserID)
	if err1 != nil {
		return err1
	}
	// 构造消息数据
	data := map[string]any{
		"creator_email": creatorEmail,
		"survey_title":  survey.Title,
		"timestamp":     time.Now().UnixNano(),
	}

	// 发送到Redis Stream
	err = pkg.PublishToStream(context.Background(), data)
	return err
}
