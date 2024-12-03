package user

import (
	"QA-System/internal/dao"
	"QA-System/internal/handler/queue"
	"QA-System/internal/pkg/code"
	"QA-System/internal/pkg/queue/asynq"
	"QA-System/internal/pkg/utils"
	"QA-System/internal/service"
	"errors"
	"strconv"

	"time"

	"github.com/gin-gonic/gin"
)

type SubmitServeyData struct {
	ID            int                 `json:"id" binding:"required"`
	StudentID     string              `json:"stu_id"`
	QuestionsList []dao.QuestionsList `json:"questions_list"`
}

func SubmitSurvey(c *gin.Context) {
	var data SubmitServeyData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取参数失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ParamError)
		return
	}
	// 判断问卷问题和答卷问题数目是否一致
	survey, err := service.GetSurveyByID(data.ID)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取问卷失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	questions, err := service.GetQuestionsBySurveyID(survey.ID)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取问题失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	if len(questions) != len(data.QuestionsList) {
		c.Error(&gin.Error{Err: errors.New("问卷问题和上传问题数量不一致"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	// 判断填写时间是否在问卷有效期内
	if !survey.Deadline.IsZero() && survey.Deadline.Before(time.Now()) {
		c.Error(&gin.Error{Err: errors.New("填写时间已过"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.TimeBeyondError)
		return
	}
	// 判断问卷是否开放
	if survey.Status != 2 {
		c.Error(&gin.Error{Err: errors.New("问卷未开放"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.SurveyNotOpen)
		return
	}
	// 逐个判断问题答案
	for _, q := range data.QuestionsList {
		question, err := service.GetQuestionByID(q.QuestionID)
		if err != nil {
			c.Error(&gin.Error{Err: errors.New("获取问题失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.ServerError)
			return
		}
		if question.SerialNum != q.SerialNum {
			c.Error(&gin.Error{Err: errors.New("问题序号" + strconv.Itoa(question.ID) + "和" + strconv.Itoa(q.SerialNum) + "不一致"), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.ServerError)
			return
		}
		if question.SurveyID != survey.ID {
			c.Error(&gin.Error{Err: errors.New("问题" + strconv.Itoa(question.SerialNum) + "不属于该问卷"), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.ServerError)
			return
		}
		// 判断必填字段是否为空
		if question.Required && q.Answer == "" {
			c.Error(&gin.Error{Err: errors.New("问题" + strconv.Itoa(q.SerialNum) + "必填字段为空"), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.ServerError)
			return
		}
		// 判断多选题选项数量是否符合要求
		if question.QuestionType == 2 {
			if question.MinimumOption != 0 && len(q.Answer) < int(question.MinimumOption) {
				c.Error(&gin.Error{Err: errors.New("问题" + strconv.Itoa(q.SerialNum) + "选项数量不符合要求"), Type: gin.ErrorTypeAny})
				utils.JsonErrorResponse(c, code.OptionNumError)
				return
			}
			if question.MaximumOption != 0 && len(q.Answer) > int(question.MaximumOption) {
				c.Error(&gin.Error{Err: errors.New("问题" + strconv.Itoa(q.SerialNum) + "选项数量不符合要求"), Type: gin.ErrorTypeAny})
				utils.JsonErrorResponse(c, code.OptionNumError)
				return
			}
		}
	}
	if survey.DailyLimit != 0 && survey.Verify == true {
		limit, err := service.GetUserLimit(c, data.StudentID, survey.ID)
		if err != nil {
			c.Error(&gin.Error{Err: errors.New("获取用户投票次数失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.ServerError)
			return
		}
		if limit >= int(survey.DailyLimit) {
			c.Error(&gin.Error{Err: errors.New("投票次数已达上限"), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.VoteLimitError)
			return
		}
	}
	// 创建并入队任务
	task, err := queue.NewSubmitSurveyTask(data.ID, data.QuestionsList)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("创建任务失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}

	_, err = asynq.Client.Enqueue(task)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("任务入队失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	err = service.InscUserLimit(c, data.StudentID, survey.ID)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("更新用户投票次数失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	utils.JsonSuccessResponse(c, nil)
}

type GetSurveyData struct {
	ID int `form:"id" binding:"required"`
}

type SurveyData struct {
	ID        int            `json:"id"`
	Time      string         `json:"time"`
	Desc      string         `json:"desc"`
	Img       string         `json:"img"`
	Questions []dao.Question `json:"questions"`
}

// 用户获取问卷
func GetSurvey(c *gin.Context) {
	var data GetSurveyData
	err := c.ShouldBindQuery(&data)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取参数失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ParamError)
		return
	}
	// 获取问卷
	survey, err := service.GetSurveyByID(data.ID)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取问卷失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	// 判断填写时间是否在问卷有效期内
	if !survey.Deadline.IsZero() && survey.Deadline.Before(time.Now()) {
		c.Error(&gin.Error{Err: errors.New("问卷填写时间已截至"), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.TimeBeyondError)
		return
	}
	// 获取相应的问题
	questions, err := service.GetQuestionsBySurveyID(survey.ID)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取问题失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	// 构建问卷响应
	questionsResponse := make([]map[string]interface{}, 0)
	for _, question := range questions {
		options, err := service.GetOptionsByQuestionID(question.ID)
		if err != nil {
			c.Error(&gin.Error{Err: errors.New("获取选项失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
			utils.JsonErrorResponse(c, code.ServerError)
			return
		}
		optionsResponse := make([]map[string]interface{}, 0)
		for _, option := range options {
			optionResponse := map[string]interface{}{
				"img":         option.Img,
				"content":     option.Content,
				"description": option.Description,
				"serial_num":  option.SerialNum,
			}
			optionsResponse = append(optionsResponse, optionResponse)
		}
		questionMap := map[string]interface{}{
			"id":             question.ID,
			"serial_num":     question.SerialNum,
			"subject":        question.Subject,
			"describe":       question.Description,
			"required":       question.Required,
			"unique":         question.Unique,
			"other_option":   question.OtherOption,
			"img":            question.Img,
			"question_type":  question.QuestionType,
			"reg":            question.Reg,
			"maximum_option": question.MaximumOption,
			"minimum_option": question.MinimumOption,
			"options":        optionsResponse,
		}
		questionsResponse = append(questionsResponse, questionMap)
	}
	response := map[string]interface{}{
		"id":          survey.ID,
		"title":       survey.Title,
		"time":        survey.Deadline,
		"desc":        survey.Desc,
		"img":         survey.Img,
		"daily_limit": survey.DailyLimit,
		"verify":      survey.Verify,
		"survey_type": survey.Type,
		"questions":   questionsResponse,
	}

	utils.JsonSuccessResponse(c, response)
}

// 上传图片
func UploadImg(c *gin.Context) {
	url, err := service.HandleImgUpload(c)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("上传图片失败" + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	utils.JsonSuccessResponse(c, url)
}

// 上传文件
func UploadFile(c *gin.Context) {
	url, err := service.HandleFileUpload(c)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("上传文件失败" + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	utils.JsonSuccessResponse(c, url)
}

type OauthData struct {
	StudenID string `json:"stu_id" binding:"required"`
	Password string `json:"password" binding:"required"`
	SurveyID int    `json:"survey_id" binding:"required"`
}

func Oauth(c *gin.Context) {
	var data OauthData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("获取参数失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ParamError)
		return
	}
	err = service.Oauth(data.StudenID, data.Password)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("认证失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	err = service.SetUserLimit(c, data.StudenID, data.SurveyID, 0)
	if err != nil {
		c.Error(&gin.Error{Err: errors.New("创建用户投票次数失败原因: " + err.Error()), Type: gin.ErrorTypeAny})
		utils.JsonErrorResponse(c, code.ServerError)
		return
	}
	utils.JsonSuccessResponse(c, nil)
}
