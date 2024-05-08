package email

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	"backend/db"
	auth "backend/localauth"
	"backend/logger"
)

type EmailService struct {
	logger logger.Logger
}

func NewEmailService(logger logger.Logger) EmailService {
	return EmailService{logger}
}

func getMerchant(body string) string {
	lines := strings.Split(body, "\n")

	var merchantLine string
	for _, line := range lines {
		if strings.HasPrefix(line, "Merchant:") {
			merchantLine = line
			break
		}
	}

	re := regexp.MustCompile(`Merchant:\s*(.*)`)
	match := re.FindStringSubmatch(merchantLine)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

func (e *EmailService) GetEmails() error {
	ctx := context.Background()
	client := auth.GetClient()

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		e.logger.Error(fmt.Sprintf("Unable to retrieve Gmail client: %v", err))
	}

	user := "me"
	r, err := srv.Users.Messages.List(user).Do()
	if err != nil {
		e.logger.Error(fmt.Sprintf("Unable to retrieve emails: %v", err))
	}

	d, err := db.NewTransactionsDB()
	if err != nil {
		return err
	}

	for _, m := range r.Messages {
		msg, err := srv.Users.Messages.Get(user, m.Id).Do()
		if err != nil {
			e.logger.Error(fmt.Sprintf("Unable to retrieve content: %v", err))
		}

		isDiscover := false
		transaction := db.Transaction{}
		for _, h := range msg.Payload.Headers {
			if h.Name == "Date" {
				transaction.Date = h.Value
			} else if h.Name == "From" {
				if strings.Contains(h.Value, "Discover") {
					isDiscover = true
				}
			}

			if h.Name == "Message-ID" {
				transaction.MessageID = h.Value
			}
		}

		if !isDiscover {
			continue
		}

		if msg.Payload.Body.Data != "" {
			bodyBytes, err := base64.StdEncoding.DecodeString(msg.Payload.Body.Data)
			if err != nil {
				e.logger.Error(fmt.Sprintf("Unable to retrieve content: %v", err))
			}

			body := string(bodyBytes)
			transaction.Name = getMerchant(body)

		} else if len(msg.Payload.Parts) > 0 {
			for _, part := range msg.Payload.Parts {
				if part.MimeType == "text/plain" || part.MimeType == "text/html" {
					if part.Body.Data != "" {
						bodyBytes, err := base64.StdEncoding.DecodeString(part.Body.Data)
						if err != nil {
							e.logger.Error(fmt.Sprintf("Unable to retrieve content: %v", err))
						}
						body := string(bodyBytes)
						name := getMerchant(body)
						if name != "" {
							transaction.Name = getMerchant(body)
						}
					}
				}
			}
		}
		err = d.Insert(transaction)
		e.logger.Error(fmt.Sprintf("DB insert err: %v", err))
	}

	return err
}
