package mailer

import (
	"fmt"
	"html"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
)

type SMTPSender struct {
	fromHeader   string
	fromEnvelope string
	baseURL      string

	host string
	port int
	user string
	pass string
}

func NewSMTPSender(from, baseURL, host, port, user, pass string) (*SMTPSender, error) {
	if from == "" {
		return nil, fmt.Errorf("MAILER_FROM is required")
	}
	if host == "" || port == "" {
		return nil, fmt.Errorf("SMTP_HOST and SMTP_PORT are required")
	}
	if (user == "") != (pass == "") {
		return nil, fmt.Errorf("SMTP_USER and SMTP_PASS must be set together")
	}

	portNum, err := strconv.Atoi(port)
	if err != nil || portNum <= 0 || portNum > 65535 {
		return nil, fmt.Errorf("invalid SMTP_PORT: %q", port)
	}

	fromAddr, err := mail.ParseAddress(from)
	if err != nil {
		return nil, fmt.Errorf("invalid MAILER_FROM: %w", err)
	}

	return &SMTPSender{
		fromHeader:   fromAddr.String(),
		fromEnvelope: fromAddr.Address,
		baseURL:      baseURL,
		host:         host,
		port:         portNum,
		user:         user,
		pass:         pass,
	}, nil
}

func (s *SMTPSender) SendVerificationCode(email, code string) error {
	return s.sendCodeEmail(email, "Verify your email", "Your verification code is:", code)
}

func (s *SMTPSender) SendPasswordResetCode(email, code string) error {
	return s.sendCodeEmail(email, "Reset your password", "Your password reset code is:", code)
}

func (s *SMTPSender) sendCodeEmail(email, subject, intro, code string) error {
	toAddr, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid recipient email: %w", err)
	}

	code = strings.TrimSpace(code)
	body := fmt.Sprintf(
		`<p>%s</p><p style="font-size:24px; font-weight:bold; letter-spacing:2px;">%s</p><p>This code expires soon. If you didn't request it, you can ignore this email.</p>`,
		html.EscapeString(intro),
		html.EscapeString(code),
	)

	msg := "" +
		fmt.Sprintf("From: %s\r\n", s.fromHeader) +
		fmt.Sprintf("To: %s\r\n", toAddr.String()) +
		fmt.Sprintf("Subject: %s\r\n", subject) +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" +
		body

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	var auth smtp.Auth
	if s.user != "" {
		auth = smtp.PlainAuth("", s.user, s.pass, s.host)
	}

	err = smtp.SendMail(addr, auth, s.fromEnvelope, []string{toAddr.Address}, []byte(msg))
	if err != nil && auth != nil && strings.Contains(err.Error(), "unencrypted connection") && isLocalSMTPHost(s.host) {
		return smtp.SendMail(addr, nil, s.fromEnvelope, []string{toAddr.Address}, []byte(msg))
	}
	return err
}

func isLocalSMTPHost(host string) bool {
	switch strings.ToLower(strings.TrimSpace(host)) {
	case "localhost", "127.0.0.1", "::1", "mailhog":
		return true
	default:
		return false
	}
}
