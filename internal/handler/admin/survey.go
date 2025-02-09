package admin

import (
	"errors"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"QA-System/internal/dao"
	"QA-System/internal/model"
	"QA-System/internal/pkg/code"
	"QA-System/internal/pkg/utils"
	"QA-System/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type createSurveyData struct {
	Title      string         `json:"title"`
	Desc       string         `json:"desc" `
	Img        string         `json:"img" `
	Status     int            `json:"status" binding:"required,oneof=1 2"`
	StartTime  string         `json:"start_time"`
	Time       string         `json:"time"`
	DailyLimit uint           `json:"day_limit"`   // 问卷每日填写限制
	SurveyType uint           `json:"survey_type"` // 问卷类型 0:调研 1:投票
	Verify     bool           `json:"verify"`      // 问卷是否需要统一验证
	Questions  []dao.Question `json:"questions"`
}

// CreateSurvey 创建问卷
func CreateSurvey(c *gin.Context) {
	var data createSurveyData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	// 鉴权
	user, err := service.GetUserSession(c)
	if err != nil {
		code.AbortWithException(c, code.NotLogin, err)
		return
	}
	// 解析时间转换为中国时间(UTC+8)
	ddlTime, err := time.Parse(time.RFC3339, data.Time)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	startTime, err := time.Parse(time.RFC3339, data.StartTime)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	if startTime.After(ddlTime) {
		code.AbortWithException(c, code.SurveyError, errors.New("开始时间晚于截止时间"))
		return
	}
	// 检查问卷每个题目的序号没有重复且按照顺序递增
	questionNumMap := make(map[int]bool)
	for i, question := range data.Questions {
		if data.SurveyType == 2 && (question.QuestionType != 2 && !question.Required) {
			code.AbortWithException(c, code.SurveyError, errors.New("投票题目只能为多选必填题"))
			return
		}
		if questionNumMap[question.SerialNum] {
			code.AbortWithException(c, code.SurveyError, errors.New("题目序号"+strconv.Itoa(question.SerialNum)+"重复"))
			return
		}
		if i > 0 && question.SerialNum != data.Questions[i-1].SerialNum+1 {
			code.AbortWithException(c, code.SurveyError, errors.New("题目序号不按顺序递增"))
			return
		}
		questionNumMap[question.SerialNum] = true
		question.SerialNum = i + 1

		// 检测多选题目的最多选项数和最少选项数
		if (question.QuestionType == 2 && data.SurveyType == 0) ||
			(question.QuestionType == 1 && data.SurveyType == 1) &&
				(question.MaximumOption < question.MinimumOption) {
			code.AbortWithException(c, code.OptionNumError, errors.New("多选最多选项数小于最少选项数"))
			return
		}
		// 检查多选选项和最少选项数是否符合要求
		if (question.QuestionType == 2 && data.SurveyType == 0) ||
			(question.QuestionType == 1 && data.SurveyType == 1) &&
				uint(len(question.Options)) < question.MinimumOption {
			code.AbortWithException(c, code.OptionNumError, errors.New("选项数量小于最少选项数"))
			return
		}
		// 检查最多选项数是否符合要求
		if (question.QuestionType == 2 && data.SurveyType == 0) ||
			(question.QuestionType == 1 && data.SurveyType == 1) &&
				question.MaximumOption <= 0 {
			code.AbortWithException(c, code.OptionNumError, errors.New("最多选项数小于等于0"))
			return
		}
	}
	// 检测问卷是否填写完整
	if data.Status == 2 {
		if data.Title == "" || len(data.Questions) == 0 {
			code.AbortWithException(c, code.SurveyIncomplete, errors.New("问卷标题为空或问卷没有问题"))
			return
		}
		questionMap := make(map[string]bool)
		for _, question := range data.Questions {
			if question.Subject == "" {
				code.AbortWithException(c, code.SurveyIncomplete,
					errors.New("问题"+strconv.Itoa(question.SerialNum)+"标题为空"))
				return
			}
			if questionMap[question.Subject] {
				code.AbortWithException(c, code.SurveyContentRepeat,
					errors.New("问题"+strconv.Itoa(question.SerialNum)+"题目"+question.Subject+"重复"))
				return
			}
			questionMap[question.Subject] = true
			if question.QuestionType == 1 || question.QuestionType == 2 {
				if len(question.Options) < 1 {
					code.AbortWithException(c, code.SurveyIncomplete,
						errors.New("问题"+strconv.Itoa(question.SerialNum)+"选项数量太少"))
					return
				}
				optionMap := make(map[string]bool)
				for _, option := range question.Options {
					if option.Content == "" {
						code.AbortWithException(c, code.SurveyIncomplete,
							errors.New("选项"+strconv.Itoa(option.SerialNum)+"内容为空"))
						return
					}
					if optionMap[option.Content] {
						code.AbortWithException(c, code.SurveyContentRepeat,
							errors.New("选项内容"+option.Content+"重复"))
						return
					}
					optionMap[option.Content] = true
				}
			}
		}
	}
	// 创建问卷
	err = service.CreateSurvey(user.ID, data.Title, data.Desc, data.Img, data.Questions,
		data.Status, data.SurveyType, data.DailyLimit, data.Verify, ddlTime, startTime)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	utils.JsonSuccessResponse(c, nil)
}

