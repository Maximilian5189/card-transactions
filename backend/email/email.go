package email

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"strings"

	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	auth "backend/localauth"
)

func GetEmails() {
	ctx := context.Background()
	client := auth.GetClient()

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Printf("Unable to retrieve Gmail client: %v", err)
	}

	user := "me"
	r, err := srv.Users.Messages.List(user).Do()
	if err != nil {
		log.Printf("Unable to retrieve emails: %v", err)
	}

	for _, m := range r.Messages {
		msg, err := srv.Users.Messages.Get(user, m.Id).Do()
		if err != nil {
			log.Printf("Unable to retrieve content: %v", err)
		}

		isDiscover := false
		var date string
		for _, h := range msg.Payload.Headers {
			if h.Name == "Date" {
				date = h.Value
			} else if h.Name == "From" {
				// maybe check Message-ID for dedupe
				fmt.Printf("headers: %s - %s\n", h.Name, h.Value)

				if strings.Contains(h.Value, "Discover") {
					isDiscover = true
				}
			}
		}

		if !isDiscover {
			continue
		}

		fmt.Printf("date: %s\n", date)

		if msg.Payload.Body.Data != "" {
			bodyBytes, err := base64.StdEncoding.DecodeString(msg.Payload.Body.Data)
			if err != nil {
				log.Printf("Unable to retrieve content: %v", err)
			}

			body := string(bodyBytes)
			fmt.Printf("body: %s\n", body)
		} else if len(msg.Payload.Parts) > 0 {
			for _, part := range msg.Payload.Parts {
				if part.MimeType == "text/plain" || part.MimeType == "text/html" {
					if part.Body.Data != "" {
						bodyBytes, err := base64.StdEncoding.DecodeString(part.Body.Data)
						if err != nil {
							log.Printf("Unable to retrieve content: %v", err)
						}
						body := string(bodyBytes)
						fmt.Printf("body: %s\n", body)
					}
				}
			}
		}
	}
}

func GetLabels() {
	ctx := context.Background()
	client := auth.GetClient()

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	user := "me"
	r, err := srv.Users.Labels.List(user).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve labels: %v", err)
	}
	if len(r.Labels) == 0 {
		fmt.Println("No labels found.")
		return
	}
	fmt.Println("Labels:")
	for _, l := range r.Labels {
		fmt.Printf("- %s\n", l.Name)
	}
}
