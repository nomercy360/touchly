package services

import (
	"fmt"
	"github.com/resend/resend-go/v2"
)

type MailMessage struct {
	From          string `json:"From"`
	To            string `json:"To"`
	Subject       string `json:"Subject"`
	HtmlBody      string `json:"HtmlBody"`
	MessageStream string `json:"MessageStream"`
}

type EmailClient struct {
	Client *resend.Client
}

func NewEmailClient(apiKey string) *EmailClient {
	return &EmailClient{
		Client: resend.NewClient(apiKey),
	}
}

func (c *EmailClient) SendEmail(message *MailMessage) error {
	params := &resend.SendEmailRequest{
		To:      []string{message.To},
		From:    message.From,
		Subject: message.Subject,
		Html:    message.HtmlBody,
	}

	sent, err := c.Client.Emails.Send(params)
	if err != nil {
		return err
	}

	fmt.Printf("Email sent: %v\n", sent)

	return nil
}
