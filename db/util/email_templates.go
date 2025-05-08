package util

import (
	"bytes"
	"errors"
	t "html/template"

	"github.com/pocketbase/pocketbase/tools/template"
)

type EmailData struct {
	AppUrl        string
	RecipientName string
	Content       t.HTML
	Link          string
	FollowerName  string
}

var notificationTemplates = map[NotificationType]string{
	TrailCreate:  "{{.Author}} has created a new trail: {{.trail}}.",
	TrailShare:   "{{.Author}} has shared a trail with you: {{.trail}}.",
	ListCreate:   "{{.Author}} has created a new list: {{.list}}.",
	ListShare:    "{{.Author}} has shared a list with you: {{.list}}.",
	NewFollower:  "Good news! You have a new follower: {{.Author}}.",
	TrailComment: "{{.Author}} commented on your trail '{{.trail}}': '{{.comment}}'.",
}

func GenerateHTML(appUrl string, recipientName string, authorName string, notificationType NotificationType, metadata map[string]string) (string, error) {
	registry := template.NewRegistry()

	// Get custom content template for the notification type
	customTemplate, exists := notificationTemplates[notificationType]
	if !exists {
		return "", errors.New("unknown notification type")
	}

	// Create content template with dynamic data
	contentTmpl, err := t.New("content").Parse(customTemplate)
	if err != nil {
		return "", err
	}

	// Render the custom content with data
	metadata["Author"] = authorName
	var contentBuffer bytes.Buffer
	err = contentTmpl.Execute(&contentBuffer, metadata)
	if err != nil {
		return "", err
	}

	// Assemble the final email with boilerplate and custom content
	content := EmailData{
		AppUrl:        appUrl,
		RecipientName: recipientName,
		Content:       t.HTML(contentBuffer.String()),
	}

	html, err := registry.LoadFiles(
		"db/templates/mail/notification.html",
	).Render(content)

	if err != nil {
		return "", err
	}

	return html, nil
}
