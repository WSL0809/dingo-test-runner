# MySQL Tester 邮件通知功能

## 功能概述

MySQL Tester 现已支持邮件通知功能，可在数据库测试完成后自动发送详细的测试结果报告。该功能支持HTML和纯文本两种格式的邮件，提供美观的测试结果展示和完整的错误信息。

## 核心功能特性

### 1. 智能邮件内容生成
- **HTML格式邮件**：美观的响应式布局，包含颜色编码的测试状态
- **纯文本邮件**：作为HTML邮件的备选方案，确保兼容性
- **安全性保障**：自动HTML转义，防止XSS攻击
- **内容截断**：合理控制邮件大小（最多显示10个错误，20个测试用例详情）

### 2. 详细的测试结果统计
- 测试总数、通过数、失败数
- 测试执行时间和总耗时
- 详细的错误信息列表
- 每个测试用例的执行状态和耗时

### 3. 灵活的配置选项
- 支持多种SMTP服务器（Gmail、QQ邮箱、企业邮箱等）
- 可选的TLS加密连接
- 多收件人支持
- 自定义发件人名称

## 实现架构

### 主要组件

```
email.go
├── EmailConfig          # 邮件配置结构体
├── TestResult          # 测试结果结构体  
├── TestCaseResult      # 单个测试用例结果
├── SendEmailNotification # 主邮件发送函数
├── validateEmailConfig # 配置验证
├── generateEmailBody   # HTML邮件正文生成
├── generateTextEmailBody # 纯文本邮件正文生成
└── parseEmailConfig    # 命令行参数解析
```

### 关键实现逻辑

#### 1. 配置验证机制
```go
func validateEmailConfig(config EmailConfig) error {
    // SMTP服务器地址验证
    // 端口范围检查 (1-65535)
    // 邮箱格式验证（支持复杂格式）
    // 收件人列表验证
}
```

#### 2. 邮件内容生成
- **状态色彩编码**：成功（绿色）/ 失败（红色）
- **响应式HTML设计**：适配不同邮件客户端
- **内容分层展示**：概览统计 → 错误详情 → 测试用例详情

#### 3. 安全性措施
- HTML内容自动转义，防止XSS攻击
- 邮箱格式严格验证
- SMTP连接支持TLS加密

## 配置参数说明

| 参数名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `--email-enable` | bool | false | 启用邮件通知功能 |
| `--email-smtp-host` | string | "" | SMTP服务器地址 |
| `--email-smtp-port` | int | 587 | SMTP服务器端口 |
| `--email-username` | string | "" | 发件人邮箱地址 |
| `--email-password` | string | "" | 邮箱密码或授权码 |
| `--email-from` | string | "MySQL Tester" | 发件人显示名称 |
| `--email-to` | string | "" | 收件人列表（逗号分隔） |
| `--email-enable-tls` | bool | true | 启用TLS加密连接 |

## 使用方法

### 1. 基本用法

```bash
./mysql-tester \
  --host="localhost" \
  --port="3306" \
  --user="root" \
  --passwd="password" \
  --email-enable=true \
  --email-smtp-host="smtp.qq.com" \
  --email-smtp-port=587 \
  --email-username="your_email@qq.com" \
  --email-password="your_app_password" \
  --email-to="recipient1@example.com,recipient2@example.com"
```

### 2. 常用邮箱配置

#### QQ邮箱配置
```bash
--email-smtp-host="smtp.qq.com"
--email-smtp-port=587
--email-enable-tls=true
# 注意：使用授权码而非登录密码
```

#### Gmail配置
```bash
--email-smtp-host="smtp.gmail.com"
--email-smtp-port=587
--email-enable-tls=true
# 注意：需要开启"应用专用密码"
```

#### 企业邮箱配置
```bash
--email-smtp-host="smtp.exmail.qq.com"  # 腾讯企业邮箱
--email-smtp-port=587
--email-enable-tls=true
```

### 3. 配置文件示例

参考 `email-config-example.sh` 脚本：

```bash
#!/bin/bash

# 邮件配置
EMAIL_ENABLE=true
EMAIL_SMTP_HOST="smtp.qq.com"
EMAIL_SMTP_PORT=587
EMAIL_USERNAME="your_email@qq.com"
EMAIL_PASSWORD="your_app_password"
EMAIL_TO="recipient1@example.com,recipient2@example.com"

# 运行测试
./mysql-tester \
  --host="127.0.0.1" \
  --port="3306" \
  --user="root" \
  --passwd="" \
  --email-enable="$EMAIL_ENABLE" \
  --email-smtp-host="$EMAIL_SMTP_HOST" \
  --email-smtp-port="$EMAIL_SMTP_PORT" \
  --email-username="$EMAIL_USERNAME" \
  --email-password="$EMAIL_PASSWORD" \
  --email-to="$EMAIL_TO"
```

