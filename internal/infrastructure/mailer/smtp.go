package mailer

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
)

// Режимы шифрования соединения с SMTP-сервером.
const (
	TLSModeSSL      = "ssl"      // TLS сразу при подключении (обычно порт 465)
	TLSModeSTARTTLS = "starttls" // переход на TLS командой STARTTLS (обычно порт 587)
	TLSModeNone     = "none"     // без шифрования — только для локальной отладки
)

const sendTimeout = 30 * time.Second

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string // адрес отправителя; по умолчанию Username
	FromName string // отображаемое имя отправителя
	TLSMode  string // ssl | starttls | none; по умолчанию ssl для 465, иначе starttls
}

type SMTPMailer struct {
	cfg Config
}

var _ ports.Mailer = (*SMTPMailer)(nil)

func NewSMTPMailer(cfg Config) (*SMTPMailer, error) {
	if cfg.Host == "" || cfg.Username == "" || cfg.Password == "" {
		return nil, errors.New("smtp: host, username and password are required")
	}
	if cfg.Port == 0 {
		cfg.Port = 465
	}
	if cfg.From == "" {
		cfg.From = cfg.Username
	}
	if cfg.TLSMode == "" {
		cfg.TLSMode = TLSModeSTARTTLS
		if cfg.Port == 465 {
			cfg.TLSMode = TLSModeSSL
		}
	}
	switch cfg.TLSMode {
	case TLSModeSSL, TLSModeSTARTTLS, TLSModeNone:
	default:
		return nil, fmt.Errorf("smtp: unknown TLS mode %q (ssl, starttls, none)", cfg.TLSMode)
	}
	if _, err := mail.ParseAddress(cfg.From); err != nil {
		return nil, fmt.Errorf("smtp: invalid from address %q: %w", cfg.From, err)
	}
	return &SMTPMailer{cfg: cfg}, nil
}

// From — адрес отправителя (ящик администратора).
func (m *SMTPMailer) From() string { return m.cfg.From }

func (m *SMTPMailer) Send(ctx context.Context, to, subject, body string) error {
	if _, err := mail.ParseAddress(to); err != nil {
		return fmt.Errorf("smtp: invalid recipient %q: %w", to, err)
	}

	ctx, cancel := context.WithTimeout(ctx, sendTimeout)
	defer cancel()

	addr := net.JoinHostPort(m.cfg.Host, strconv.Itoa(m.cfg.Port))
	tlsConfig := &tls.Config{ServerName: m.cfg.Host, MinVersion: tls.VersionTLS12}

	var conn net.Conn
	var err error
	if m.cfg.TLSMode == TLSModeSSL {
		conn, err = (&tls.Dialer{Config: tlsConfig}).DialContext(ctx, "tcp", addr)
	} else {
		conn, err = (&net.Dialer{}).DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("smtp: connect %s: %w", addr, err)
	}
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}

	c, err := smtp.NewClient(conn, m.cfg.Host)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("smtp: handshake: %w", err)
	}
	defer c.Close()

	if m.cfg.TLSMode == TLSModeSTARTTLS {
		if ok, _ := c.Extension("STARTTLS"); !ok {
			return errors.New("smtp: server does not support STARTTLS")
		}
		if err := c.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("smtp: starttls: %w", err)
		}
	}

	if ok, _ := c.Extension("AUTH"); ok {
		// PlainAuth отправляет пароль только по TLS или на localhost.
		if err := c.Auth(smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)); err != nil {
			return fmt.Errorf("smtp: auth: %w", err)
		}
	}

	if err := c.Mail(m.cfg.From); err != nil {
		return fmt.Errorf("smtp: MAIL FROM: %w", err)
	}
	if err := c.Rcpt(to); err != nil {
		return fmt.Errorf("smtp: RCPT TO %s: %w", to, err)
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("smtp: DATA: %w", err)
	}
	if _, err := w.Write(m.buildMessage(to, subject, body)); err != nil {
		return fmt.Errorf("smtp: write message: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp: send message: %w", err)
	}
	return c.Quit()
}

// buildMessage собирает письмо: заголовки в UTF-8 (RFC 2047), тело text/plain в base64.
func (m *SMTPMailer) buildMessage(to, subject, body string) []byte {
	from := (&mail.Address{Name: m.cfg.FromName, Address: m.cfg.From}).String()

	var b bytes.Buffer
	header := func(k, v string) { b.WriteString(k + ": " + v + "\r\n") }
	header("From", from)
	header("To", to)
	header("Subject", mime.BEncoding.Encode("UTF-8", subject))
	header("Date", time.Now().Format(time.RFC1123Z))
	header("Message-ID", "<"+randomID()+"@"+domainOf(m.cfg.From)+">")
	header("MIME-Version", "1.0")
	header("Content-Type", "text/plain; charset=UTF-8")
	header("Content-Transfer-Encoding", "base64")
	b.WriteString("\r\n")

	text := strings.ReplaceAll(body, "\n", "\r\n")
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	for len(encoded) > 76 {
		b.WriteString(encoded[:76] + "\r\n")
		encoded = encoded[76:]
	}
	b.WriteString(encoded + "\r\n")
	return b.Bytes()
}

func randomID() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

func domainOf(addr string) string {
	if i := strings.LastIndex(addr, "@"); i >= 0 {
		return addr[i+1:]
	}
	return "localhost"
}
