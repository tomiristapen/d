package mailer

import (
	"log"
)

type Sender interface {
	SendVerificationCode(email, code string) error
	SendPasswordResetCode(email, code string) error
}

type LogSender struct {
	from    string
	baseURL string
}

func NewLogSender(from, baseURL string) *LogSender {
	return &LogSender{from: from, baseURL: baseURL}
}

func (s *LogSender) SendVerificationCode(email, code string) error {
	_ = s.baseURL
	log.Printf("from=%s to=%s verify_code=%s", s.from, email, code)
	return nil
}

func (s *LogSender) SendPasswordResetCode(email, code string) error {
	_ = s.baseURL
	log.Printf("from=%s to=%s password_reset_code=%s", s.from, email, code)
	return nil
}
