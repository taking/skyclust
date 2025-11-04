/**
 * Email Notification Service
 * 이메일 알림 전송 서비스
 */

package notification

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"skyclust/internal/domain"
	"strings"

	"skyclust/pkg/logger"
)

type EmailService struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromEmail    string
	fromName     string
}

// NewEmailService 이메일 서비스 생성
func NewEmailService(smtpHost, smtpPort, smtpUsername, smtpPassword, fromEmail, fromName string) *EmailService {
	return &EmailService{
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		smtpUsername: smtpUsername,
		smtpPassword: smtpPassword,
		fromEmail:    fromEmail,
		fromName:     fromName,
	}
}

// SendNotification 이메일 알림 전송
func (s *EmailService) SendNotification(userEmail string, notification *domain.Notification) error {
	// 이메일 템플릿 생성
	subject := fmt.Sprintf("[%s] %s", strings.ToUpper(notification.Type), notification.Title)
	body, err := s.generateEmailBody(notification)
	if err != nil {
		return fmt.Errorf("failed to generate email body: %w", err)
	}

	// 이메일 전송
	return s.sendEmail(userEmail, subject, body)
}

// sendEmail 이메일 전송
func (s *EmailService) sendEmail(to, subject, body string) error {
	// SMTP 설정
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)

	// 이메일 헤더 생성
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	// 이메일 메시지 생성
	var msg bytes.Buffer
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(body)

	// 이메일 전송
	err := smtp.SendMail(addr, auth, s.fromEmail, []string{to}, msg.Bytes())
	if err != nil {
		logger.Errorf("Failed to send email: %v", err)
		return err
	}

	logger.Infof("Email sent successfully to %s", to)
	return nil
}

// generateEmailBody 이메일 본문 생성
func (s *EmailService) generateEmailBody(notification *domain.Notification) (string, error) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background-color: white;
            border-radius: 8px;
            padding: 30px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .header {
            border-bottom: 2px solid {{.HeaderColor}};
            padding-bottom: 20px;
            margin-bottom: 20px;
        }
        .title {
            font-size: 24px;
            font-weight: bold;
            color: {{.HeaderColor}};
            margin: 0;
        }
        .type {
            display: inline-block;
            background-color: {{.HeaderColor}};
            color: white;
            padding: 4px 12px;
            border-radius: 20px;
            font-size: 12px;
            font-weight: bold;
            text-transform: uppercase;
            margin-bottom: 10px;
        }
        .message {
            font-size: 16px;
            line-height: 1.6;
            margin: 20px 0;
        }
        .footer {
            margin-top: 30px;
            padding-top: 20px;
            border-top: 1px solid #eee;
            font-size: 14px;
            color: #666;
        }
        .priority {
            display: inline-block;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 12px;
            font-weight: bold;
            margin-left: 10px;
        }
        .priority-high { background-color: #fee2e2; color: #dc2626; }
        .priority-medium { background-color: #fef3c7; color: #d97706; }
        .priority-low { background-color: #d1fae5; color: #059669; }
        .priority-urgent { background-color: #fecaca; color: #b91c1c; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="type">{{.Type}}</div>
            <h1 class="title">{{.Title}}</h1>
            {{if .Priority}}
            <span class="priority priority-{{.Priority}}">{{.Priority}} Priority</span>
            {{end}}
        </div>
        
        <div class="message">
            {{.Message}}
        </div>
        
        {{if .Category}}
        <div style="margin: 20px 0; padding: 10px; background-color: #f8f9fa; border-radius: 4px;">
            <strong>Category:</strong> {{.Category}}
        </div>
        {{end}}
        
        <div class="footer">
            <p>This notification was sent from SkyClust at {{.CreatedAt}}.</p>
            <p>You can manage your notification preferences in your account settings.</p>
        </div>
    </div>
</body>
</html>
`

	// 타입별 색상 설정
	headerColor := "#3b82f6" // 기본 파란색
	switch notification.Type {
	case "success":
		headerColor = "#10b981"
	case "warning":
		headerColor = "#f59e0b"
	case "error":
		headerColor = "#ef4444"
	case "info":
		headerColor = "#3b82f6"
	}

	// 템플릿 데이터
	data := struct {
		Title       string
		Message     string
		Type        string
		Category    string
		Priority    string
		HeaderColor string
		CreatedAt   string
	}{
		Title:       notification.Title,
		Message:     notification.Message,
		Type:        notification.Type,
		Category:    notification.Category,
		Priority:    notification.Priority,
		HeaderColor: headerColor,
		CreatedAt:   notification.CreatedAt.Format("2006-01-02 15:04:05 UTC"),
	}

	// 템플릿 실행
	t, err := template.New("email").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
