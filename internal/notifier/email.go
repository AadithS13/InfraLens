package notifier

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/infralens/infralens/internal/model"
)

// EmailAdapter delivers notifications via SMTP.
// Supports port 587 with STARTTLS (Gmail, most providers).
//
// Required env vars (set via cmd/api/main.go):
//   SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS, NOTIFY_EMAIL_TO
type EmailAdapter struct {
	host     string // e.g. smtp.gmail.com
	port     string // e.g. 587
	user     string // sender + auth username
	pass     string // app password
	from     string // from address (defaults to user)
	to       []string
}

type EmailConfig struct {
	Host string
	Port string
	User string
	Pass string
	From string   // optional — defaults to User
	To   []string // at least one recipient required
}

func NewEmailAdapter(cfg EmailConfig) *EmailAdapter {
	from := cfg.From
	if from == "" {
		from = cfg.User
	}
	return &EmailAdapter{
		host: cfg.Host,
		port: cfg.Port,
		user: cfg.User,
		pass: cfg.Pass,
		from: from,
		to:   cfg.To,
	}
}

func (a *EmailAdapter) Name() string { return "email" }

func (a *EmailAdapter) Send(_ context.Context, c model.NotifiableChange) error {
	label := fieldLabels[c.FieldName]
	if label == "" {
		label = c.FieldName
	}

	subject := fmt.Sprintf("InfraLens Alert: %s Changed — %s", label, c.ProjectName)
	body := fmt.Sprintf(
		"Project  : %s\nField    : %s\nChanged  : %s → %s\nDetected : %s\n\n---\nSent by InfraLens · MahaRERA Intelligence Platform",
		c.ProjectName,
		label,
		c.OldValue,
		c.NewValue,
		c.DetectedAt.UTC().Format(time.RFC3339),
	)

	msg := fmt.Sprintf(
		"From: InfraLens <%s>\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		a.from,
		strings.Join(a.to, ", "),
		subject,
		body,
	)

	auth := smtp.PlainAuth("", a.user, a.pass, a.host)
	addr := a.host + ":" + a.port
	if err := smtp.SendMail(addr, auth, a.from, a.to, []byte(msg)); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	return nil
}
