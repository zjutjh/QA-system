package dao

import (
	"QA-System/internal/models"
	"context"
)

type Option struct {
	SerialNum   int    `json:"serial_num"`  //选项序号
	Content     string `json:"content"`     //选项内容
	Description string `json:"description"` //选项描述
	Img         string `json:"img"`         //图片
}

func (d *Dao) CreateOption(ctx context.Context, option models.Option) error {
	err := d.orm.WithContext(ctx).Create(&option).Error
	return err
}

func (d *Dao) GetOptionsByQuestionID(ctx context.Context, questionID int) ([]models.Option, error) {
	var options []models.Option
	err := d.orm.WithContext(ctx).Model(models.Option{}).Where("question_id = ?", questionID).Find(&options).Error
	return options, err
}

func (d *Dao) DeleteOption(ctx context.Context, optionID int) error {
	err := d.orm.WithContext(ctx).Where("id = ?", optionID).Delete(&models.Option{}).Error
	return err
}

func (d *Dao) GetOptionByQIDAndAnswer(ctx context.Context, qid int, answer string) (*models.Option, error) {
	var option models.Option
	err := d.orm.WithContext(ctx).Model(models.Option{}).Where("question_id = ?  AND content = ?", qid, answer).First(&option).Error
	return &option, err
}

func (d *Dao) GetOptionByQIDAndSerialNum(ctx context.Context, qid int, serialNum int) (*models.Option, error) {
	var option models.Option
	err := d.orm.WithContext(ctx).Model(models.Option{}).Where("question_id = ?  AND serial_num = ?", qid, serialNum).First(&option).Error
	return &option, err
}

func (d *Dao) GetQuestionsByIDs(ctx context.Context, qids []int) ([]models.Question, error) {
	var questions []models.Question
	err := d.orm.WithContext(ctx).Model(models.Question{}).Where("id in ?", qids).Find(&questions).Error
	return questions, err
}
