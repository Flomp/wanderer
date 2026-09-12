package federation

import (
	"bytes"
	"context"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	pub "github.com/go-ap/activitypub"

	"sync"

	"github.com/go-ap/jsonld"
	"github.com/go-fed/httpsig"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/security"
	"golang.org/x/sync/semaphore"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

// followerInboxes returns inbox URLs for all accepted followers of actorId
// in a single JOIN query instead of one query per follower.
func followerInboxes(app core.App, actorId string) ([]string, error) {
	rows, err := app.DB().
		Select("aa.inbox").
		From("follows f").
		InnerJoin("activitypub_actors aa", dbx.NewExp("f.follower = aa.id")).
		Where(dbx.NewExp("f.followee = {:followee} AND f.status = 'accepted' AND aa.inbox != ''",
			dbx.Params{"followee": actorId})).
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inboxes []string
	for rows.Next() {
		var inbox string
		if err := rows.Scan(&inbox); err != nil {
			return nil, err
		}
		inboxes = append(inboxes, inbox)
	}
	return inboxes, rows.Err()
}

// deleteRecipientInboxes returns the inboxes to notify when actorId is deleted:
//   - accepted followers
//   - actors it follows, any status
//   - authors of remote trails it commented on, logged a summit on, or liked
//   - remote actors who commented on, logged a summit on, or liked its trails
//   - the other party of every trail or list share it is involved in
func deleteRecipientInboxes(app core.App, actorId string) ([]string, error) {
	rows, err := app.DB().NewQuery(`
		SELECT aa.inbox
		FROM follows f
		INNER JOIN activitypub_actors aa ON f.follower = aa.id
		WHERE f.followee = {:actor} AND f.status = 'accepted' AND aa.inbox != ''
		UNION
		SELECT aa.inbox
		FROM follows f
		INNER JOIN activitypub_actors aa ON f.followee = aa.id
		WHERE f.follower = {:actor} AND aa.is_local = 0 AND aa.inbox != ''
		UNION
		SELECT aa.inbox
		FROM comments c
		INNER JOIN trails t ON c.trail = t.id
		INNER JOIN activitypub_actors aa ON t.author = aa.id
		WHERE c.author = {:actor} AND aa.is_local = 0 AND aa.inbox != ''
		UNION
		SELECT aa.inbox
		FROM summit_logs s
		INNER JOIN trails t ON s.trail = t.id
		INNER JOIN activitypub_actors aa ON t.author = aa.id
		WHERE s.author = {:actor} AND aa.is_local = 0 AND aa.inbox != ''
		UNION
		SELECT aa.inbox
		FROM trail_like l
		INNER JOIN trails t ON l.trail = t.id
		INNER JOIN activitypub_actors aa ON t.author = aa.id
		WHERE l.actor = {:actor} AND aa.is_local = 0 AND aa.inbox != ''
		UNION
		SELECT aa.inbox
		FROM comments c
		INNER JOIN trails t ON c.trail = t.id
		INNER JOIN activitypub_actors aa ON c.author = aa.id
		WHERE t.author = {:actor} AND aa.is_local = 0 AND aa.inbox != ''
		UNION
		SELECT aa.inbox
		FROM summit_logs s
		INNER JOIN trails t ON s.trail = t.id
		INNER JOIN activitypub_actors aa ON s.author = aa.id
		WHERE t.author = {:actor} AND aa.is_local = 0 AND aa.inbox != ''
		UNION
		SELECT aa.inbox
		FROM trail_like l
		INNER JOIN trails t ON l.trail = t.id
		INNER JOIN activitypub_actors aa ON l.actor = aa.id
		WHERE t.author = {:actor} AND aa.is_local = 0 AND aa.inbox != ''
		UNION
		SELECT aa.inbox
		FROM trail_share ts
		INNER JOIN trails t ON ts.trail = t.id
		INNER JOIN activitypub_actors aa ON aa.id = ts.actor
		WHERE t.author = {:actor} AND aa.is_local = 0 AND aa.inbox != ''
		UNION
		SELECT aa.inbox
		FROM trail_share ts
		INNER JOIN trails t ON ts.trail = t.id
		INNER JOIN activitypub_actors aa ON aa.id = t.author
		WHERE ts.actor = {:actor} AND aa.is_local = 0 AND aa.inbox != ''
		UNION
		SELECT aa.inbox
		FROM list_share ls
		INNER JOIN lists li ON ls.list = li.id
		INNER JOIN activitypub_actors aa ON aa.id = ls.actor
		WHERE li.author = {:actor} AND aa.is_local = 0 AND aa.inbox != ''
		UNION
		SELECT aa.inbox
		FROM list_share ls
		INNER JOIN lists li ON ls.list = li.id
		INNER JOIN activitypub_actors aa ON aa.id = li.author
		WHERE ls.actor = {:actor} AND aa.is_local = 0 AND aa.inbox != ''
	`).Bind(dbx.Params{"actor": actorId}).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inboxes []string
	for rows.Next() {
		var inbox string
		if err := rows.Scan(&inbox); err != nil {
			return nil, err
		}
		inboxes = append(inboxes, inbox)
	}
	return inboxes, rows.Err()
}

func PostActivity(app core.App, actor *core.Record, activity *pub.Activity, recipients []string) error {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				app.Logger().Error(fmt.Sprintf("Recovered from panic in PostActivity: %v", r))
			}
		}()

		encryptionKey := os.Getenv("POCKETBASE_ENCRYPTION_KEY")
		if len(encryptionKey) == 0 {
			app.Logger().Error("POCKETBASE_ENCRYPTION_KEY not set")
			return
		}
		origin := os.Getenv("ORIGIN")
		if origin == "" {
			app.Logger().Error("ORIGIN not set")
			return
		}

		algs := []httpsig.Algorithm{httpsig.RSA_SHA256}
		postHeaders := []string{"(request-target)", "Date", "Digest", "Content-Type", "Host"}
		expiresIn := 60

		body, err := jsonld.WithContext(
			jsonld.IRI(pub.ActivityBaseURI),
			jsonld.IRI(pub.SecurityContextURI),
		).Marshal(activity)
		if err != nil {
			app.Logger().Error(fmt.Sprintf("Failed to marshal activity: %s", err))
			return
		}

		decryptedPrivateKey, err := security.Decrypt(actor.GetString("private_key"), encryptionKey)
		if err != nil {
			app.Logger().Error(fmt.Sprintf("Failed to decrypt key: %s", err))
			return
		}
		privateKey, err := x509.ParsePKCS1PrivateKey(decryptedPrivateKey)
		if err != nil {
			app.Logger().Error(fmt.Sprintf("Failed to parse private key: %s", err))
			return
		}
		pubID := actor.GetString("iri") + "#main-key"

		sem := semaphore.NewWeighted(5)

		slices.Sort(recipients)
		uniqueRecipients := slices.Compact(recipients)

		var wg sync.WaitGroup
		for _, v := range uniqueRecipients {
			wg.Add(1)
			go func(inbox string) {
				defer wg.Done()

				signer, _, err := httpsig.NewSigner(algs, httpsig.DigestSha256, postHeaders, httpsig.Signature, int64(expiresIn))
				if err != nil {
					app.Logger().Error(fmt.Sprintf("Signer creation failed: %s", err))
					return
				}

				if err := sem.Acquire(context.Background(), 1); err != nil {
					app.Logger().Error(fmt.Sprintf("Semaphore acquire failed: %s", err))
					return
				}
				defer sem.Release(1)

				req, err := http.NewRequest(http.MethodPost, inbox, bytes.NewBuffer(body))
				if err != nil {
					app.Logger().Error(fmt.Sprintf("Request creation failed: %s", err))
					return
				}
				req.Header.Add("Content-Type", "application/activity+json")
				req.Header.Add("Date", strings.ReplaceAll(time.Now().UTC().Format(time.RFC1123), "UTC", "GMT"))
				req.Header.Add("Host", req.Host)

				if err := signer.SignRequest(privateKey, pubID, req, body); err != nil {
					app.Logger().Error(fmt.Sprintf("Signing request failed: %s", err))
					return
				}

				resp, err := httpClient.Do(req)
				if err != nil {
					app.Logger().Error(fmt.Sprintf("Error sending to inbox %s: %s", inbox, err))
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
					respBody, _ := io.ReadAll(resp.Body)
					app.Logger().Error(fmt.Sprintf("Inbox %s responded with %d: %s", inbox, resp.StatusCode, respBody))
				} else {
					app.Logger().Info(fmt.Sprintf("Sent %s to %s", activity.Type, inbox), "activity", activity)
				}
			}(v)
		}
		wg.Wait()
	}()
	return nil
}
