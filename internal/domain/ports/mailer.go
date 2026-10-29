package ports

import "context"

// Mailer отправляет письмо в виде простого текста.
type Mailer interface {
	Send(ctx context.Context, to, subject, body string) error
}
