package email

// getDefaultTemplates 获取默认邮件模板
func (s *emailService) getDefaultTemplates() []*EmailTemplate {
	return []*EmailTemplate{
		// 验证码模板 - 中文
		{
			Name:        TemplateVerificationCode,
			Language:    "zh-CN",
			Subject:     "【{{.app_name}}】邮箱验证码",
			HTMLBody:    getVerificationCodeHTML_ZH(),
			TextBody:    getVerificationCodeText_ZH(),
			IsActive:    true,
			Description: "邮箱验证码模板",
		},
		// 密码重置模板 - 中文
		{
			Name:        TemplatePasswordReset,
			Language:    "zh-CN",
			Subject:     "【{{.app_name}}】密码重置",
			HTMLBody:    getPasswordResetHTML_ZH(),
			TextBody:    getPasswordResetText_ZH(),
			IsActive:    true,
			Description: "密码重置模板",
		},
		// 欢迎邮件模板 - 中文
		{
			Name:        TemplateWelcome,
			Language:    "zh-CN",
			Subject:     "欢迎使用{{.app_name}}！",
			HTMLBody:    getWelcomeHTML_ZH(),
			TextBody:    getWelcomeText_ZH(),
			IsActive:    true,
			Description: "欢迎邮件模板",
		},
		// 安全警告模板 - 中文
		{
			Name:        TemplateSecurityAlert,
			Language:    "zh-CN",
			Subject:     "【{{.app_name}}】安全警告",
			HTMLBody:    getSecurityAlertHTML_ZH(),
			TextBody:    getSecurityAlertText_ZH(),
			IsActive:    true,
			Description: "安全警告模板",
		},
	}
}

// 验证码HTML模板
func getVerificationCodeHTML_ZH() string {
	return `<!DOCTYPE html>
<html><head><meta charset="UTF-8"><title>邮箱验证码</title>
<style>
body{font-family:'Microsoft YaHei',Arial;margin:0;padding:20px;background:#f5f5f5}
.container{max-width:600px;margin:0 auto;background:#fff;border-radius:8px;box-shadow:0 2px 10px rgba(0,0,0,0.1)}
.header{background:linear-gradient(135deg,#667eea 0%,#764ba2 100%);color:white;padding:30px;text-align:center}
.content{padding:40px 30px}
.code-box{background:#f8f9fa;border:2px dashed #007bff;border-radius:8px;padding:20px;text-align:center;margin:20px 0}
.code{font-size:32px;font-weight:bold;color:#007bff;letter-spacing:8px;font-family:monospace}
.footer{background:#f8f9fa;padding:20px;text-align:center;color:#666;font-size:12px}
.warning{background:#fff3cd;border:1px solid #ffeaa7;border-radius:4px;padding:15px;margin:20px 0;color:#856404}
</style></head>
<body>
<div class="container">
<div class="header"><h1>{{.app_name}}</h1><p>邮箱验证码</p></div>
<div class="content">
<h2>验证您的邮箱地址</h2>
<p>您好！感谢您注册{{.app_name}}，请使用以下验证码完成邮箱验证：</p>
<div class="code-box"><div class="code">{{.code}}</div><p style="margin:10px 0 0 0;color:#666">验证码</p></div>
<div class="warning"><strong>注意事项：</strong>
<ul><li>验证码有效期为 {{.expires_in}} 分钟</li><li>请不要将验证码泄露给他人</li><li>如果您没有申请验证码，请忽略此邮件</li></ul>
</div>
</div>
<div class="footer"><p>此邮件由系统自动发送，请勿回复</p><p>&copy; {{.app_name}} 团队</p></div>
</div></body></html>`
}

// 验证码文本模板
func getVerificationCodeText_ZH() string {
	return `{{.app_name}} - 邮箱验证码

您好！感谢您注册{{.app_name}}，请使用以下验证码完成邮箱验证：

验证码：{{.code}}

注意事项：
- 验证码有效期为 {{.expires_in}} 分钟
- 请不要将验证码泄露给他人
- 如果您没有申请验证码，请忽略此邮件

此邮件由系统自动发送，请勿回复
© {{.app_name}} 团队`
}

// 密码重置HTML模板
func getPasswordResetHTML_ZH() string {
	return `<!DOCTYPE html>
<html><head><meta charset="UTF-8"><title>密码重置</title>
<style>
body{font-family:'Microsoft YaHei',Arial;margin:0;padding:20px;background:#f5f5f5}
.container{max-width:600px;margin:0 auto;background:#fff;border-radius:8px;box-shadow:0 2px 10px rgba(0,0,0,0.1)}
.header{background:linear-gradient(135deg,#ff6b6b 0%,#feca57 100%);color:white;padding:30px;text-align:center}
.content{padding:40px 30px}
.btn{display:inline-block;background:#007bff;color:white;padding:15px 30px;text-decoration:none;border-radius:5px;font-weight:bold;margin:20px 0}
.footer{background:#f8f9fa;padding:20px;text-align:center;color:#666;font-size:12px}
.warning{background:#f8d7da;border:1px solid #f5c6cb;border-radius:4px;padding:15px;margin:20px 0;color:#721c24}
</style></head>
<body>
<div class="container">
<div class="header"><h1>{{.app_name}}</h1><p>密码重置请求</p></div>
<div class="content">
<h2>重置您的密码</h2>
<p>我们收到了您的密码重置请求。点击下面的按钮来重置您的密码：</p>
<div style="text-align:center;margin:30px 0"><a href="{{.reset_url}}" class="btn">重置密码</a></div>
<div class="warning"><strong>安全提醒：</strong>
<ul><li>此链接有效期为 {{.expires_in}} 小时</li><li>如果您没有申请密码重置，请忽略此邮件</li><li>为了您的账户安全，请不要将此链接分享给他人</li></ul>
</div>
</div>
<div class="footer"><p>此邮件由系统自动发送，请勿回复</p><p>&copy; {{.app_name}} 团队</p></div>
</div></body></html>`
}

