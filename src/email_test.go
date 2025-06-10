// Copyright 2020 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateEmailConfig 测试邮件配置验证功能
func TestValidateEmailConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      EmailConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "有效配置",
			config: EmailConfig{
				SMTPHost: "smtp.example.com",
				SMTPPort: 587,
				Username: "test@example.com",
				Password: "password123",
				To:       []string{"user@example.com"},
			},
			expectError: false,
		},
		{
			name: "SMTP主机为空",
			config: EmailConfig{
				SMTPHost: "",
				SMTPPort: 587,
				Username: "test@example.com",
				Password: "password123",
				To:       []string{"user@example.com"},
			},
			expectError: true,
			errorMsg:    "SMTP服务器地址不能为空",
		},
		{
			name: "SMTP端口无效 - 小于1",
			config: EmailConfig{
				SMTPHost: "smtp.example.com",
				SMTPPort: 0,
				Username: "test@example.com",
				Password: "password123",
				To:       []string{"user@example.com"},
			},
			expectError: true,
			errorMsg:    "SMTP端口必须在1-65535之间",
		},
		{
			name: "SMTP端口无效 - 大于65535",
			config: EmailConfig{
				SMTPHost: "smtp.example.com",
				SMTPPort: 65536,
				Username: "test@example.com",
				Password: "password123",
				To:       []string{"user@example.com"},
			},
			expectError: true,
			errorMsg:    "SMTP端口必须在1-65535之间",
		},
		{
			name: "用户名为空",
			config: EmailConfig{
				SMTPHost: "smtp.example.com",
				SMTPPort: 587,
				Username: "",
				Password: "password123",
				To:       []string{"user@example.com"},
			},
			expectError: true,
			errorMsg:    "发件人邮箱不能为空",
		},
		{
			name: "密码为空",
			config: EmailConfig{
				SMTPHost: "smtp.example.com",
				SMTPPort: 587,
				Username: "test@example.com",
				Password: "",
				To:       []string{"user@example.com"},
			},
			expectError: true,
			errorMsg:    "邮箱密码/授权码不能为空",
		},
		{
			name: "收件人列表为空",
			config: EmailConfig{
				SMTPHost: "smtp.example.com",
				SMTPPort: 587,
				Username: "test@example.com",
				Password: "password123",
				To:       []string{},
			},
			expectError: true,
			errorMsg:    "收件人列表不能为空",
		},
		{
			name: "无效的邮箱地址",
			config: EmailConfig{
				SMTPHost: "smtp.example.com",
				SMTPPort: 587,
				Username: "test@example.com",
				Password: "password123",
				To:       []string{"invalid-email"},
			},
			expectError: true,
			errorMsg:    "无效的邮箱地址: invalid-email",
		},
		{
			name: "多个收件人其中一个无效",
			config: EmailConfig{
				SMTPHost: "smtp.example.com",
				SMTPPort: 587,
				Username: "test@example.com",
				Password: "password123",
				To:       []string{"valid@example.com", "invalid-email", "another@example.com"},
			},
			expectError: true,
			errorMsg:    "无效的邮箱地址: invalid-email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEmailConfig(tt.config)
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestIsValidEmail 测试邮箱格式验证功能
func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email    string
		expected bool
	}{
		{"test@example.com", true},
		{"user.name@domain.co.uk", true},
		{"admin@localhost.localdomain", true},
		{"test123@test-domain.com", true},
		{"invalid-email", false},
		{"@example.com", false},
		{"test@", false},
		{"", false},
		{"test.example.com", false},
		{"test@@example.com", false},
		{"test@.com", false},
		{"test@com", false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("email_%s", tt.email), func(t *testing.T) {
			result := isValidEmail(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGenerateEmailSubject 测试邮件主题生成功能
func TestGenerateEmailSubject(t *testing.T) {
	tests := []struct {
		name            string
		result          TestResult
		expectedSubject string
	}{
		{
			name: "所有测试通过",
			result: TestResult{
				TotalTests:  10,
				PassedTests: 10,
				FailedTests: 0,
			},
			expectedSubject: "数据库测试结果通知 - 通过 (10/10)",
		},
		{
			name: "部分测试失败",
			result: TestResult{
				TotalTests:  10,
				PassedTests: 7,
				FailedTests: 3,
			},
			expectedSubject: "数据库测试结果通知 - 失败 (7/10)",
		},
		{
			name: "所有测试失败",
			result: TestResult{
				TotalTests:  5,
				PassedTests: 0,
				FailedTests: 5,
			},
			expectedSubject: "数据库测试结果通知 - 失败 (0/5)",
		},
		{
			name: "零测试情况",
			result: TestResult{
				TotalTests:  0,
				PassedTests: 0,
				FailedTests: 0,
			},
			expectedSubject: "数据库测试结果通知 - 通过 (0/0)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subject := generateEmailSubject(tt.result)
			assert.Equal(t, tt.expectedSubject, subject)
		})
	}
}

// TestGenerateEmailBody 测试HTML邮件正文生成功能
func TestGenerateEmailBody(t *testing.T) {
	startTime := time.Date(2023, 12, 25, 10, 0, 0, 0, time.UTC)
	endTime := startTime.Add(30 * time.Second)

	tests := []struct {
		name            string
		result          TestResult
		expectedContain []string
		notContain      []string
	}{
		{
			name: "成功测试结果",
			result: TestResult{
				StartTime:   startTime,
				EndTime:     endTime,
				TotalTests:  5,
				PassedTests: 5,
				FailedTests: 0,
				Duration:    30 * time.Second,
				Errors:      []error{},
				TestDetails: []TestCaseResult{
					{Name: "test1", Status: "passed", Duration: 5 * time.Second, Error: ""},
					{Name: "test2", Status: "passed", Duration: 10 * time.Second, Error: ""},
				},
			},
			expectedContain: []string{
				"数据库测试结果通知",
				"整体状态: 通过",
				"总测试数</div>",
				"<div class=\"stat-number\">5</div>",
				"<div class=\"stat-number\" style=\"color: #28a745;\">5</div>",
				"<div class=\"stat-number\" style=\"color: #dc3545;\">0</div>",
				"30.00s",
				"2023-12-25 10:00:00",
				"2023-12-25 10:00:30",
				"30s",
				"test1</strong> - passed",
				"test2</strong> - passed",
				"test-passed",
			},
			notContain: []string{
				"错误详情",
				"class=\"test-case test-failed\"",
			},
		},
		{
			name: "失败测试结果",
			result: TestResult{
				StartTime:   startTime,
				EndTime:     endTime,
				TotalTests:  3,
				PassedTests: 1,
				FailedTests: 2,
				Duration:    45 * time.Second,
				Errors: []error{
					errors.New("test error 1"),
					errors.New("test error 2"),
				},
				TestDetails: []TestCaseResult{
					{Name: "test1", Status: "passed", Duration: 5 * time.Second, Error: ""},
					{Name: "test2", Status: "failed", Duration: 10 * time.Second, Error: "connection failed"},
					{Name: "test3", Status: "failed", Duration: 5 * time.Second, Error: "timeout"},
				},
			},
			expectedContain: []string{
				"整体状态: 失败",
				"<div class=\"stat-number\">3</div>",
				"<div class=\"stat-number\" style=\"color: #28a745;\">1</div>",
				"<div class=\"stat-number\" style=\"color: #dc3545;\">2</div>",
				"45.00s",
				"错误详情",
				"test error 1",
				"test error 2",
				"test1</strong> - passed",
				"test2</strong> - failed",
				"test3</strong> - failed",
				"test-passed",
				"test-failed",
				"connection failed",
				"timeout",
			},
			notContain: []string{},
		},
		{
			name: "大量错误测试（测试截断功能）",
			result: TestResult{
				StartTime:   startTime,
				EndTime:     endTime,
				TotalTests:  15,
				PassedTests: 0,
				FailedTests: 15,
				Duration:    60 * time.Second,
				Errors: func() []error {
					var errs []error
					for i := 1; i <= 15; i++ {
						errs = append(errs, fmt.Errorf("error %d", i))
					}
					return errs
				}(),
				TestDetails: func() []TestCaseResult {
					var details []TestCaseResult
					for i := 1; i <= 25; i++ {
						details = append(details, TestCaseResult{
							Name:     fmt.Sprintf("test%d", i),
							Status:   "failed",
							Duration: time.Second,
							Error:    fmt.Sprintf("error in test %d", i),
						})
					}
					return details
				}(),
			},
			expectedContain: []string{
				"error 1",
				"error 10",
				"还有 5 个错误未显示", // 只显示前10个错误
				"test1</strong> - failed",
				"test20</strong> - failed",
				"还有 5 个测试用例未显示", // 只显示前20个测试用例
			},
			notContain: []string{
				"error 11", // 第11个错误应该被截断
				"test21",   // 第21个测试用例应该被截断
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := generateEmailBody(tt.result)

			// 检查期望包含的内容
			for _, expected := range tt.expectedContain {
				assert.Contains(t, body, expected, "邮件正文应该包含: %s", expected)
			}

			// 检查不应该包含的内容
			for _, notExpected := range tt.notContain {
				assert.NotContains(t, body, notExpected, "邮件正文不应该包含: %s", notExpected)
			}

			// 验证HTML结构的基本要素
			assert.Contains(t, body, "<!DOCTYPE html>")
			assert.Contains(t, body, "<html>")
			assert.Contains(t, body, "</html>")
			assert.Contains(t, body, "<head>")
			assert.Contains(t, body, "<body>")
			assert.Contains(t, body, "此邮件由 MySQL Tester 自动发送")
		})
	}
}

// TestGenerateTextEmailBody 测试文本邮件正文生成功能
func TestGenerateTextEmailBody(t *testing.T) {
	startTime := time.Date(2023, 12, 25, 10, 0, 0, 0, time.UTC)
	endTime := startTime.Add(30 * time.Second)

	tests := []struct {
		name            string
		result          TestResult
		expectedContain []string
		notContain      []string
	}{
		{
			name: "成功测试结果",
			result: TestResult{
				StartTime:   startTime,
				EndTime:     endTime,
				TotalTests:  3,
				PassedTests: 3,
				FailedTests: 0,
				Duration:    30 * time.Second,
				Errors:      []error{},
				TestDetails: []TestCaseResult{
					{Name: "test1", Status: "passed", Duration: 5 * time.Second, Error: ""},
					{Name: "test2", Status: "passed", Duration: 10 * time.Second, Error: ""},
				},
			},
			expectedContain: []string{
				"数据库测试结果通知",
				"整体状态: 通过",
				"总测试数: 3",
				"通过测试: 3",
				"失败测试: 0",
				"执行时长: 30.00s",
				"执行时间: 2023-12-25 10:00:00 至 2023-12-25 10:00:30",
				"总耗时: 30s",
				"测试用例详情:",
				"- test1: passed (5.000s)",
				"- test2: passed (10.000s)",
				"此邮件由 MySQL Tester 自动发送",
			},
			notContain: []string{
				"错误详情:",
				"错误:",
			},
		},
		{
			name: "失败测试结果",
			result: TestResult{
				StartTime:   startTime,
				EndTime:     endTime,
				TotalTests:  3,
				PassedTests: 1,
				FailedTests: 2,
				Duration:    45 * time.Second,
				Errors: []error{
					errors.New("database connection failed"),
					errors.New("query timeout"),
				},
				TestDetails: []TestCaseResult{
					{Name: "test1", Status: "passed", Duration: 5 * time.Second, Error: ""},
					{Name: "test2", Status: "failed", Duration: 10 * time.Second, Error: "connection error"},
					{Name: "test3", Status: "failed", Duration: 5 * time.Second, Error: "timeout"},
				},
			},
			expectedContain: []string{
				"整体状态: 失败",
				"总测试数: 3",
				"通过测试: 1",
				"失败测试: 2",
				"错误详情:",
				"- database connection failed",
				"- query timeout",
				"测试用例详情:",
				"- test1: passed (5.000s)",
				"- test2: failed (10.000s) - 错误: connection error",
				"- test3: failed (5.000s) - 错误: timeout",
			},
			notContain: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := generateTextEmailBody(tt.result)

			// 检查期望包含的内容
			for _, expected := range tt.expectedContain {
				assert.Contains(t, body, expected, "文本邮件正文应该包含: %s", expected)
			}

			// 检查不应该包含的内容
			for _, notExpected := range tt.notContain {
				assert.NotContains(t, body, notExpected, "文本邮件正文不应该包含: %s", notExpected)
			}
		})
	}
}

// TestParseEmailList 测试邮箱列表解析功能
func TestParseEmailList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "单个邮箱",
			input:    "test@example.com",
			expected: []string{"test@example.com"},
		},
		{
			name:     "多个邮箱",
			input:    "test1@example.com,test2@example.com,test3@example.com",
			expected: []string{"test1@example.com", "test2@example.com", "test3@example.com"},
		},
		{
			name:     "带空格的邮箱列表",
			input:    "test1@example.com, test2@example.com , test3@example.com",
			expected: []string{"test1@example.com", "test2@example.com", "test3@example.com"},
		},
		{
			name:     "空字符串",
			input:    "",
			expected: nil,
		},
		{
			name:     "只有逗号",
			input:    ",,,",
			expected: []string{},
		},
		{
			name:     "包含空项的列表",
			input:    "test1@example.com,,test2@example.com, ,test3@example.com",
			expected: []string{"test1@example.com", "test2@example.com", "test3@example.com"},
		},
		{
			name:     "只有空格",
			input:    "   ",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseEmailList(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestParseEmailConfig 测试邮件配置解析功能
func TestParseEmailConfig(t *testing.T) {
	// 保存原始值
	originalEnable := emailEnable
	originalSMTPHost := emailSMTPHost
	originalSMTPPort := emailSMTPPort
	originalUsername := emailUsername
	originalPassword := emailPassword
	originalFrom := emailFrom
	originalTo := emailTo
	originalEnableTLS := emailEnableTLS

	// 恢复原始值
	defer func() {
		emailEnable = originalEnable
		emailSMTPHost = originalSMTPHost
		emailSMTPPort = originalSMTPPort
		emailUsername = originalUsername
		emailPassword = originalPassword
		emailFrom = originalFrom
		emailTo = originalTo
		emailEnableTLS = originalEnableTLS
	}()

	tests := []struct {
		name     string
		setup    func()
		expected EmailConfig
	}{
		{
			name: "基本配置",
			setup: func() {
				emailEnable = true
				emailSMTPHost = "smtp.example.com"
				emailSMTPPort = 587
				emailUsername = "test@example.com"
				emailPassword = "password123"
				emailFrom = "Test Sender"
				emailTo = "user1@example.com,user2@example.com"
				emailEnableTLS = true
			},
			expected: EmailConfig{
				Enable:    true,
				SMTPHost:  "smtp.example.com",
				SMTPPort:  587,
				Username:  "test@example.com",
				Password:  "password123",
				From:      "Test Sender",
				To:        []string{"user1@example.com", "user2@example.com"},
				EnableTLS: true,
			},
		},
		{
			name: "空的发件人名称使用默认值",
			setup: func() {
				emailEnable = false
				emailSMTPHost = "smtp.test.com"
				emailSMTPPort = 25
				emailUsername = "sender@test.com"
				emailPassword = "pass"
				emailFrom = ""
				emailTo = "recipient@test.com"
				emailEnableTLS = false
			},
			expected: EmailConfig{
				Enable:    false,
				SMTPHost:  "smtp.test.com",
				SMTPPort:  25,
				Username:  "sender@test.com",
				Password:  "pass",
				From:      "MySQL Tester", // 默认值
				To:        []string{"recipient@test.com"},
				EnableTLS: false,
			},
		},
		{
			name: "空的收件人列表",
			setup: func() {
				emailEnable = true
				emailSMTPHost = "smtp.test.com"
				emailSMTPPort = 465
				emailUsername = "sender@test.com"
				emailPassword = "pass"
				emailFrom = "Sender"
				emailTo = ""
				emailEnableTLS = true
			},
			expected: EmailConfig{
				Enable:    true,
				SMTPHost:  "smtp.test.com",
				SMTPPort:  465,
				Username:  "sender@test.com",
				Password:  "pass",
				From:      "Sender",
				To:        nil,
				EnableTLS: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			config := parseEmailConfig()
			assert.Equal(t, tt.expected, config)
		})
	}
}

// TestSendEmailNotificationDisabled 测试邮件发送功能在禁用时的行为
func TestSendEmailNotificationDisabled(t *testing.T) {
	config := EmailConfig{
		Enable: false,
	}

	result := TestResult{
		TotalTests:  1,
		PassedTests: 1,
		FailedTests: 0,
	}

	err := SendEmailNotification(config, result)
	assert.NoError(t, err, "禁用邮件发送时不应该返回错误")
}

// TestSendEmailNotificationInvalidConfig 测试邮件发送功能在配置无效时的行为
func TestSendEmailNotificationInvalidConfig(t *testing.T) {
	config := EmailConfig{
		Enable:   true,
		SMTPHost: "", // 无效配置
	}

	result := TestResult{
		TotalTests:  1,
		PassedTests: 1,
		FailedTests: 0,
	}

	err := SendEmailNotification(config, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "邮件配置验证失败")
}

// TestEmailConfigEdgeCases 测试邮件配置的边界情况
func TestEmailConfigEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		config      EmailConfig
		expectError bool
		description string
	}{
		{
			name: "最小有效端口",
			config: EmailConfig{
				SMTPHost: "smtp.example.com",
				SMTPPort: 1,
				Username: "test@example.com",
				Password: "password",
				To:       []string{"user@example.com"},
			},
			expectError: false,
			description: "端口1应该是有效的",
		},
		{
			name: "最大有效端口",
			config: EmailConfig{
				SMTPHost: "smtp.example.com",
				SMTPPort: 65535,
				Username: "test@example.com",
				Password: "password",
				To:       []string{"user@example.com"},
			},
			expectError: false,
			description: "端口65535应该是有效的",
		},
		{
			name: "复杂的有效邮箱地址",
			config: EmailConfig{
				SMTPHost: "smtp.example.com",
				SMTPPort: 587,
				Username: "test@example.com",
				Password: "password",
				To:       []string{"user.name+tag@sub.domain.co.uk"},
			},
			expectError: false,
			description: "复杂但有效的邮箱地址应该通过验证",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEmailConfig(tt.config)
			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

// BenchmarkGenerateEmailBody 性能测试：邮件正文生成
func BenchmarkGenerateEmailBody(b *testing.B) {
	result := TestResult{
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(30 * time.Second),
		TotalTests:  100,
		PassedTests: 95,
		FailedTests: 5,
		Duration:    30 * time.Second,
		Errors: []error{
			errors.New("error 1"),
			errors.New("error 2"),
		},
		TestDetails: func() []TestCaseResult {
			var details []TestCaseResult
			for i := 0; i < 100; i++ {
				status := "passed"
				errorMsg := ""
				if i < 5 {
					status = "failed"
					errorMsg = fmt.Sprintf("error in test %d", i+1)
				}
				details = append(details, TestCaseResult{
					Name:     fmt.Sprintf("test%d", i+1),
					Status:   status,
					Duration: time.Second,
					Error:    errorMsg,
				})
			}
			return details
		}(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateEmailBody(result)
	}
}

// BenchmarkGenerateTextEmailBody 性能测试：文本邮件正文生成
func BenchmarkGenerateTextEmailBody(b *testing.B) {
	result := TestResult{
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(30 * time.Second),
		TotalTests:  100,
		PassedTests: 95,
		FailedTests: 5,
		Duration:    30 * time.Second,
		Errors: []error{
			errors.New("error 1"),
			errors.New("error 2"),
		},
		TestDetails: func() []TestCaseResult {
			var details []TestCaseResult
			for i := 0; i < 100; i++ {
				details = append(details, TestCaseResult{
					Name:     fmt.Sprintf("test%d", i+1),
					Status:   "passed",
					Duration: time.Second,
					Error:    "",
				})
			}
			return details
		}(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateTextEmailBody(result)
	}
}

// TestEmailBodyHTMLSafety 测试HTML邮件正文的安全性
func TestEmailBodyHTMLSafety(t *testing.T) {
	result := TestResult{
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(30 * time.Second),
		TotalTests:  2,
		PassedTests: 1,
		FailedTests: 1,
		Duration:    30 * time.Second,
		Errors: []error{
			errors.New("<script>alert('xss')</script>"),
		},
		TestDetails: []TestCaseResult{
			{
				Name:     "test<script>",
				Status:   "failed",
				Duration: time.Second,
				Error:    "<img src=x onerror=alert('xss')>",
			},
		},
	}

	body := generateEmailBody(result)

	// 确保潜在的XSS内容被适当处理
	// 验证HTML转义是否正常工作
	assert.Contains(t, body, "test&lt;script&gt;", "HTML标签应该被转义")
	assert.Contains(t, body, "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;", "HTML标签应该被转义")
	assert.Contains(t, body, "&lt;img src=x onerror=alert(&#39;xss&#39;)&gt;", "HTML标签应该被转义")
}
