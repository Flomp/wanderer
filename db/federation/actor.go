package federation

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	pub "github.com/go-ap/activitypub"
	"github.com/go-ap/jsonld"
	"github.com/pocketbase/pocketbase/core"
)

func actor(e *core.RequestEvent) error {
	origin := os.Getenv("ORIGIN")
	if origin == "" {
		return e.InternalServerError("ORIGIN not set", nil)
	}

	username := e.Request.PathValue("username")

	if username == "" {
		return e.BadRequestError("Missing param: username", nil)
	}

	username = strings.TrimPrefix(username, "@")

	user, err := e.App.FindFirstRecordByData("users", "username", username)
	if err != nil {
		return err
	}

	key, err := e.App.FindFirstRecordByData("activitypub_actors", "user", user.Id)
	if err != nil {
		return err
	}
	publicKey := key.GetString("public_key")

	id := fmt.Sprintf("%s/api/v1/activitypub/user/%s", origin, username)
	profile := fmt.Sprintf("%s/profile/%s", origin, user.Id)
	actor := pub.Actor{
		Context: pub.IRI("https://www.w3.org/ns/activitystreams"),
		ID:      pub.IRI(id),
		Type:    "Person",
		Inbox:   pub.IRI(id + "/inbox"),
		Outbox:  pub.IRI(id + "/outbox"),
		Summary: pub.NaturalLanguageValues{{
			Ref:   pub.NilLangRef,
			Value: pub.Content(user.GetString("bio"))},
		},
		Name: pub.NaturalLanguageValues{{
			Ref:   pub.NilLangRef,
			Value: pub.Content(user.GetString("username"))},
		},
		PreferredUsername: pub.NaturalLanguageValues{{
			Ref:   pub.NilLangRef,
			Value: pub.Content(user.GetString("username"))},
		},
		Following: pub.IRI(profile + "/following"),
		Followers: pub.IRI(profile + "/followers"),
		URL:       pub.IRI(profile),
		Published: user.GetDateTime("created").Time(),
		Icon: &pub.Object{
			Type:      "Image",
			URL:       pub.IRI(fmt.Sprintf("%s/api/v1/files/users/%s/%s", origin, user.Id, user.GetString("avatar"))),
			MediaType: "image/jpeg",
		},
		PublicKey: pub.PublicKey{
			ID:           pub.IRI(id + "#main-key"),
			Owner:        pub.IRI(id),
			PublicKeyPem: publicKey,
		},
	}

	e.Response.Header().Add("Content-Type", jsonld.ContentType)
	e.Response.WriteHeader(http.StatusOK)

	binary, err := jsonld.WithContext(
		jsonld.IRI(pub.ActivityBaseURI),
		jsonld.IRI(pub.SecurityContextURI),
	).Marshal(actor)

	e.Response.Write(binary)
	return err
}
