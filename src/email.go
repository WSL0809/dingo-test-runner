package main

import (
	"fmt"
	"html"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

// EmailConfig 邮件配置结构体
type EmailConfig struct {
	Enable    bool     // 是否启用邮件发送
	SMTPHost  string   // SMTP服务器地址
	SMTPPort  int      // SMTP服务器端口
	Username  string   // 发件人邮箱
	Password  string   // 发件人邮箱密码/授权码
	From      string   // 发件人名称
	To        []string // 收件人列表
	EnableTLS bool     // 是否启用TLS
}

// TestResult 测试结果结构体
type TestResult struct {
	StartTime   time.Time
	EndTime     time.Time
	TotalTests  int
	PassedTests int
	FailedTests int
	Duration    time.Duration
	Errors      []error
	TestDetails []TestCaseResult
}

// TestCaseResult 单个测试用例结果
type TestCaseResult struct {
	Name     string
	Status   string // "passed" 或 "failed"
	Duration time.Duration
	Error    string
}

// SendEmailNotification 发送测试结果邮件通知
func SendEmailNotification(config EmailConfig, result TestResult) error {
	if !config.Enable {
		log.Debug("邮件发送功能已禁用，跳过发送")
		return nil
	}

	// 验证邮件配置
	if err := validateEmailConfig(config); err != nil {
		return fmt.Errorf("邮件配置验证失败: %v", err)
	}

	// 创建邮件消息
	m := gomail.NewMessage()

	// 设置邮件头
	m.SetHeader("From", fmt.Sprintf("%s <%s>", config.From, config.Username))
	m.SetHeader("To", config.To...)
	m.SetHeader("Subject", generateEmailSubject(result))

	// 设置邮件正文
	htmlBody := generateEmailBody(result)
	m.SetBody("text/html", htmlBody)

	// 添加纯文本版本作为备选
	textBody := generateTextEmailBody(result)
	m.AddAlternative("text/plain", textBody)

	// 创建SMTP拨号器
	d := gomail.NewDialer(config.SMTPHost, config.SMTPPort, config.Username, config.Password)

	// 配置TLS
	if !config.EnableTLS {
		d.TLSConfig = nil
	}

	// 发送邮件
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("邮件发送失败: %v", err)
	}

	log.Infof("测试结果邮件已成功发送到: %s", strings.Join(config.To, ", "))
	return nil
}

// validateEmailConfig 验证邮件配置
func validateEmailConfig(config EmailConfig) error {
	if config.SMTPHost == "" {
		return fmt.Errorf("SMTP服务器地址不能为空")
	}
	if config.SMTPPort <= 0 || config.SMTPPort > 65535 {
		return fmt.Errorf("SMTP端口必须在1-65535之间")
	}
	if config.Username == "" {
		return fmt.Errorf("发件人邮箱不能为空")
	}
	if config.Password == "" {
		return fmt.Errorf("邮箱密码/授权码不能为空")
	}
	if len(config.To) == 0 {
		return fmt.Errorf("收件人列表不能为空")
	}
	for _, email := range config.To {
		if !isValidEmail(email) {
			return fmt.Errorf("无效的邮箱地址: %s", email)
		}
	}
	return nil
}

// isValidEmail 邮箱格式验证
func isValidEmail(email string) bool {
	if email == "" {
		return false
	}

	// 基本的邮箱格式检查
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	localPart := parts[0]
	domainPart := parts[1]

	// 检查本地部分不能为空
	if localPart == "" {
		return false
	}

	// 检查域名部分不能为空且必须包含点
	if domainPart == "" || !strings.Contains(domainPart, ".") {
		return false
	}

	// 检查域名不能以点开头或结尾
	if strings.HasPrefix(domainPart, ".") || strings.HasSuffix(domainPart, ".") {
		return false
	}

	// 检查不能有连续的@符号
	if strings.Contains(email, "@@") {
		return false
	}

	return true
}

