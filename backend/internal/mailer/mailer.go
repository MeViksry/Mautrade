package mailer

import (
	"crypto/tls"
	"fmt"
	"net/mail"
	"net/smtp"
	"strings"

	"github.com/MeViksry/Mautrade/backend/internal/config"
)

// Mailer handles sending emails via SMTP
type Mailer struct {
	cfg *config.Config
}

type otpEmailContent struct {
	Subject    string
	Badge      string
	Title      string
	Lead       string
	Warning    string
	CodeLabel  string
	ExpiryNote string
}

func NewMailer(cfg *config.Config) *Mailer {
	return &Mailer{
		cfg: cfg,
	}
}

func (m *Mailer) sendHTML(from, to mail.Address, subject, htmlBody string) error {
	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"utf-8\"\r\n\r\n%s",
		from.String(), to.String(), subject, htmlBody)

	serverName := fmt.Sprintf("%s:%s", m.cfg.SMTPHost, m.cfg.SMTPPort)
	auth := smtp.PlainAuth("", m.cfg.SMTPUsername, m.cfg.SMTPPassword, m.cfg.SMTPHost)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         m.cfg.SMTPHost,
	}

	var client *smtp.Client
	var err error

	// If port is 465, use implicit TLS (dial TLS directly)
	if m.cfg.SMTPPort == "465" {
		conn, err := tls.Dial("tcp", serverName, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to dial SMTP server via TLS: %v", err)
		}
		client, err = smtp.NewClient(conn, m.cfg.SMTPHost)
		if err != nil {
			conn.Close()
			return fmt.Errorf("failed to create SMTP client: %v", err)
		}
	} else {
		// For other ports (like 587), start with a plain connection and then use STARTTLS
		client, err = smtp.Dial(serverName)
		if err != nil {
			return fmt.Errorf("failed to dial SMTP server: %v", err)
		}

		// Use STARTTLS if supported/required
		if err = client.StartTLS(tlsConfig); err != nil {
			client.Close()
			return fmt.Errorf("failed to upgrade to STARTTLS: %v", err)
		}
	}

	defer client.Quit()

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate SMTP: %v", err)
	}

	if err = client.Mail(from.Address); err != nil {
		return fmt.Errorf("failed to set sender: %v", err)
	}

	if err = client.Rcpt(to.Address); err != nil {
		return fmt.Errorf("failed to set recipient: %v", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to start data stream: %v", err)
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write email body: %v", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data stream: %v", err)
	}

	return nil
}

// SendOTP sends a 6-digit OTP to the user
func (m *Mailer) SendOTP(toEmail, firstName, otp, reason string) error {
	fromAddress := m.cfg.SMTPFrom
	if reason == "register" {
		fromAddress = "verify@mautrade.com"
	} else if reason == "login" {
		fromAddress = "otp@mautrade.com"
	}

	from := mail.Address{Name: "Mautrade", Address: fromAddress}
	to := mail.Address{Name: firstName, Address: toEmail}

	content := buildOTPEmailContent(reason)

	body := fmt.Sprintf(`<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta name="color-scheme" content="light dark">
  <meta name="supported-color-schemes" content="light dark">
  <style>
    @import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap');
  </style>
</head>
<body style="margin:0;padding:40px 16px;background-color:#0a0a0a;font-family:'Inter',sans-serif;color:#ebebeb;">
  <div style="max-width:560px;margin:0 auto;background-color:#121212;border:1px solid #2a2a2a;border-radius:20px;overflow:hidden;">
    <div style="height:4px;background:linear-gradient(90deg,#ff5a00 0%%,#cc4700 50%%,#ff5a00 100%%);"></div>
    <div style="padding:32px;">
      <h2 style="font-size:24px;font-weight:700;color:#ebebeb;margin:0 0 20px 0;">MAU<span style="color:#ff5a00;">TRADE</span></h2>
      <div style="display:inline-block;background-color:#2a1400;border:1px solid #ff5a00;border-radius:999px;padding:6px 14px;font-size:11px;font-weight:700;letter-spacing:0.9px;text-transform:uppercase;color:#ff5a00;margin-bottom:18px;">%s</div>
      <h3 style="font-size:22px;color:#ebebeb;margin:0 0 14px 0;">%s</h3>
      <p style="font-size:15px;margin:0 0 16px 0;">Hi <strong>%s</strong>,</p>
      <p style="font-size:14px;color:#888888;line-height:1.7;margin:0 0 20px 0;">%s</p>

      <div style="background-color:#1a1a1a;border:1px solid #2a2a2a;border-radius:16px;padding:18px 18px 8px 18px;margin-bottom:22px;">
        <div style="font-size:12px;font-weight:700;letter-spacing:0.08em;text-transform:uppercase;color:#c8c8c8;margin-bottom:14px;">%s</div>
        <div style="background-color:#121212;border:1px solid #2a2a2a;border-radius:14px;padding:22px 16px;text-align:center;margin-bottom:10px;">
          <span style="font-size:34px;font-weight:700;letter-spacing:8px;color:#ff5a00;">%s</span>
        </div>
        <div style="font-size:12px;color:#888888;line-height:1.6;">%s</div>
      </div>

      <div style="background-color:#1a1407;border:1px solid #6b4f1d;border-radius:14px;padding:14px 16px;margin-bottom:20px;">
        <p style="font-size:13px;line-height:1.6;color:#fcd34d;margin:0;">
          %s
        </p>
      </div>

      <p style="font-size:12px;color:#555555;margin:0;">This is an automated security message from Mautrade.</p>
    </div>
  </div>
</body>
</html>`,
		content.Badge,
		content.Title,
		firstName,
		content.Lead,
		content.CodeLabel,
		otp,
		content.ExpiryNote,
		content.Warning,
	)

	return m.sendHTML(from, to, content.Subject, body)
}

func buildOTPEmailContent(reason string) otpEmailContent {
	reason = strings.TrimSpace(reason)
	lowerReason := strings.ToLower(reason)

	content := otpEmailContent{
		Subject:    "Mautrade Security Action Required",
		Badge:      "Security Verification",
		Title:      "Your verification code is ready",
		Lead:       "Use the one-time code below to continue your security check in Mautrade.",
		Warning:    "If you did not request this code, please ignore this email and review your account security.",
		CodeLabel:  "One-time verification code",
		ExpiryNote: "This code expires in <strong style=\"color:#e2e8f0;\">10 minutes</strong>.",
	}

	switch {
	case strings.Contains(lowerReason, "login"):
		content.Subject = "Mautrade login verification code"
		content.Badge = "Login Verification"
		content.Title = "Your login verification code is ready"
		content.Lead = "Here is your login verification code. Use the one-time code below to continue your secure login into Mautrade."
	case strings.Contains(lowerReason, "register"):
		content.Subject = "Verify your new Mautrade account"
		content.Badge = "Account Verification"
		content.Title = "Verify your email address"
		content.Lead = "Welcome to Mautrade! Use the one-time code below to verify your email address and activate your account."
	}

	return content
}
