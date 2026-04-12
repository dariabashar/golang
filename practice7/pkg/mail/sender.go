package mail

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strconv"
)

type Sender struct {
	host     string
	port     int
	user     string
	password string
	from     string
}

func NewFromEnv() *Sender {
	port := 587
	if p := os.Getenv("SMTP_PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			port = n
		}
	}
	return &Sender{
		host:     os.Getenv("SMTP_HOST"),
		port:     port,
		user:     os.Getenv("SMTP_USER"),
		password: os.Getenv("SMTP_PASSWORD"),
		from:     os.Getenv("SMTP_FROM"),
	}
}

func (s *Sender) SendVerificationCode(to, code string) error {
	body := fmt.Sprintf("Subject: Verify your account\r\nTo: %s\r\n\r\nYour verification code is: %s\r\n", to, code)
	if s.host == "" || s.from == "" {
		log.Printf("[mail disabled] would send to %s code=%s", to, code)
		return nil
	}
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	auth := smtp.PlainAuth("", s.user, s.password, s.host)
	if err := smtp.SendMail(addr, auth, s.from, []string{to}, []byte(body)); err != nil {
		return fmt.Errorf("smtp: %w", err)
	}
	return nil
}
