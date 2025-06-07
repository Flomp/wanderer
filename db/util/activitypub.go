package util

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	pub "github.com/go-ap/activitypub"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"github.com/pocketbase/pocketbase/tools/security"
)

func ActorFromUser(app core.App, u *core.Record) (*core.Record, error) {
	encryptionKey := os.Getenv("POCKETBASE_ENCRYPTION_KEY")
	if len(encryptionKey) == 0 {
		return nil, fmt.Errorf("POCKETBASE_ENCRYPTION_KEY not set")
	}

	collection, err := app.FindCollectionByNameOrId("activitypub_actors")
	if err != nil {
		return nil, err
	}
	priv, pub, err := generateKeyPair()
	if err != nil {
		return nil, err
	}
	privBytes := x509.MarshalPKCS1PrivateKey(priv)

	privEncrypted, err := security.Encrypt(privBytes, encryptionKey)
	if err != nil {
		return nil, err
	}

	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, err
	}
	pubPem := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})

	settings, err := app.FindFirstRecordByData("settings", "user", u.Id)
	if err != nil {
		return nil, err
	}

	record := core.NewRecord(collection)

	origin := os.Getenv("ORIGIN")
	if origin == "" {
		return nil, fmt.Errorf("ORIGIN environment variable not set")
	}
	id := fmt.Sprintf("%s/api/v1/activitypub/user/%s", origin, strings.ToLower(u.GetString("username")))

	url, err := url.Parse(origin)
	if err != nil {
		return nil, err
	}
	domain := strings.TrimPrefix(url.Hostname(), "www.")

	record.Set("username", strings.ToLower(u.GetString("username")))
	record.Set("preferred_username", u.GetString("username"))
	record.Set("domain", domain)
	record.Set("summary", settings.GetString("bio"))
	record.Set("published", u.GetDateTime("created"))
	record.Set("iri", id)
	if u.GetString("avatar") != "" {
		record.Set("icon", fmt.Sprintf("%s/api/v1/files/users/%s/%s", origin, u.Id, u.GetString("avatar")))
	}
	record.Set("inbox", id+"/inbox")
	record.Set("outbox", id+"/outbox")
	record.Set("followers", id+"/followers")
	record.Set("following", id+"/following")
	record.Set("isLocal", true)
	record.Set("public_key", string(pubPem))
	record.Set("private_key", privEncrypted)
	record.Set("user", u.Id)
	record.Set("last_fetched", time.Now())

	err = app.Save(record)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func generateKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	pub := &priv.PublicKey
	return priv, pub, nil
}

func SyncOutbox(app core.App, actor *core.Record) error {
	return fetchOutboxPage(app, actor, actor.GetString("outbox")+"?page=1")
}

