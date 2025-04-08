package util

import (
	"net/mail"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/mailer"
)

type NotificationType string

const (
	TrailCreate  NotificationType = "trail_create"
	TrailShare   NotificationType = "trail_share"
	ListCreate   NotificationType = "list_create"
	ListShare    NotificationType = "list_share"
	NewFollower  NotificationType = "new_follower"
	TrailComment NotificationType = "trail_comment"
)

type Notification struct {
	Type     NotificationType  `json:"type"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Seen     bool              `json:"seen"`
	Author   string            `json:"author"`
}

type NotificationSettings struct {
	Web   bool `json:"web"`
	Email bool `json:"email"`
}

func getNotificationPermissions(app core.App, user string, notificationType NotificationType) (*NotificationSettings, error) {
	settings, err := app.FindFirstRecordByFilter("settings", "user={:user}", dbx.Params{"user": user})
	if err != nil {
		return nil, err
	}
	var notificationPreferences map[NotificationType]NotificationSettings

	err = settings.UnmarshalJSONField("notifications", &notificationPreferences)
	if err != nil {
		return nil, err
	}

	settingsForType, exists := notificationPreferences[notificationType]
	if !exists {
		return &NotificationSettings{
			Web:   true,
			Email: true,
		}, nil
	}

	return &settingsForType, nil
}

func SendNotification(app core.App, notification Notification, recipient string) error {
	if notification.Author == recipient {
		return nil
	}
	permissions, err := getNotificationPermissions(app, recipient, notification.Type)
	if err != nil {
		return err
	}

	notifications, err := app.FindCollectionByNameOrId("notifications")
	if err != nil {
		return err
	}

	if permissions.Web {
		n := core.NewRecord(notifications)
		n.Set("type", string(notification.Type))
		n.Set("metadata", notification.Metadata)
		n.Set("seen", notification.Seen)
		n.Set("recipient", recipient)
		n.Set("author", notification.Author)

		if err := app.Save(n); err != nil {
			return err
		}
	}

	if permissions.Email {
		recipientUser, err := app.FindRecordById("users", recipient)
		if err != nil {
			return err
		}
		authorUser, err := app.FindRecordById("users", notification.Author)
		if err != nil {
			return err
		}
		html, err := GenerateHTML(app.Settings().Meta.AppURL, recipientUser.GetString("username"), authorUser.GetString("username"), notification.Type, notification.Metadata)
		if err != nil {
			return err
		}

		message := &mailer.Message{
			From: mail.Address{
				Address: app.Settings().Meta.SenderAddress,
				Name:    app.Settings().Meta.SenderName,
			},
			To:      []mail.Address{{Address: recipientUser.Email()}},
			Subject: "wanderer - New Notification",
			HTML:    html,
		}

		app.NewMailClient().Send(message)
	}
	return nil
}

func SendNotificationToFollowers(app core.App, notification Notification) error {
	followers, err := app.FindRecordsByFilter("follows", "followee={:user}", "", -1, 0, dbx.Params{"user": notification.Author})

	if err != nil {
		return err
	}

	for _, f := range followers {
		recipient := f.GetString("follower")
		SendNotification(app, notification, recipient)

	}
	return nil
}
