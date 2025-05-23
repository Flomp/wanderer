package federation

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
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

func GetActor(app core.App, handle string) (*core.Record, bool, error) {
	origin := os.Getenv("ORIGIN")
	if origin == "" {
		return nil, false, fmt.Errorf("ORIGIN environment variable not set")
	}

	username, domain := SplitHandle(handle)

	var dbActor *core.Record
	dbActor, err := app.FindFirstRecordByFilter("activitypub_actors", "username={:username}&&(domain={:domain}||isLocal=true)", dbx.Params{"username": username, "domain": domain})
	if err != nil && err == sql.ErrNoRows {
		collection, err := app.FindCollectionByNameOrId("activitypub_actors")
		if err != nil {
			return nil, false, err
		}

		dbActor = core.NewRecord(collection)
		dbActor.Set("isLocal", false)
	} else if err != nil {
		return nil, false, err
	}

	private := false
	if dbActor.GetBool("isLocal") {
		user, err := app.FindRecordById("users_anonymous", dbActor.GetString("user"))
		if err != nil {
			return nil, false, err
		}

		if user.GetString("avatar") != "" {
			dbActor.Set("icon", fmt.Sprintf("%s/api/v1/files/users/%s/%s", origin, user.Id, user.GetString("avatar")))
		}
		dbActor.Set("summary", user.GetString("bio"))
		followerCount, err := app.CountRecords("follows", dbx.NewExp("followee={:user}", dbx.Params{"user": dbActor.Id}))
		if err != nil {
			return nil, false, err
		}
		dbActor.Set("followerCount", followerCount)
		followingCount, err := app.CountRecords("follows", dbx.NewExp("follower={:user}", dbx.Params{"user": dbActor.Id}))
		if err != nil {
			return nil, false, err
		}
		dbActor.Set("followingCount", followingCount)

		dbActor.Set("last_fetched", time.Now())

		settings, err := app.FindFirstRecordByData("settings", "user", user.Id)
		if err != nil {
			return nil, false, err
		}
		privacy := settings.GetString("privacy")
		result := make(map[string]interface{})
		json.Unmarshal([]byte(privacy), &result)

		private = result["account"] == "private"

	} else {
		pubActor, followers, following, err := fetchRemoteActor(dbActor.GetString("iri"), username, domain)
		if err != nil {
			if dbActor.Id != "" {
				return dbActor, false, err
			}
			return nil, false, err
		}

		dbActor.Set("domain", domain)
		dbActor.Set("followers", pubActor.Followers.GetID().String())
		dbActor.Set("inbox", pubActor.Inbox.GetID().String())
		dbActor.Set("iri", pubActor.GetID().String())
		dbActor.Set("username", pubActor.Name.String())
		dbActor.Set("preferred_username", pubActor.PreferredUsername.String())
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
	if err != nil && err.Error() != "iri: Value must be unique." {
		return nil, false, err
	}

	return dbActor, private, nil
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
