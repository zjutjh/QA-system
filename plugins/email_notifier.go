package plugins

import (
	"QA-System/internal/global/config"
	"QA-System/internal/pkg/extension"
	"QA-System/internal/pkg/redis"
	"bytes"
	"context"
	"fmt"
	"net/smtp"
	"text/template"
	"time"

	redisv8 "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type EmailNotifier struct {
	smtpHost     string
	smtpPort     int
	smtpUsername string
	smtpPassword string
	from         string
	mailTemplate *template.Template
	consumerName string
	streamName   string
	groupName    string
}

const emailTemplateText = `Subject: 您的问卷"{{.title}}"收到了新回复

您的问卷"{{.title}}"收到了新回复，请及时查收。`

func init() {
	notifier := &EmailNotifier{
		consumerName: "email_notifier",
	}
	if err := notifier.initialize(); err != nil {
		panic(fmt.Sprintf("Failed to initialize email_notifier: %v", err))
	}
	extension.RegisterPlugin(notifier)
}

func (p *EmailNotifier) initialize() error {
	// 读取SMTP配置
	p.smtpHost = config.Config.GetString("email_notifier.smtp.host")
	p.smtpPort = config.Config.GetInt("email_notifier.smtp.port")
	p.smtpUsername = config.Config.GetString("email_notifier.smtp.username")
	p.smtpPassword = config.Config.GetString("email_notifier.smtp.password")
	p.from = config.Config.GetString("email_notifier.smtp.from")

	// 读取Stream配置
	p.streamName = config.Config.GetString("redis.stream.name")
	p.groupName = config.Config.GetString("redis.stream.group")

	// 初始化邮件模板
	tpl, err := template.New("email").Parse(emailTemplateText)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %v", err)
	}
	p.mailTemplate = tpl

	return nil
}

func (p *EmailNotifier) GetMetadata() extension.PluginMetadata {
	return extension.PluginMetadata{
		Name:        "email_notifier",
		Version:     "0.0.1",
		Author:      "System",
		Description: "Send email notifications for new survey responses",
	}
}

// Execute 作为插件的服务入口，运行消息处理循环
func (p *EmailNotifier) Execute() error {
	ctx := context.Background()
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	zap.L().Info("Starting email notifier service",
		zap.String("consumer", p.consumerName),
		zap.String("stream", p.streamName))

	for {
		select {
		case <-ticker.C:
			if err := p.processMessages(ctx); err != nil {
				zap.L().Error("Failed to process messages", zap.Error(err))
			}
		}
	}
}

func (p *EmailNotifier) processMessages(ctx context.Context) error {
	streams, err := redis.ConsumeFromStream(ctx, p.consumerName)
	if err != nil {
		return fmt.Errorf("failed to read from stream: %v", err)
	}

	for _, stream := range streams {
		for _, message := range stream.Messages {
			if err := p.handleMessage(ctx, message); err != nil {
				zap.L().Error("Failed to handle message",
					zap.String("messageID", message.ID),
					zap.Error(err))
				continue
			}

			// 确认消息已处理
			if err := redis.AckMessage(ctx, message.ID); err != nil {
				zap.L().Error("Failed to ack message",
					zap.String("messageID", message.ID),
					zap.Error(err))
			}
		}
	}
	return nil
}

func (p *EmailNotifier) handleMessage(ctx context.Context, message redisv8.XMessage) error {
	// 从消息中提取数据
	title, ok := message.Values["survey_title"].(string)
	if !ok {
		return fmt.Errorf("invalid survey_title in message")
	}

	recipient, ok := message.Values["creator_email"].(string)
	if !ok {
		return fmt.Errorf("invalid creator_email in message")
	}

	// 准备邮件数据
	data := map[string]interface{}{
		"title": title,
	}

	return p.sendEmail(recipient, data)
}

func (p *EmailNotifier) sendEmail(recipient string, data map[string]interface{}) error {
	var body bytes.Buffer
	if err := p.mailTemplate.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to render email template: %v", err)
	}

	auth := smtp.PlainAuth("", p.smtpUsername, p.smtpPassword, p.smtpHost)
	addr := fmt.Sprintf("%s:%d", p.smtpHost, p.smtpPort)

	err := smtp.SendMail(
		addr,
		auth,
		p.from,
		[]string{recipient},
		body.Bytes(),
	)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	zap.L().Info("Email sent successfully",
		zap.String("recipient", recipient),
		zap.String("title", data["title"].(string)))
	return nil
}
