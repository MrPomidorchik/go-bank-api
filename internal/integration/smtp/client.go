package smtp

import (
	"crypto/tls"
	"fmt"
	"log"

	gomail "github.com/go-mail/mail/v2"
)

type Client struct {
	host     string
	port     int
	username string
	password string
	from     string
}

func NewClient(host string, port int, username, password, from string) *Client {
	return &Client{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (c *Client) SendEmail(to string, subject string, body string) error {
	message := gomail.NewMessage()

	message.SetHeader("From", c.from)
	message.SetHeader("To", to)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", body)

	dialer := gomail.NewDialer(c.host, c.port, c.username, c.password)

	dialer.TLSConfig = &tls.Config{
		ServerName:         c.host,
		InsecureSkipVerify: false,
	}

	if err := dialer.DialAndSend(message); err != nil {
		log.Printf("SMTP error: %v", err)
		return fmt.Errorf("email sending failed: %w", err)
	}

	return nil
}

func (c *Client) SendPaymentEmail(to string, amount float64) error {
	body := fmt.Sprintf(`
		<h1>Платеж успешно проведен</h1>
		<p>Сумма операции: <strong>%.2f RUB</strong></p>
		<small>Это автоматическое уведомление банковского сервиса.</small>
	`, amount)

	return c.SendEmail(to, "Платеж успешно проведен", body)
}
