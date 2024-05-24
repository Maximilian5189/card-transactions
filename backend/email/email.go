package email

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	"backend/db"
	auth "backend/localauth"
	"backend/logger"
)

type EmailService struct {
	logger logger.Logger
	srv    *gmail.Service
}

func NewEmailService(logger logger.Logger) (EmailService, error) {
	ctx := context.Background()
	client := auth.GetClient()

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		logger.Error(fmt.Sprintf("Unable to retrieve Gmail client: %v", err))
		return EmailService{}, err
	}

	return EmailService{logger, srv}, nil
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
	user := "me"
	r, err := e.srv.Users.Messages.List(user).Do()
	if err != nil {
		e.logger.Error(fmt.Sprintf("Unable to retrieve emails: %v", err))
	}

	d, err := db.NewTransactionsDB(e.logger)
	if err != nil {
		e.logger.Error(fmt.Sprintf("Unable to open DB: %v", err))
	}

	for _, m := range r.Messages {
		msg, err := e.srv.Users.Messages.Get(user, m.Id).Do()
		if err != nil {
			e.logger.Error(fmt.Sprintf("Unable to retrieve content: %v", err))
			continue
		}

		isDiscover := false
		dateError := false
		from := ""
		transaction := db.Transaction{}
		for _, h := range msg.Payload.Headers {
			if h.Name == "Date" {
				sanitized := strings.ReplaceAll(h.Value, "(UTC)", "")
				sanitized = strings.ReplaceAll(sanitized, "(CET)", "")
				sanitized = strings.ReplaceAll(sanitized, "(GMT)", "")
				sanitized = strings.ReplaceAll(sanitized, "(EST)", "")
				sanitized = strings.TrimSpace(sanitized)
				layout := "Mon, 2 Jan 2006 15:04:05 -0700"
				t, err := time.Parse(layout, sanitized)
				if err != nil {
					dateError = true
					e.logger.Error(fmt.Sprintf("Unable to convert date: %v", err))
					transaction.Date = time.Now().Unix()
				} else {
					transaction.Date = t.Unix()
				}
			} else if h.Name == "From" {
				from = h.Value
				if strings.Contains(h.Value, "Discover") {
					isDiscover = true
				}
			}

			if h.Name == "Message-ID" {
				transaction.MessageID = h.Value
			}
		}
		if dateError {
			e.logger.Info("attention! Date error for email from " + from)
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
			name := getValue(body, "Merchant")
			if name != "" {
				transaction.Name = name
			}

			amount := getValue(body, "Amount")
			if amount != "" {
				a, err := strconv.ParseFloat(amount, 64)
				if err != nil {
					e.logger.Error(fmt.Sprintf("unable to get amount: %v", err))
				} else {
					transaction.Amount = a
				}
			}

		} else if len(msg.Payload.Parts) > 0 {
			for _, part := range msg.Payload.Parts {
				if part.MimeType == "text/plain" || part.MimeType == "text/html" {
					if part.Body.Data != "" {
						bodyBytes, err := base64.StdEncoding.DecodeString(part.Body.Data)
						if err != nil && !strings.Contains(err.Error(), "illegal base64 data") {
							e.logger.Error(fmt.Sprintf("error getting content: %v", err))
						}
						body := string(bodyBytes)
						name := getValue(body, "Merchant")
						if name != "" {
							transaction.Name = name
						}

						amount := getValue(body, "Amount")
						if amount != "" {
							a, err := strconv.ParseFloat(amount, 64)
							if err != nil {
								e.logger.Error(fmt.Sprintf("unable to get amount: %v", err))
							} else {
								transaction.Amount = a
							}
						}
					}
				}
			}
		}
		// todo not insert if amount = 0? Probably not a transaction alert
		// or identify by subject?
		err = d.Insert(transaction)

		// we always process all emails and messageid UNIQUE constraint on DB level avoids duplicates
		// so we don't log these errors to not clutter the logs, as these errors are expected
		if err != nil && !strings.Contains(err.Error(), "UNIQUE constraint failed") {
			e.logger.Error(fmt.Sprintf("DB insert err: %v", err))
		}
	}
}
