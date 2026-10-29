package mailer

import (
	"bufio"
	"context"
	"encoding/base64"
	"io"
	"mime"
	"net"
	"net/mail"
	"strconv"
	"strings"
	"testing"
)

// fakeSMTP — минимальный SMTP-сервер: принимает одно письмо и запоминает AUTH и DATA.
type fakeSMTP struct {
	ln   net.Listener
	auth chan string
	data chan string
}

func startFakeSMTP(t *testing.T) *fakeSMTP {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	s := &fakeSMTP{ln: ln, auth: make(chan string, 1), data: make(chan string, 1)}
	t.Cleanup(func() { _ = ln.Close() })

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		r := bufio.NewReader(conn)
		write := func(s string) { _, _ = io.WriteString(conn, s+"\r\n") }
		write("220 fake ready")
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				return
			}
			cmd := strings.TrimSpace(line)
			switch {
			case strings.HasPrefix(cmd, "EHLO"):
				write("250-fake")
				write("250 AUTH PLAIN")
			case strings.HasPrefix(cmd, "AUTH PLAIN"):
				s.auth <- strings.TrimPrefix(cmd, "AUTH PLAIN ")
				write("235 ok")
			case strings.HasPrefix(cmd, "MAIL"), strings.HasPrefix(cmd, "RCPT"):
				write("250 ok")
			case cmd == "DATA":
				write("354 go")
				var msg strings.Builder
				for {
					l, err := r.ReadString('\n')
					if err != nil {
						return
					}
					if l == ".\r\n" {
						break
					}
					msg.WriteString(l)
				}
				s.data <- msg.String()
				write("250 queued")
			case cmd == "QUIT":
				write("221 bye")
				return
			default:
				write("250 ok")
			}
		}
	}()
	return s
}

func TestSMTPMailerSendsRussianPlainText(t *testing.T) {
	srv := startFakeSMTP(t)
	_, port, _ := net.SplitHostPort(srv.ln.Addr().String())
	p, _ := strconv.Atoi(port)

	m, err := NewSMTPMailer(Config{
		Host: "127.0.0.1", Port: p, Username: "admin@example.com", Password: "secret",
		FromName: "Калькулятор отпусков", TLSMode: TLSModeNone,
	})
	if err != nil {
		t.Fatal(err)
	}

	subject := "Напоминание о начале отпуска с 26.11.2026"
	body := "Здравствуйте, Пётр Александрович!\n\nНапоминаем об отпуске."
	if err := m.Send(context.Background(), "petr@example.com", subject, body); err != nil {
		t.Fatalf("Send: %v", err)
	}

	creds, _ := base64.StdEncoding.DecodeString(<-srv.auth)
	if string(creds) != "\x00admin@example.com\x00secret" {
		t.Errorf("auth = %q, want login and password of the admin mailbox", creds)
	}

	msg, err := mail.ReadMessage(strings.NewReader(<-srv.data))
	if err != nil {
		t.Fatalf("parse message: %v", err)
	}
	gotSubject, _ := new(mime.WordDecoder).DecodeHeader(msg.Header.Get("Subject"))
	if gotSubject != subject {
		t.Errorf("subject = %q, want %q", gotSubject, subject)
	}
	from, _ := mail.ParseAddress(msg.Header.Get("From"))
	if from == nil || from.Address != "admin@example.com" || from.Name != "Калькулятор отпусков" {
		t.Errorf("from = %v", from)
	}
	if msg.Header.Get("To") != "petr@example.com" {
		t.Errorf("to = %q", msg.Header.Get("To"))
	}
	raw, _ := io.ReadAll(msg.Body)
	decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(string(raw), "\r\n", ""))
	if err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if !strings.Contains(string(decoded), "Здравствуйте, Пётр Александрович!") {
		t.Errorf("body = %q", decoded)
	}
}

func TestNewSMTPMailerDefaults(t *testing.T) {
	m, err := NewSMTPMailer(Config{Host: "smtp.example.com", Username: "admin@example.com", Password: "x"})
	if err != nil {
		t.Fatal(err)
	}
	if m.cfg.Port != 465 || m.cfg.TLSMode != TLSModeSSL || m.cfg.From != "admin@example.com" {
		t.Errorf("defaults = %+v", m.cfg)
	}
	m, _ = NewSMTPMailer(Config{Host: "smtp.example.com", Port: 587, Username: "a@b.ru", Password: "x"})
	if m.cfg.TLSMode != TLSModeSTARTTLS {
		t.Errorf("port 587 mode = %q, want starttls", m.cfg.TLSMode)
	}
	if _, err := NewSMTPMailer(Config{Host: "smtp.example.com"}); err == nil {
		t.Error("missing credentials must be an error")
	}
}
