package federation

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

type link struct {
	Rel      string `json:"rel,omitempty"`
	Type     string `json:"type,omitempty"`
	Href     string `json:"href,omitempty"`
	Template string `json:"template,omitempty"`
}

type node struct {
	Subject string   `json:"subject"`
	Aliases []string `json:"aliases"`
	Links   []link   `json:"links"`
}

func webfinger(e *core.RequestEvent) error {
	resource := e.Request.URL.Query().Get("resource")

	if resource == "" || !strings.HasPrefix(resource, "acct:") {
		return e.BadRequestError("Bad Request", nil)
	}

	origin := os.Getenv("ORIGIN")
	if origin == "" {
		return e.InternalServerError("ORIGIN not set", nil)
	}

	url, err := url.Parse(origin)
	if err != nil {
		log.Fatal(err)
	}
	hostname := strings.TrimPrefix(url.Hostname(), "www.")

	parts := strings.Split(strings.TrimPrefix(resource, "acct:"), "@")
	if len(parts) != 2 {
		return e.BadRequestError("Bad request", nil)
	}

	username := parts[0]
	domain := parts[1]

	if hostname != domain {
		return e.NotFoundError("Not found", nil)
	}

	user, err := e.App.FindFirstRecordByData("users", "username", username)
	if err != nil {
		return err
	}

	id := fmt.Sprintf("%s/api/v1/activitpub/user/@%s", origin, username)

	webfinger := node{
		Subject: resource,
		Aliases: []string{
			id,
		},
		Links: []link{
			{
				Rel:  "self",
				Type: "application/activity+json",
				Href: id,
			},
			{
				Rel:  "http://webfinger.net/rel/profile-page",
				Type: "text/html",
				Href: fmt.Sprintf("%s/profile/%s", origin, user.Id),
			},
		},
	}

	return e.JSON(http.StatusOK, webfinger)
}