## 邮件内容展示

### HTML邮件效果
- 🎨 **美观的卡片式布局**：测试统计以卡片形式展示
- 📊 **颜色编码状态**：绿色表示成功，红色表示失败
- 📋 **分层信息展示**：从概览到详情的清晰层次
- 📱 **响应式设计**：适配桌面和移动端邮件客户端

### 邮件内容结构
```
数据库测试结果通知
├── 整体状态（通过/失败）
├── 统计卡片
│   ├── 总测试数
│   ├── 通过测试数
│   ├── 失败测试数
│   └── 执行时长
├── 执行时间信息
├── 错误详情（如有）
├── 测试用例详情
└── 发送时间戳
```

## 质量保证

### 单元测试覆盖
- ✅ 邮件配置验证：8个测试用例
- ✅ 邮箱格式验证：12个测试用例
- ✅ 邮件内容生成：6个测试用例
- ✅ 配置解析功能：3个测试用例
- ✅ 边界情况处理：3个测试用例
- ✅ 安全性测试：XSS防护验证
- ✅ 性能基准测试：邮件生成性能测试

### 性能指标
```
BenchmarkGenerateEmailBody-8         29638    40057 ns/op
BenchmarkGenerateTextEmailBody-8    119530     9708 ns/op
```

### 安全特性
- **HTML转义**：所有用户输入内容自动转义
- **邮箱验证**：严格的邮箱格式检查
- **TLS加密**：支持SMTP连接加密
- **参数验证**：全面的配置参数验证

## 故障排除

### 常见问题

#### 1. 邮件发送失败
```
错误：邮件配置验证失败: SMTP服务器地址不能为空
解决：检查 --email-smtp-host 参数是否正确设置
```

#### 2. 身份验证失败
```
错误：535 Error: authentication failed
解决：
- QQ邮箱：使用授权码而不是登录密码
- Gmail：启用"应用专用密码"
- 检查用户名和密码是否正确
```

#### 3. 连接超时
```
错误：dial tcp: i/o timeout
解决：
- 检查网络连接
- 确认SMTP服务器地址和端口
- 尝试禁用TLS（设置 --email-enable-tls=false）
```

#### 4. 邮箱格式错误
```
错误：无效的邮箱地址: invalid-email
解决：确保邮箱地址格式正确，包含@和域名
```

### 调试建议
1. **逐步测试**：先用简单配置测试连通性
2. **查看日志**：使用 `--log-level=debug` 获取详细日志
3. **验证配置**：使用邮件客户端验证SMTP设置
4. **网络检查**：确认能够访问SMTP服务器

## 集成示例

### CI/CD集成
```yaml
# GitHub Actions 示例
- name: Run MySQL Tests with Email Notification
  run: |
    ./mysql-tester \
      --email-enable=true \
      --email-smtp-host="${{ secrets.SMTP_HOST }}" \
      --email-username="${{ secrets.EMAIL_USERNAME }}" \
      --email-password="${{ secrets.EMAIL_PASSWORD }}" \
      --email-to="${{ secrets.EMAIL_RECIPIENTS }}"
```

### 定时任务集成
```bash
# Crontab 示例：每天凌晨2点运行测试并发送报告
0 2 * * * /path/to/mysql-tester --email-enable=true --email-to="team@company.com"
```

## 扩展功能规划

- [ ] 支持邮件模板自定义
- [ ] 添加邮件发送重试机制
- [ ] 支持附件功能（测试报告文件）
- [ ] 集成更多通知渠道（钉钉、企业微信等）
- [ ] 支持邮件发送状态回调

---

## 技术实现细节

### 依赖库
- `gopkg.in/gomail.v2`：邮件发送核心库
- `html`：HTML内容转义
- `time`：时间格式化
- `strings`：字符串处理

### 关键算法
1. **邮箱验证算法**：多层验证确保邮箱格式正确
2. **内容截断算法**：智能控制邮件大小
3. **HTML安全算法**：防XSS攻击的内容转义

### 设计模式
- **策略模式**：HTML和纯文本两种邮件格式
- **建造者模式**：邮件内容分步构建
- **模板方法模式**：邮件发送流程标准化

这个邮件功能为MySQL Tester增加了强大的通知能力，让测试结果能够及时、美观地传达给相关人员，大大提升了测试流程的自动化程度和用户体验。 