// generateEmailSubject 生成邮件主题
func generateEmailSubject(result TestResult) string {
	status := "通过"
	if result.FailedTests > 0 {
		status = "失败"
	}

	return fmt.Sprintf("数据库测试结果通知 - %s (%d/%d)",
		status, result.PassedTests, result.TotalTests)
}

// generateEmailBody 生成HTML格式的邮件正文
func generateEmailBody(result TestResult) string {
	status := "通过"
	statusColor := "#28a745" // 绿色
	if result.FailedTests > 0 {
		status = "失败"
		statusColor = "#dc3545" // 红色
	}

	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>数据库测试结果通知</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background-color: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 30px; padding: 20px; background-color: %s; color: white; border-radius: 5px; }
        .summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin-bottom: 30px; }
        .stat-card { background-color: #f8f9fa; padding: 20px; border-radius: 5px; text-align: center; border-left: 4px solid #007bff; }
        .stat-number { font-size: 2em; font-weight: bold; color: #333; }
        .stat-label { color: #666; margin-top: 5px; }
        .test-details { margin-top: 30px; }
        .test-case { margin: 10px 0; padding: 15px; border-radius: 5px; border-left: 4px solid #ddd; }
        .test-passed { background-color: #d4edda; border-left-color: #28a745; }
        .test-failed { background-color: #f8d7da; border-left-color: #dc3545; }
        .error-details { margin-top: 15px; padding: 15px; background-color: #f8f9fa; border-radius: 5px; }
        .footer { margin-top: 30px; text-align: center; color: #666; font-size: 0.9em; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>数据库测试结果通知</h1>
            <h2>整体状态: %s</h2>
        </div>
        
        <div class="summary">
            <div class="stat-card">
                <div class="stat-number">%d</div>
                <div class="stat-label">总测试数</div>
            </div>
            <div class="stat-card" style="border-left-color: #28a745;">
                <div class="stat-number" style="color: #28a745;">%d</div>
                <div class="stat-label">通过测试</div>
            </div>
            <div class="stat-card" style="border-left-color: #dc3545;">
                <div class="stat-number" style="color: #dc3545;">%d</div>
                <div class="stat-label">失败测试</div>
            </div>
            <div class="stat-card" style="border-left-color: #ffc107;">
                <div class="stat-number" style="color: #ffc107;">%.2fs</div>
                <div class="stat-label">执行时长</div>
            </div>
        </div>
        
        <div style="background-color: #e9ecef; padding: 15px; border-radius: 5px; margin-bottom: 20px;">
            <strong>执行时间:</strong> %s 至 %s<br>
            <strong>总耗时:</strong> %s
        </div>`,
		statusColor, status, result.TotalTests, result.PassedTests, result.FailedTests,
		result.Duration.Seconds(),
		result.StartTime.Format("2006-01-02 15:04:05"),
		result.EndTime.Format("2006-01-02 15:04:05"),
		result.Duration.String())

	// 添加错误详情
	if len(result.Errors) > 0 {
		htmlContent += `
        <div class="error-details">
            <h3 style="color: #dc3545; margin-top: 0;">错误详情:</h3>
            <ul>`
		for i, err := range result.Errors {
			if i >= 10 { // 最多显示10个错误
				htmlContent += fmt.Sprintf("<li>... 还有 %d 个错误未显示</li>", len(result.Errors)-10)
				break
			}
			htmlContent += fmt.Sprintf("<li style=\"margin: 5px 0;\">%s</li>", html.EscapeString(err.Error()))
		}
		htmlContent += `
            </ul>
        </div>`
	}

	// 添加测试用例详情（如果有）
	if len(result.TestDetails) > 0 {
		htmlContent += `
        <div class="test-details">
            <h3>测试用例详情:</h3>`

		for i, testCase := range result.TestDetails {
			if i >= 20 { // 最多显示20个测试用例
				htmlContent += fmt.Sprintf("<div class=\"test-case\">... 还有 %d 个测试用例未显示</div>", len(result.TestDetails)-20)
				break
			}

			cssClass := "test-passed"
			if testCase.Status == "failed" {
				cssClass = "test-failed"
			}

			htmlContent += fmt.Sprintf(`
            <div class="test-case %s">
                <strong>%s</strong> - %s (%.3fs)`,
				cssClass, html.EscapeString(testCase.Name), testCase.Status, testCase.Duration.Seconds())

			if testCase.Error != "" {
				htmlContent += fmt.Sprintf("<br><small style=\"color: #dc3545;\">错误: %s</small>", html.EscapeString(testCase.Error))
			}

			htmlContent += "</div>"
		}
		htmlContent += "</div>"
	}

	htmlContent += `
        <div class="footer">
            <p>此邮件由 MySQL Tester 自动发送</p>
            <p>发送时间: ` + time.Now().Format("2006-01-02 15:04:05") + `</p>
        </div>
    </div>
</body>
</html>`

	return htmlContent
}

// generateTextEmailBody 生成纯文本格式的邮件正文
func generateTextEmailBody(result TestResult) string {
	status := "通过"
	if result.FailedTests > 0 {
		status = "失败"
	}

	text := fmt.Sprintf(`数据库测试结果通知

整体状态: %s
总测试数: %d
通过测试: %d
失败测试: %d
执行时长: %.2fs

执行时间: %s 至 %s
总耗时: %s

`, status, result.TotalTests, result.PassedTests, result.FailedTests,
		result.Duration.Seconds(),
		result.StartTime.Format("2006-01-02 15:04:05"),
		result.EndTime.Format("2006-01-02 15:04:05"),
		result.Duration.String())

	if len(result.Errors) > 0 {
		text += "错误详情:\n"
		for i, err := range result.Errors {
			if i >= 10 {
				text += fmt.Sprintf("... 还有 %d 个错误未显示\n", len(result.Errors)-10)
				break
			}
			text += fmt.Sprintf("- %s\n", err.Error())
		}
		text += "\n"
	}

	if len(result.TestDetails) > 0 {
		text += "测试用例详情:\n"
		for i, testCase := range result.TestDetails {
			if i >= 20 {
				text += fmt.Sprintf("... 还有 %d 个测试用例未显示\n", len(result.TestDetails)-20)
				break
			}
			text += fmt.Sprintf("- %s: %s (%.3fs)", testCase.Name, testCase.Status, testCase.Duration.Seconds())
			if testCase.Error != "" {
				text += fmt.Sprintf(" - 错误: %s", testCase.Error)
			}
			text += "\n"
		}
	}

	text += fmt.Sprintf("\n此邮件由 MySQL Tester 自动发送\n发送时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))

	return text
}

// parseEmailList 解析邮箱列表字符串
func parseEmailList(emailStr string) []string {
	if emailStr == "" {
		return nil
	}

	emails := strings.Split(emailStr, ",")
	var result []string

	for _, email := range emails {
		email = strings.TrimSpace(email)
		if email != "" {
			result = append(result, email)
		}
	}

	// 如果没有有效的邮箱地址，返回空slice而不是nil
	if len(result) == 0 {
		return []string{}
	}

	return result
}

// parseEmailConfig 从命令行参数解析邮件配置
func parseEmailConfig() EmailConfig {
	config := EmailConfig{
		Enable:    emailEnable,
		SMTPHost:  emailSMTPHost,
		SMTPPort:  emailSMTPPort,
		Username:  emailUsername,
		Password:  emailPassword,
		From:      emailFrom,
		To:        parseEmailList(emailTo),
		EnableTLS: emailEnableTLS,
	}

	// 如果没有设置发件人名称，使用邮箱地址
	if config.From == "" {
		config.From = "MySQL Tester"
	}

	return config
}
