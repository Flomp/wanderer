package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"pocketbase/models"
	"strings"
	"time"

	pub "github.com/go-ap/activitypub"
	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/security"
	"github.com/valyala/fastjson"
)

func AddCustomTypesToPub() {
	pub.ItemTyperFunc = func(typ pub.ActivityVocabularyType) (pub.Item, error) {
		if typ == models.TrailType {
			return models.TrailNew(), nil
		} else if typ == models.SummitLogType {
			return models.SummitLogNew(), nil
		}
		return pub.GetItemByType(typ)
	}
	pub.JSONItemUnmarshal = func(typ pub.ActivityVocabularyType, v *fastjson.Value, i pub.Item) error {
		if typ == models.TrailType {
			return models.OnTrail(i, func(t *models.Trail) error {
				return models.JSONLoadTrail(v, t)
			})
		} else if typ == models.SummitLogType {
			return models.OnSummitLog(i, func(s *models.SummitLog) error {
				return models.JSONLoadSummitLog(v, s)
			})
		}
		return nil
	}
	pub.IsNotEmpty = func(i pub.Item) bool {
		if i.GetType() == models.TrailType || i.GetType() == models.SummitLogType {
			return true
		}

		return pub.NotEmpty(i)
	}
}

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
		if activity.Object.GetType() == models.TrailType {
			err = IndexActivity(*activity, app, actor)
			if err != nil {
				return err
			}
		}
	}

	if page.Next != nil {
		return fetchOutboxPage(app, actor, page.Next.GetID().String())
	}

	return nil
}

func IndexActivity(activity pub.Activity, app core.App, actor *core.Record) error {
	client := meilisearch.New(
		os.Getenv("MEILI_URL"),
		meilisearch.WithAPIKey(os.Getenv("MEILI_MASTER_KEY")),
	)

	trailObject, err := models.ToTrail(activity.Object)
	if err != nil {
		return err
	}

	doc, err := DocumentFromActivity(app, trailObject, actor)
	if err != nil {
		return err
	}
	documents := []map[string]interface{}{doc}

	if _, err := client.Index("trails").AddDocuments(documents); err != nil {
		return err
	}
	return nil
}