func fetchOutboxPage(app core.App, actor *core.Record, pageURL string) error {
	client := &http.Client{}

	req, err := http.NewRequest(http.MethodGet, pageURL, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", `application/ld+json; profile="https://www.w3.org/ns/activitystreams"`)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var page pub.OrderedCollectionPage
	err = json.Unmarshal(body, &page)
	if err != nil {
		return err
	}

	for _, item := range page.OrderedItems {
		activity, err := pub.ToActivity(item)
		if err != nil {
			return err
		}
		if activity.Type != pub.CreateType {
			continue
		}
	}

	if page.Next != nil {
		return fetchOutboxPage(app, actor, page.Next.GetID().String())
	}

	return nil
}

func TrailFromActivity(activity pub.Activity, app core.App, actor *core.Record) error {
	t, err := pub.ToObject(activity.Object)
	if err != nil {
		return err
	}

	record, err := app.FindFirstRecordByData("trails", "iri", t.ID.String())
	if err != nil {
		if err == sql.ErrNoRows {
			collection, err := app.FindCollectionByNameOrId("trails")
			if err != nil {
				return err
			}

			record = core.NewRecord(collection)
		} else {
			return err
		}
	}

	var distance, duration, elevation_gain, elevation_loss float64
	var diffculty, category string
	tags, err := pub.ToItemCollection(t.Tag)
	if err != nil {
		return err
	}

	for _, tag := range tags.Collection() {
		tagObj, err := pub.ToObject(tag)
		if err != nil {
			continue
		}
		content := tagObj.Content.First().Value.String()
		switch tagObj.Name.First().Value.String() {
		case "category":
			category = content
		case "difficulty":
			diffculty = content
		case "elevation_gain":
			elevation_gain, err = strconv.ParseFloat(content[:len(content)-1], 64)
		case "elevation_loss":
			elevation_loss, err = strconv.ParseFloat(content[:len(content)-1], 64)
		case "duration":
			duration, err = strconv.ParseFloat(content[:len(content)-1], 64)
		case "distance":
			distance, err = strconv.ParseFloat(content[:len(content)-1], 64)
		}
		if err != nil {
			continue
		}
	}

	record.Set("name", t.Name.First().Value)
	record.Set("description", t.Content.First().Value)
	record.Set("location", t.Location.(*pub.Place).Name.First().Value)
	record.Set("lat", t.Location.(*pub.Place).Latitude)
	record.Set("lon", t.Location.(*pub.Place).Longitude)
	record.Set("distance", distance)
	record.Set("elevation_gain", elevation_gain)
	record.Set("elevation_loss", elevation_loss)
	record.Set("duration", duration)
	record.Set("difficulty", diffculty)
	record.Set("date", t.StartTime.Unix())
	record.Set("public", true)
	record.Set("iri", t.ID.String())
	record.Set("author", actor.Id)

	categoryRecord, err := app.FindFirstRecordByData("categories", "name", category)
	if err == nil {
		record.Set("category", categoryRecord.Id)
	}

	if t.Attachment != nil {

		attachments, err := pub.ToItemCollection(t.Attachment)
		if err != nil {
			return err
		}

		photoURLs := []string{}
		gpxURL := ""
		for _, a := range attachments.Collection() {
			attachment, err := pub.ToObject(a)
			if err != nil {
				continue
			}
			if attachment.Type == pub.DocumentType && attachment.MediaType == "application/xml+gpx" {
				gpxURL = attachment.URL.GetLink().String()
			} else if attachment.Type == pub.ImageType {
				photoURLs = append(photoURLs, attachment.URL.GetLink().String())
			}
		}

		if len(photoURLs) > 0 {
			photos := make([]*filesystem.File, len(photoURLs))
			for i, purl := range photoURLs {
				photo, err := filesystem.NewFileFromURL(context.Background(), purl)
				if err != nil {
					continue
				}
				photos[i] = photo
			}

			record.Set("photos", photos)
		}

		if gpxURL != "" {
			gpx, err := filesystem.NewFileFromURL(context.Background(), gpxURL)
			if err != nil {
				return err
			}

			record.Set("gpx", gpx)
		}
	}

	return app.Save(record)
}

func ListFromActivity(activity pub.Activity, app core.App, actor *core.Record) error {
	l, err := pub.ToObject(activity.Object)
	if err != nil {
		return err
	}

	record, err := app.FindFirstRecordByData("lists", "iri", l.ID.String())
	if err != nil {
		if err == sql.ErrNoRows {
			collection, err := app.FindCollectionByNameOrId("lists")
			if err != nil {
				return err
			}

			record = core.NewRecord(collection)
		} else {
			return err
		}
	}

	record.Set("name", l.Name.First().Value)
	record.Set("description", l.Content.First().Value)
	record.Set("public", true)
	record.Set("iri", l.ID.String())
	record.Set("author", actor.Id)

	if l.Attachment != nil {

		avatarURL := ""
		attachments, err := pub.ToItemCollection(l.Attachment)
		if err != nil {
			return err
		}

		for _, a := range attachments.Collection() {
			attachment, err := pub.ToObject(a)
			if err != nil {
				continue
			}
			if attachment.Type == pub.ImageType {
				avatarURL = attachment.URL.GetLink().String()
			}
		}

		if avatarURL != "" {
			avatar, err := filesystem.NewFileFromURL(context.Background(), avatarURL)

			if err != nil {
				return err
			}

			record.Set("avatar", avatar)
		}
	}

	err = app.Save(record)

	return err
}