// 密码重置文本模板
func getPasswordResetText_ZH() string {
	return `{{.app_name}} - 密码重置请求

我们收到了您的密码重置请求。

请访问以下链接来重置您的密码：
{{.reset_url}}

安全提醒：
- 此链接有效期为 {{.expires_in}} 小时
- 如果您没有申请密码重置，请忽略此邮件
- 为了您的账户安全，请不要将此链接分享给他人

此邮件由系统自动发送，请勿回复
© {{.app_name}} 团队`
}

// 欢迎邮件HTML模板
func getWelcomeHTML_ZH() string {
	return `<!DOCTYPE html>
<html><head><meta charset="UTF-8"><title>欢迎使用</title>
<style>
body{font-family:'Microsoft YaHei',Arial;margin:0;padding:20px;background:#f5f5f5}
.container{max-width:600px;margin:0 auto;background:#fff;border-radius:8px;box-shadow:0 2px 10px rgba(0,0,0,0.1)}
.header{background:linear-gradient(135deg,#4facfe 0%,#00f2fe 100%);color:white;padding:30px;text-align:center}
.content{padding:40px 30px}
.feature{background:#f8f9fa;padding:20px;border-radius:8px;margin:15px 0}
.footer{background:#f8f9fa;padding:20px;text-align:center;color:#666;font-size:12px}
</style></head>
<body>
<div class="container">
<div class="header"><h1>欢迎使用 {{.app_name}}！</h1><p>您的云存储之旅从这里开始</p></div>
<div class="content">
<h2>{{.username}}，欢迎您！</h2>
<p>感谢您注册{{.app_name}}！我们很高兴您加入我们的社区。</p>
<h3>开始使用以下功能：</h3>
<div class="feature"><h4>📁 文件管理</h4><p>上传、下载、分享您的文件，支持多种格式</p></div>
<div class="feature"><h4>👥 团队协作</h4><p>与团队成员共享文件夹，实时协作</p></div>
<div class="feature"><h4>💬 即时通讯</h4><p>与团队成员实时交流，提高工作效率</p></div>
<div class="feature"><h4>🔒 安全存储</h4><p>企业级安全保护，您的数据安全无忧</p></div>
</div>
<div class="footer"><p>此邮件由系统自动发送，请勿回复</p><p>&copy; {{.app_name}} 团队</p></div>
</div></body></html>`
}

// 欢迎邮件文本模板
func getWelcomeText_ZH() string {
	return `欢迎使用 {{.app_name}}！

{{.username}}，欢迎您！

感谢您注册{{.app_name}}！我们很高兴您加入我们的社区。

开始使用以下功能：
📁 文件管理 - 上传、下载、分享您的文件
👥 团队协作 - 与团队成员共享文件夹，实时协作
💬 即时通讯 - 与团队成员实时交流，提高工作效率
🔒 安全存储 - 企业级安全保护，您的数据安全无忧

此邮件由系统自动发送，请勿回复
© {{.app_name}} 团队`
}

// 安全警告HTML模板
func getSecurityAlertHTML_ZH() string {
	return `<!DOCTYPE html>
<html><head><meta charset="UTF-8"><title>安全警告</title>
<style>
body{font-family:'Microsoft YaHei',Arial;margin:0;padding:20px;background:#f5f5f5}
.container{max-width:600px;margin:0 auto;background:#fff;border-radius:8px;box-shadow:0 2px 10px rgba(0,0,0,0.1)}
.header{background:linear-gradient(135deg,#ff4757 0%,#c44569 100%);color:white;padding:30px;text-align:center}
.content{padding:40px 30px}
.alert{background:#f8d7da;border:1px solid #f5c6cb;border-radius:4px;padding:15px;margin:20px 0;color:#721c24}
.footer{background:#f8f9fa;padding:20px;text-align:center;color:#666;font-size:12px}
</style></head>
<body>
<div class="container">
<div class="header"><h1>⚠️ 安全警告</h1><p>{{.app_name}} 安全中心</p></div>
<div class="content">
<h2>检测到{{.alert_type}}安全事件</h2>
<p>我们在您的账户中检测到了一个安全事件，详情如下：</p>
<div class="alert">
<h4>事件详情：</h4>
<p><strong>时间：</strong> {{.timestamp}}</p>
<p><strong>类型：</strong> {{.alert_type}}</p>
</div>
<h3>建议的安全措施：</h3>
<ul><li>立即更改您的密码</li><li>检查账户的最近活动</li><li>启用双因素认证</li><li>如果这不是您的操作，请立即联系我们</li></ul>
</div>
<div class="footer"><p>此邮件由系统自动发送，请勿回复</p><p>&copy; {{.app_name}} 安全中心</p></div>
</div></body></html>`
}

// 安全警告文本模板
func getSecurityAlertText_ZH() string {
	return `{{.app_name}} - 安全警告

检测到{{.alert_type}}安全事件

我们在您的账户中检测到了一个安全事件：
时间：{{.timestamp}}
类型：{{.alert_type}}

建议的安全措施：
- 立即更改您的密码
- 检查账户的最近活动
- 启用双因素认证
- 如果这不是您的操作，请立即联系我们

此邮件由系统自动发送，请勿回复
© {{.app_name}} 安全中心`
}
