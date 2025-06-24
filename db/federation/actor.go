package federation

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	pub "github.com/go-ap/activitypub"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

type WebfingerResponse struct {
	Subject string `json:"subject"`
	Links   []struct {
		Rel  string `json:"rel"`
		Href string `json:"href"`
	} `json:"links"`
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

func GetActorByHandle(app core.App, handle string, includeFollows bool) (*core.Record, error) {
	username, domain := SplitHandle(handle)

	filter := "preferred_username={:username}&&"
	if domain != "" {
		filter += "domain={:domain}"
	} else {
		filter += "isLocal=true"
	}

	var dbActor *core.Record
	dbActor, err := app.FindFirstRecordByFilter("activitypub_actors", filter, dbx.Params{"username": username, "domain": domain})
	if err != nil && err == sql.ErrNoRows {
		collection, err := app.FindCollectionByNameOrId("activitypub_actors")
		if err != nil {
			return nil, err
		}

		dbActor = core.NewRecord(collection)
		dbActor.Set("isLocal", false)
		iri, err := iriFromHandle(domain, username)
		if err != nil {
			return nil, err
		}
		dbActor.Set("iri", iri)

	} else if err != nil {
		return nil, err
	}

	return assembleActor(dbActor, app, includeFollows)
}

func GetActorByIRI(app core.App, iri string, includeFollows bool) (*core.Record, error) {
	var dbActor *core.Record
	dbActor, err := app.FindFirstRecordByFilter("activitypub_actors", "iri={:iri}", dbx.Params{"iri": iri})
	if err != nil && err == sql.ErrNoRows {
		collection, err := app.FindCollectionByNameOrId("activitypub_actors")
		if err != nil {
			return nil, err
		}

		dbActor = core.NewRecord(collection)
		dbActor.Set("isLocal", false)
		dbActor.Set("iri", iri)

	} else if err != nil {
		return nil, err
	}

	return assembleActor(dbActor, app, includeFollows)
}

func iriFromHandle(domain string, username string) (string, error) {
	client := &http.Client{}

	webfingerURL := fmt.Sprintf("https://%s/.well-known/webfinger?resource=acct:%s@%s", domain, username, domain)
	resp, err := client.Get(webfingerURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("webfinger request failed: %v", err)
	}
	defer resp.Body.Close()

	var wf WebfingerResponse
	if err := json.NewDecoder(resp.Body).Decode(&wf); err != nil {
		return "", err
	}

	for _, link := range wf.Links {
		if link.Rel == "self" {
			return link.Href, nil
		}
	}
	return "", fmt.Errorf("no iri in response")
}

func assembleActor(dbActor *core.Record, app core.App, includeFollows bool) (*core.Record, error) {
	origin := os.Getenv("ORIGIN")
	if origin == "" {
		return nil, fmt.Errorf("ORIGIN environment variable not set")
	}

	private := false
	if dbActor.GetBool("isLocal") {
		user, err := app.FindRecordById("users", dbActor.GetString("user"))
		if err != nil {
			return nil, err
		}
		settings, err := app.FindFirstRecordByData("settings", "user", user.Id)
		if err != nil {
			return nil, err
		}

		if user.GetString("avatar") != "" {
			dbActor.Set("icon", fmt.Sprintf("%s/api/v1/files/users/%s/%s", origin, user.Id, user.GetString("avatar")))
		}
		dbActor.Set("summary", settings.GetString("bio"))
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

		privacy := settings.GetString("privacy")
		result := make(map[string]interface{})
		json.Unmarshal([]byte(privacy), &result)

		private = result["account"] == "private"

	} else {

		// check if value is still cached
		twoHoursAgo := time.Now().Add(-2 * time.Hour)
		if !includeFollows && dbActor.GetDateTime("last_fetched").Time().After(twoHoursAgo) {
			return dbActor, nil
		}
		pubActor, followers, following, err := fetchRemoteActor(dbActor.GetString("iri"), includeFollows)
		if err != nil {
			if dbActor.Id != "" {
				return dbActor, err
			}
			return nil, err
		}

		icon := ""
		if pub.IsObject(pubActor.Icon) {
			iconObject, err := pub.ToObject(pubActor.Icon)
			if err == nil && iconObject.URL != nil {
				icon = iconObject.URL.GetID().String()
			}
		}

		parsedUrl, err := url.Parse(dbActor.GetString("iri"))
		if err != nil {
			return nil, err
		}
		domain := strings.TrimPrefix(parsedUrl.Hostname(), "www.")

		dbActor.Set("domain", domain)
		dbActor.Set("followers", pubActor.Followers.GetID().String())
		dbActor.Set("inbox", pubActor.Inbox.GetID().String())
		dbActor.Set("iri", pubActor.GetID().String())
		dbActor.Set("username", pubActor.Name.String())
		dbActor.Set("preferred_username", pubActor.PreferredUsername.String())
		dbActor.Set("following", pubActor.Following.GetID().String())
		dbActor.Set("summary", pubActor.Summary.String())
		dbActor.Set("outbox", pubActor.Outbox.GetID().String())
		dbActor.Set("icon", icon)
		dbActor.Set("published", pubActor.Published.String())
		dbActor.Set("public_key", pubActor.PublicKey.PublicKeyPem)
		dbActor.Set("last_fetched", time.Now())

		if includeFollows {
			dbActor.Set("followerCount", int(followers.TotalItems))
			dbActor.Set("followingCount", int(following.TotalItems))
		}
	}

	err := app.Save(dbActor)
	if err != nil && err.Error() == "iri: Value must be unique." {
		dbActor, err = app.FindFirstRecordByData("activitypub_actors", "iri", dbActor.GetString("iri"))
		if err != nil {
			return nil, err
		}
		return dbActor, nil
	} else if err != nil {
		return nil, err
	}

	if private {
		return dbActor, fmt.Errorf("profile is private")
	}

	return dbActor, nil
}

// Fetches an AP actor and optionally followers/following collections
func fetchRemoteActor(iri string, includeFollows bool) (*pub.Actor, *pub.OrderedCollection, *pub.OrderedCollection, error) {
	client := &http.Client{}
	headers := map[string]string{
		"Accept": "application/ld+json",
	}

	req, _ := http.NewRequest("GET", iri, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("actor fetch failed: %v", err)
	} else if resp.StatusCode != http.StatusOK {
		return nil, nil, nil, fmt.Errorf("actor fetch failed: status %v", resp.StatusCode)
	}

	defer resp.Body.Close()

	var pubActor pub.Actor
	if err := json.NewDecoder(resp.Body).Decode(&pubActor); err != nil {
		return nil, nil, nil, err
	}

	var followers, following pub.OrderedCollection

	if includeFollows {
		// Fetch followers
		if data, err := fetchCollection(pubActor.Followers.GetID().String(), headers); err == nil {
			followers = *data
		}

		// Fetch following
		if data, err := fetchCollection(pubActor.Following.GetID().String(), headers); err == nil {
			following = *data
		}
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
