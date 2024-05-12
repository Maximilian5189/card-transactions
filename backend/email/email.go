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

func getValue(body string, key string) string {
	lines := strings.Split(body, "\n")

	var merchantLine string
	for _, line := range lines {
		if strings.HasPrefix(line, fmt.Sprintf("%s:", key)) {
			merchantLine = line
			break
		}
	}

	regex := fmt.Sprintf("%s:", key) + `\s*(.*)`
	re := regexp.MustCompile(regex)
	match := re.FindStringSubmatch(merchantLine)
	if len(match) > 1 {
		sanitized := strings.ReplaceAll(match[1], "\r", "")
		return strings.ReplaceAll(sanitized, "$", "")
	}
	return ""
}

func (e *EmailService) GetEmails() {
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

	d, err := db.NewTransactionsDB(e.logger)
	if err != nil {
		e.logger.Error(fmt.Sprintf("Unable to open DB: %v", err))
	}

	for _, m := range r.Messages {
		msg, err := srv.Users.Messages.Get(user, m.Id).Do()
		if err != nil {
			e.logger.Error(fmt.Sprintf("Unable to retrieve content: %v", err))
		}

		transaction := db.Transaction{}
		for _, h := range msg.Payload.Headers {
			if h.Name == "Date" {
				transaction.Date = h.Value
			} else if h.Name == "From" {
				if !strings.Contains(h.Value, "Discover") {
					continue
				}
			}

			if h.Name == "Message-ID" {
				transaction.MessageID = h.Value
			}
		}

		if msg.Payload.Body.Data != "" {
			bodyBytes, err := base64.StdEncoding.DecodeString(msg.Payload.Body.Data)
			if err != nil {
				e.logger.Error(fmt.Sprintf("Unable to retrieve content: %v", err))
			}

			body := string(bodyBytes)
			name := getValue(body, "Merchant")
			if name != "" {
				transaction.Name = name
			}

			amount := getValue(body, "Amount")
			if amount != "" {
				transaction.Amount = amount
			}

		} else if len(msg.Payload.Parts) > 0 {
			for _, part := range msg.Payload.Parts {
				if part.MimeType == "text/plain" || part.MimeType == "text/html" {
					if part.Body.Data != "" {
						bodyBytes, err := base64.StdEncoding.DecodeString(part.Body.Data)
						if err != nil {
							e.logger.Error(fmt.Sprintf("error getting content: %v", err))
						}
						body := string(bodyBytes)
						name := getValue(body, "Merchant")
						if name != "" {
							transaction.Name = name
						}

						amount := getValue(body, "Amount")
						if amount != "" {
							transaction.Amount = amount
						}
					}
				}
			}
		}
		err = d.Insert(transaction)
		e.logger.Error(fmt.Sprintf("DB insert err: %v", err))
	}
}