type updateSurveyStatusData struct {
	UUID   string `json:"uuid" binding:"required"`
	Status int    `json:"status" binding:"required,oneof=1 2"`
}

// UpdateSurveyStatus 修改问卷状态
func UpdateSurveyStatus(c *gin.Context) {
	var data updateSurveyStatusData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	// 鉴权
	user, err := service.GetUserSession(c)
	if err != nil {
		code.AbortWithException(c, code.NotLogin, err)
		return
	}
	// 获取问卷
	survey, err := service.GetSurveyByUUID(data.UUID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	// 判断权限
	if (user.AdminType != 2) && (user.AdminType != 1 || survey.UserID != user.ID) &&
		!service.UserInManage(user.ID, survey.UUID) {
		code.AbortWithException(c, code.NoPermission, errors.New(user.Username+"无权限"))
		return
	}
	// 判断问卷状态
	if survey.Status == data.Status {
		code.AbortWithException(c, code.StatusRepeatError, errors.New("问卷状态重复"))
		return
	}
	// 检测问卷是否填写完整
	if data.Status == 2 {
		if survey.Title == "" {
			code.AbortWithException(c, code.SurveyIncomplete, errors.New("问卷信息填写不完整"))
			return
		}
		questions, err := service.GetQuestionsBySurveyID(survey.UUID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			code.AbortWithException(c, code.SurveyIncomplete, errors.New("问卷问题不存在"))
			return
		} else if err != nil {
			code.AbortWithException(c, code.ServerError, err)
			return
		}
		questionMap := make(map[string]bool)
		for _, question := range questions {
			if question.Subject == "" {
				code.AbortWithException(c, code.SurveyIncomplete,
					errors.New("问题"+strconv.Itoa(question.SerialNum)+"内容填写为空"))
				return
			}
			if questionMap[question.Subject] {
				code.AbortWithException(c, code.SurveyContentRepeat,
					errors.New("问题题目"+question.Subject+"重复"))
				return
			}
			questionMap[question.Subject] = true
			if question.QuestionType == 1 || question.QuestionType == 2 {
				options, err := service.GetOptionsByQuestionID(question.ID)
				if err != nil {
					code.AbortWithException(c, code.ServerError, err)
					return
				}
				if len(options) < 1 {
					code.AbortWithException(c, code.SurveyIncomplete,
						errors.New("问题"+strconv.Itoa(question.ID)+"选项太少"))
					return
				}
				optionMap := make(map[string]bool)
				for _, option := range options {
					if option.Content == "" {
						code.AbortWithException(c, code.SurveyIncomplete,
							errors.New("选项"+strconv.Itoa(option.SerialNum)+"内容未填"))
						return
					}
					if optionMap[option.Content] {
						code.AbortWithException(c, code.SurveyContentRepeat,
							errors.New("选项内容"+option.Content+"重复"))
						return
					}
					optionMap[option.Content] = true
				}
			}
		}
	}
	// 修改问卷状态
	err = service.UpdateSurveyStatus(data.UUID, data.Status)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	utils.JsonSuccessResponse(c, nil)
}

type updateSurveyData struct {
	UUID       string         `json:"uuid" binding:"required"`
	Title      string         `json:"title"`
	Desc       string         `json:"desc" `
	Img        string         `json:"img" `
	Time       string         `json:"time"`
	StartTime  string         `json:"start_time"`
	DailyLimit uint           `json:"day_limit"`   // 问卷每日填写限制
	SurveyType uint           `json:"survey_type"` // 问卷类型 1:调研 2:投票
	Verify     bool           `json:"verify"`      // 问卷是否需要统一验证
	Questions  []dao.Question `json:"questions"`
}

// UpdateSurvey 修改问卷
func UpdateSurvey(c *gin.Context) {
	var data updateSurveyData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	// 鉴权
	user, err := service.GetUserSession(c)
	if err != nil {
		code.AbortWithException(c, code.NotLogin, err)
		return
	}
	// 获取问卷
	survey, err := service.GetSurveyByUUID(data.UUID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	// 判断权限
	if (user.AdminType != 2) && (user.AdminType != 1 || survey.UserID != user.ID) &&
		!service.UserInManage(user.ID, survey.UUID) {
		code.AbortWithException(c, code.NoPermission, errors.New(user.Username+"无权限"))
		return
	}
	// 判断问卷状态
	if survey.Status != 1 {
		code.AbortWithException(c, code.StatusOpenError, errors.New("问卷状态不为未发布"))
		return
	}
	// 判断问卷的填写数量是否为零
	if survey.Num != 0 {
		code.AbortWithException(c, code.SurveyNumError, errors.New("问卷已有填写数量"))
		return
	}
	// 解析时间转换为中国时间(UTC+8)
	ddlTime, err := time.Parse(time.RFC3339, data.Time)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	startTime, err := time.Parse(time.RFC3339, data.StartTime)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	if startTime.After(ddlTime) {
		code.AbortWithException(c, code.SurveyError, errors.New("开始时间晚于截止时间"))
		return
	}
	// 检查问卷每个题目的序号没有重复且按照顺序递增
	questionNumMap := make(map[int]bool)
	for i, question := range data.Questions {
		if questionNumMap[question.SerialNum] {
			code.AbortWithException(c, code.SurveyError, errors.New("题目序号"+strconv.Itoa(question.SerialNum)+"重复"))
			return
		}
		if i > 0 && question.SerialNum != data.Questions[i-1].SerialNum+1 {
			code.AbortWithException(c, code.SurveyError, errors.New("题目序号不按顺序递增"))
			return
		}
		questionNumMap[question.SerialNum] = true
		question.SerialNum = i + 1

		// 检测多选题目的最多选项数和最少选项数
		if (question.QuestionType == 2 && survey.Type == 0) ||
			(question.QuestionType == 1 && survey.Type == 1) &&
				(question.MaximumOption < question.MinimumOption) {
			code.AbortWithException(c, code.OptionNumError, errors.New("多选最多选项数小于最少选项数"))
			return
		}
		// 检查多选选项和最少选项数是否符合要求
		if (question.QuestionType == 2 && survey.Type == 0) ||
			(question.QuestionType == 1 && survey.Type == 1) &&
				uint(len(question.Options)) < question.MinimumOption {
			code.AbortWithException(c, code.OptionNumError, errors.New("选项数量小于最少选项数"))
			return
		}
		// 检查最多选项数是否符合要求
		if (question.QuestionType == 2 && survey.Type == 0) ||
			(question.QuestionType == 1 && survey.Type == 1) &&
				question.MaximumOption <= 0 {
			code.AbortWithException(c, code.OptionNumError, errors.New("最多选项数小于等于0"))
			return
		}
	}
	// 修改问卷
	err = service.UpdateSurvey(data.UUID, data.SurveyType, data.DailyLimit,
		data.Verify, data.Title, data.Desc, data.Img, data.Questions, ddlTime, startTime)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	utils.JsonSuccessResponse(c, nil)
}

type deleteSurveyData struct {
	UUID string `form:"uuid" binding:"required"`
}

// DeleteSurvey 删除问卷
func DeleteSurvey(c *gin.Context) {
	var data deleteSurveyData
	err := c.ShouldBindQuery(&data)
	if err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	// 鉴权
	user, err := service.GetUserSession(c)
	if err != nil {
		code.AbortWithException(c, code.NotLogin, err)
		return
	}
	// 获取问卷
	survey, err := service.GetSurveyByUUID(data.UUID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		code.AbortWithException(c, code.SurveyNotExist, errors.New("问卷不存在"))
		return
	} else if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	// 判断权限
	if (user.AdminType != 2) && (user.AdminType != 1 || survey.UserID != user.ID) &&
		!service.UserInManage(user.ID, survey.UUID) {
		code.AbortWithException(c, code.NoPermission, errors.New(user.Username+"无权限"))
		return
	}
	// 删除问卷
	err = service.DeleteSurvey(data.UUID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	err = service.DeleteOauthRecord(data.UUID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	utils.JsonSuccessResponse(c, nil)
}

type getSurveyAnswersData struct {
	UUID     string `form:"uuid" binding:"required"`
	Text     string `form:"text"`
	Unique   bool   `form:"unique"`
	PageNum  int    `form:"page_num" binding:"required"`
	PageSize int    `form:"page_size" binding:"required"`
}

// GetSurveyAnswers 获取问卷收集数据
func GetSurveyAnswers(c *gin.Context) {
	var data getSurveyAnswersData
	err := c.ShouldBindQuery(&data)
	if err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	// 鉴权
	user, err := service.GetUserSession(c)
	if err != nil {
		code.AbortWithException(c, code.NotLogin, err)
		return
	}
	// 获取问卷
	survey, err := service.GetSurveyByUUID(data.UUID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		code.AbortWithException(c, code.SurveyNotExist, errors.New("问卷不存在"))
		return
	} else if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	// 判断权限
	if (user.AdminType != 2) && (user.AdminType != 1 || survey.UserID != user.ID) &&
		!service.UserInManage(user.ID, survey.UUID) {
		code.AbortWithException(c, code.NoPermission, errors.New(user.Username+"无权限"))
		return
	}
	// 获取问卷收集数据
	var num *int64
	answers, num, err := service.GetSurveyAnswers(data.UUID, data.PageNum, data.PageSize, data.Text, data.Unique)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	utils.JsonSuccessResponse(c, gin.H{
		"answers_data":   answers,
		"total_page_num": math.Ceil(float64(*num) / float64(data.PageSize)),
	})
}

type getAllSurveyData struct {
	PageNum  int    `form:"page_num" binding:"required"`
	PageSize int    `form:"page_size" binding:"required"`
	Title    string `form:"title"`
}

// GetAllSurvey 获取所有问卷
func GetAllSurvey(c *gin.Context) {
	var data getAllSurveyData
	err := c.ShouldBindQuery(&data)
	if err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	user, err := service.GetUserSession(c)
	if err != nil {
		code.AbortWithException(c, code.NotLogin, err)
		return
	}
	// 获取问卷
	var response []any
	var surveys []model.Survey
	var totalPageNum *int64
	if user.AdminType == 2 {
		surveys, totalPageNum, err = service.GetAllSurvey(data.PageNum, data.PageSize, data.Title)
		if err != nil {
			code.AbortWithException(c, code.ServerError, err)
			return
		}
		surveys = service.SortSurvey(surveys)
		response = service.GetSurveyResponse(surveys)
	} else {
		surveys, err = service.GetAllSurveyByUserID(user.ID)
		if err != nil {
			code.AbortWithException(c, code.ServerError, err)
			return
		}
		managedSurveys, err := service.GetManagedSurveyByUserID(user.ID)
		if err != nil {
			code.AbortWithException(c, code.ServerError, err)
			return
		}
		for _, manage := range managedSurveys {
			managedSurvey, err := service.GetSurveyByUUID(manage.SurveyID)
			if err != nil {
				code.AbortWithException(c, code.ServerError, err)
				return
			}
			surveys = append(surveys, *managedSurvey)
		}
		surveys = service.SortSurvey(surveys)
		response = service.GetSurveyResponse(surveys)
		response, totalPageNum = service.ProcessResponse(response, data.PageNum, data.PageSize, data.Title)
	}

	utils.JsonSuccessResponse(c, gin.H{
		"survey_list":    response,
		"total_page_num": math.Ceil(float64(*totalPageNum) / float64(data.PageSize)),
	})
}

type getSurveyData struct {
	UUID string `form:"uuid" binding:"required"`
}

// GetSurvey 管理员获取问卷题面
func GetSurvey(c *gin.Context) {
	var data getSurveyData
	err := c.ShouldBindQuery(&data)
	if err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	user, err := service.GetUserSession(c)
	if err != nil {
		code.AbortWithException(c, code.NotLogin, err)
		return
	}
	// 获取问卷
	survey, err := service.GetSurveyByUUID(data.UUID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	// 判断权限
	if (user.AdminType != 2) && (user.AdminType != 1 || survey.UserID != user.ID) &&
		!service.UserInManage(user.ID, survey.UUID) {
		code.AbortWithException(c, code.NoPermission, errors.New(user.Username+"无权限"))
		return
	}
	// 获取相应的问题
	questions, err := service.GetQuestionsBySurveyID(survey.UUID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	// 构建问卷响应
	questionsResponse := make([]map[string]any, 0)
	for _, question := range questions {
		options, err := service.GetOptionsByQuestionID(question.ID)
		if err != nil {
			code.AbortWithException(c, code.ServerError, err)
			return
		}
		optionsResponse := make([]map[string]any, 0)
		for _, option := range options {
			optionResponse := map[string]any{
				"img":         option.Img,
				"content":     option.Content,
				"description": option.Description,
				"serial_num":  option.SerialNum,
			}
			optionsResponse = append(optionsResponse, optionResponse)
		}
		questionMap := map[string]any{
			"id":             question.SerialNum,
			"serial_num":     question.SerialNum,
			"subject":        question.Subject,
			"description":    question.Description,
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
	response := map[string]any{
		"id":          survey.UUID,
		"title":       survey.Title,
		"time":        survey.Deadline,
		"desc":        survey.Desc,
		"img":         survey.Img,
		"status":      survey.Status,
		"survey_type": survey.Type,
		"verify":      survey.Verify,
		"day_limit":   survey.DailyLimit,
		"start_time":  survey.StartTime,
		"questions":   questionsResponse,
	}

	utils.JsonSuccessResponse(c, response)
}

type downloadFileData struct {
	UUID string `form:"uuid" binding:"required"`
}

// DownloadFile 下载
func DownloadFile(c *gin.Context) {
	var data downloadFileData
	err := c.ShouldBindQuery(&data)
	if err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}
	user, err := service.GetUserSession(c)
	if err != nil {
		code.AbortWithException(c, code.NotLogin, err)
		return
	}
	// 获取问卷
	survey, err := service.GetSurveyByUUID(data.UUID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	// 判断权限
	if (user.AdminType != 2) && (user.AdminType != 1 || survey.UserID != user.ID) &&
		!service.UserInManage(user.ID, survey.UUID) {
		code.AbortWithException(c, code.NoPermission, errors.New(user.Username+"无权限"))
		return
	}
	// 获取数据
	answers, err := service.GetAllSurveyAnswers(data.UUID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	url, err := service.HandleDownloadFile(answers, survey)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	utils.JsonSuccessResponse(c, url)
}

type getSurveyStatisticsData struct {
	UUID     string `form:"uuid" binding:"required"`
	PageNum  int    `form:"page_num" binding:"required"`
	PageSize int    `form:"page_size" binding:"required"`
}

type getOptionCount struct {
	SerialNum int    `json:"serial_num"` // 选项序号
	Content   string `json:"content"`    // 选项内容
	Count     int    `json:"count"`      // 选项数量
}

type getSurveyStatisticsResponse struct {
	SerialNum    int              `json:"serial_num"`    // 问题序号
	Question     string           `json:"question"`      // 问题内容
	QuestionType int              `json:"question_type"` // 问题类型  1:单选 2:多选
	Options      []getOptionCount `json:"options"`       // 选项内容
}

// GetSurveyStatistics 获取统计问卷选择题数据
func GetSurveyStatistics(c *gin.Context) {
	var data getSurveyStatisticsData
	if err := c.ShouldBindQuery(&data); err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}

	user, err := service.GetUserSession(c)
	if err != nil {
		code.AbortWithException(c, code.NotLogin, err)
		return
	}

	survey, err := service.GetSurveyByUUID(data.UUID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}

	if (user.AdminType != 2) && (user.AdminType != 1 || survey.UserID != user.ID) &&
		!service.UserInManage(user.ID, survey.UUID) {
		code.AbortWithException(c, code.NoPermission, errors.New(user.Username+"无权限"))
		return
	}

	answersheets, err := service.GetSurveyAnswersBySurveyID(data.UUID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}

	questions, err := service.GetQuestionsBySurveyID(data.UUID)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}

	questionMap := make(map[int]model.Question)
	optionsMap := make(map[int][]model.Option)
	optionAnswerMap := make(map[int]map[string]model.Option)
	optionSerialNumMap := make(map[int]map[int]model.Option)
	for _, question := range questions {
		questionMap[question.ID] = question
		optionAnswerMap[question.ID] = make(map[string]model.Option)
		optionSerialNumMap[question.ID] = make(map[int]model.Option)
		options, err := service.GetOptionsByQuestionID(question.ID)
		if err != nil {
			code.AbortWithException(c, code.ServerError, err)
			return
		}
		optionsMap[question.ID] = options
		for _, option := range options {
			optionAnswerMap[question.ID][option.Content] = option
			optionSerialNumMap[question.ID][option.SerialNum] = option
		}
	}

	optionCounts := make(map[int]map[int]int)
	for _, sheet := range answersheets {
		for _, answer := range sheet.Answers {
			options := optionsMap[answer.QuestionID]
			question := questionMap[answer.QuestionID]
			// 初始化选项统计（确保每个选项的计数存在且为 0）
			if _, initialized := optionCounts[question.ID]; !initialized {
				counts := ensureMap(optionCounts, question.ID)
				for _, option := range options {
					counts[option.SerialNum] = 0
				}
			}
			if question.QuestionType == 1 || question.QuestionType == 2 {
				answerOptions := strings.Split(answer.Content, "┋")
				questionOptions := optionAnswerMap[answer.QuestionID]
				for _, answerOption := range answerOptions {
					// 查找选项
					if questionOptions != nil {
						option, exists := questionOptions[answerOption]
						if exists {
							// 如果找到选项，处理逻辑
							ensureMap(optionCounts, answer.QuestionID)[option.SerialNum]++
							continue
						}
					}
					// 如果选项不存在，处理为 "其他" 选项
					ensureMap(optionCounts, answer.QuestionID)[0]++
				}
			}
		}
	}
	response := make([]getSurveyStatisticsResponse, 0, len(optionCounts))
	for qid, options := range optionCounts {
		q := questionMap[qid]
		var qOptions []getOptionCount
		if q.OtherOption {
			qOptions = make([]getOptionCount, 0, len(options)+1)
			// 添加其他选项
			qOptions = append(qOptions, getOptionCount{
				SerialNum: 0,
				Content:   "其他",
				Count:     options[0],
			})
		} else {
			qOptions = make([]getOptionCount, 0, len(options))
		}

		// 按序号排序
		sortedSerialNums := make([]int, 0, len(options))
		for oSerialNum := range options {
			sortedSerialNums = append(sortedSerialNums, oSerialNum)
		}
		sort.Ints(sortedSerialNums)
		for _, oSerialNum := range sortedSerialNums {
			if oSerialNum == 0 {
				continue
			}
			count := options[oSerialNum]
			op := optionSerialNumMap[qid][oSerialNum]
			qOptions = append(qOptions, getOptionCount{
				SerialNum: op.SerialNum,
				Content:   op.Content,
				Count:     count,
			})
		}
		response = append(response, getSurveyStatisticsResponse{
			SerialNum:    q.SerialNum,
			Question:     q.Subject,
			QuestionType: q.QuestionType,
			Options:      qOptions,
		})
	}
	start := (data.PageNum - 1) * data.PageSize
	end := start + data.PageSize
	// 确保 start 和 end 在有效范围内
	if start < 0 {
		start = 0
	}
	if end > len(response) {
		end = len(response)
	}
	if start > end {
		start = end
	}

	// 按序号排序
	sort.Slice(response, func(i, j int) bool {
		return response[i].SerialNum < response[j].SerialNum
	})

	// 访问切片
	resp := response[start:end]
	totalSumPage := math.Ceil(float64(len(response)) / float64(data.PageSize))

	utils.JsonSuccessResponse(c, gin.H{
		"statistics":     resp,
		"total":          len(answersheets),
		"total_sum_page": totalSumPage,
	})
}

type getQuestionPreData struct {
	Type string `form:"type"`
}

// GetQuestionPre 获取预先信息
func GetQuestionPre(c *gin.Context) {
	var data getQuestionPreData
	if err := c.ShouldBindQuery(&data); err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}

	user, err := service.GetUserSession(c)
	if err != nil {
		code.AbortWithException(c, code.NotLogin, err)
		return
	}

	if (user.AdminType != 2) && (user.AdminType != 1) {
		code.AbortWithException(c, code.NoPermission, errors.New(user.Username+"无权限"))
		return
	}

	// 获取预先信息
	value, err := service.GetQuestionPre(data.Type)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	utils.JsonSuccessResponse(c, gin.H{
		"value": value,
	})
}

type createQuestionPreData struct {
	Type  string   `json:"type"`
	Value []string `json:"value"`
}

// CreateQuestionPre 创建预先信息
func CreateQuestionPre(c *gin.Context) {
	var data createQuestionPreData
	if err := c.ShouldBindJSON(&data); err != nil {
		code.AbortWithException(c, code.ParamError, err)
		return
	}

	user, err := service.GetUserSession(c)
	if err != nil {
		code.AbortWithException(c, code.NotLogin, err)
		return
	}

	if (user.AdminType != 2) && (user.AdminType != 1) {
		code.AbortWithException(c, code.NoPermission, errors.New(user.Username+"无权限"))
		return
	}

	// 创建预先信息
	err = service.CreateQuestionPre(data.Type, data.Value)
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	utils.JsonSuccessResponse(c, nil)
}

func ensureMap(m map[int]map[int]int, key int) map[int]int {
	if m[key] == nil {
		m[key] = make(map[int]int)
	}
	return m[key]
}
