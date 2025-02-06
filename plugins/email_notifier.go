package plugins

import (
	"QA-System/internal/global/config"
	"QA-System/internal/pkg/extension"
	"bytes"
	"fmt"
	"net/smtp"
	"text/template"
)

// EmailNotifier 插件结构体
type EmailNotifier struct {
	smtpHost     string
	smtpPort     int
	smtpUsername string
	smtpPassword string
	from         string
	recipients   []string
	mailTemplate *template.Template
}

// GetMetadata 实现插件元数据
func (p *EmailNotifier) GetMetadata() extension.PluginMetadata {
	return extension.PluginMetadata{
		Name:        "email_notifier",
		Version:     "0.0.1",
		Author:      "SituChengxiang",
		Description: "Send email notification on new questionnaire response",
	}
}

// init 自动注册插件
func init() {
	notifier := &EmailNotifier{}
	// 初始化时加载配置和模板
	if err := notifier.initialize(); err != nil {
		panic(fmt.Sprintf("Failed to initialize email_notifier: %v", err))
	}
	extension.RegisterPlugin(notifier)
}

// initialize 从配置加载SMTP参数和模板
func (p *EmailNotifier) initialize() error {
	// 读取SMTP配置
	p.smtpHost = config.Config.GetString("email_notifier.smtp.host")
	p.smtpPort = config.Config.GetInt("email_notifier.smtp.port")
	p.smtpUsername = config.Config.GetString("email_notifier.smtp.username")
	p.smtpPassword = config.Config.GetString("email_notifier.smtp.password")
	p.from = config.Config.GetString("email_notifier.smtp.from")
	p.recipients = config.Config.GetStringSlice("email_notifier.recipients")

	// 解析邮件模板
	tplSubject := config.Config.GetString("email_notifier.template.subject")
	tplBody := config.Config.GetString("email_notifier.template.body")
	tpl, err := template.New("email").Parse(fmt.Sprintf("Subject: %s\n\n%s", tplSubject, tplBody))
	if err != nil {
		return fmt.Errorf("failed to parse email template: %v", err)
	}
	p.mailTemplate = tpl

	return nil
}

// Execute 执行邮件发送逻辑
func (p *EmailNotifier) Execute(params map[string]interface{}) error {
	// 从参数中获取问卷标题
	questionnaireTitle, ok := params["title"].(string)
	if !ok {
		return fmt.Errorf("missing questionnaire title in params")
	}

	// 构造模板数据
	data := map[string]interface{}{
		"QuestionnaireTitle": questionnaireTitle,
	}

	// 渲染邮件内容
	var body bytes.Buffer
	if err := p.mailTemplate.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to render email template: %v", err)
	}

	// 发送邮件
	auth := smtp.PlainAuth("", p.smtpUsername, p.smtpPassword, p.smtpHost)
	msg := body.Bytes()
	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", p.smtpHost, p.smtpPort),
		auth,
		p.from,
		p.recipients,
		msg,
	)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}
