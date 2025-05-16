package util

import (
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
	"pocketbase/models"
	"strings"
	"time"

	pub "github.com/go-ap/activitypub"
	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/dbx"
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

func SplitHandle(handle string) (string, string) {

	cleaned := strings.TrimPrefix(handle, "@")
	cleaned = strings.TrimSpace(cleaned)

	if !strings.Contains(cleaned, "@") {
		return cleaned, ""
	}

	parts := strings.SplitN(cleaned, "@", 2)
	user := parts[0]
	domain := parts[1]

	return user, domain
}

type WebfingerResponse struct {
	Subject string `json:"subject"`
	Links   []struct {
		Rel  string `json:"rel"`
		Href string `json:"href"`
	} `json:"links"`
}

func GetActor(app core.App, handle string) (*core.Record, error) {
	origin := os.Getenv("ORIGIN")
	if origin == "" {
		return nil, fmt.Errorf("ORIGIN environment variable not set")
	}

	username, domain := SplitHandle(handle)

	if domain == "" {
		url, err := url.Parse(origin)
		if err != nil {
			return nil, err
		}
		domain = strings.TrimPrefix(url.Hostname(), "www.")

	}

	var dbActor *core.Record
	dbActor, err := app.FindFirstRecordByFilter("activitypub_actors", "username={:username}&&domain={:domain}", dbx.Params{"username": username, "domain": domain})
	if err != nil && err == sql.ErrNoRows {
		collection, err := app.FindCollectionByNameOrId("activitypub_actors")
		if err != nil {
			return nil, err
		}

		dbActor = core.NewRecord(collection)
		dbActor.Set("isLocal", false)
	} else if err != nil {
		return nil, err
	}

	if dbActor.GetBool("isLocal") {
		user, err := app.FindRecordById("users_anonymous", dbActor.GetString("user"))
		if err != nil {
			return nil, err
		}

		if user.GetString("avatar") != "" {
			dbActor.Set("icon", fmt.Sprintf("%s/api/v1/files/users/%s/%s", origin, user.Id, user.GetString("avatar")))
		}
		dbActor.Set("summary", user.GetString("bio"))
		followerCount, err := app.CountRecords("follows", dbx.NewExp("followee={:user}", dbx.Params{"user": dbActor.Id}))
		if err != nil {
			return nil, err
		}
		dbActor.Set("followerCount", followerCount)
		followingCount, err := app.CountRecords("follows", dbx.NewExp("follower={:user}", dbx.Params{"user": dbActor.Id}))
		if err != nil {
			return nil, err
		}
		dbActor.Set("followingCount", followingCount)

		dbActor.Set("last_fetched", time.Now())

	} else {
		pubActor, followers, following, err := fetchRemoteActor(dbActor.GetString("iri"), username, domain)
		if err != nil {
			return nil, err
		}

		dbActor.Set("domain", domain)
		dbActor.Set("followers", pubActor.Followers.GetID().String())
		dbActor.Set("inbox", pubActor.Inbox.GetID().String())
		dbActor.Set("iri", pubActor.GetID().String())
		dbActor.Set("username", pubActor.Name.String())
		dbActor.Set("followerCount", int(followers.TotalItems))
		dbActor.Set("followingCount", int(following.TotalItems))
		dbActor.Set("following", pubActor.Following.GetID().String())
		dbActor.Set("summary", pubActor.Summary.String())
		dbActor.Set("outbox", pubActor.Outbox.GetID().String())
		dbActor.Set("icon", pubActor.Icon.GetID().String())
		dbActor.Set("published", pubActor.Published.String())
		dbActor.Set("public_key", pubActor.PublicKey.PublicKeyPem)
		dbActor.Set("last_fetched", time.Now())
	}

	err = app.Save(dbActor)
	if err != nil {
		return nil, err
	}

	return dbActor, nil
}

// Fetches an AP actor and optionally followers/following collections
func fetchRemoteActor(iri string, username string, domain string) (*pub.Actor, *pub.OrderedCollection, *pub.OrderedCollection, error) {
	client := &http.Client{}
	headers := map[string]string{
		"Accept": "application/ld+json",
	}

	// Resolve actor URI via Webfinger if domain is provided
	if iri == "" {
		webfingerURL := fmt.Sprintf("https://%s/.well-known/webfinger?resource=acct:%s@%s", domain, username, domain)
		resp, err := client.Get(webfingerURL)
		if err != nil || resp.StatusCode != http.StatusOK {
			return nil, nil, nil, fmt.Errorf("webfinger request failed: %v", err)
		}
		defer resp.Body.Close()

		var wf WebfingerResponse
		if err := json.NewDecoder(resp.Body).Decode(&wf); err != nil {
			return nil, nil, nil, err
		}

		for _, link := range wf.Links {
			if link.Rel == "self" {
				iri = link.Href
				break
			}
		}
	}

	req, _ := http.NewRequest("GET", iri, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, nil, nil, fmt.Errorf("actor fetch failed: %v", err)
	}
	defer resp.Body.Close()

	var pubActor pub.Actor
	if err := json.NewDecoder(resp.Body).Decode(&pubActor); err != nil {
		return nil, nil, nil, err
	}

	var followers, following pub.OrderedCollection

	// Fetch followers
	if data, err := fetchCollection(pubActor.Followers.GetID().String(), headers); err == nil {
		followers = *data
	}

	// Fetch following
	if data, err := fetchCollection(pubActor.Following.GetID().String(), headers); err == nil {
		following = *data
	}

	return &pubActor, &followers, &following, nil
}

func fetchCollection(url string, headers map[string]string) (*pub.OrderedCollection, error) {
	req, _ := http.NewRequest("GET", url, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("collection fetch failed for %s: %v", url, err)
	}
	defer resp.Body.Close()

	var collection pub.OrderedCollection
	if err := json.NewDecoder(resp.Body).Decode(&collection); err != nil {
		return nil, err
	}

	return &collection, nil
}
