package federation

import (
	"bytes"
	"context"
	"crypto/x509"
	"database/sql"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	pub "github.com/go-ap/activitypub"

	"sync"

	"github.com/go-ap/jsonld"
	"github.com/go-fed/httpsig"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/security"
	"golang.org/x/sync/semaphore"
)

func PostActivity(app core.App, actor *core.Record, activity *pub.Activity, recipients []string) error {
	encryptionKey := os.Getenv("POCKETBASE_ENCRYPTION_KEY")
	if len(encryptionKey) == 0 {
		return fmt.Errorf("POCKETBASE_ENCRYPTION_KEY not set")
	}

	algs := []httpsig.Algorithm{httpsig.RSA_SHA256}
	postHeaders := []string{"(request-target)", "Date", "Digest"}
	expiresIn := 60

	signer, _, err := httpsig.NewSigner(algs, httpsig.DigestSha256, postHeaders, httpsig.Signature, int64(expiresIn))
	if err != nil {
		return err
	}

	body, err := jsonld.WithContext(
		jsonld.IRI(pub.ActivityBaseURI),
		jsonld.IRI(pub.SecurityContextURI),
	).Marshal(activity)
	if err != nil {
		return err
	}

	decryptedPrivateKey, err := security.Decrypt(actor.GetString("private_key"), encryptionKey)
	if err != nil {
		return err
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(decryptedPrivateKey)
	if err != nil {
		return err
	}
	pubID := actor.GetString("iri") + "#main-key"

	client := &http.Client{}
	alreadySentTo := make(map[string]bool)
	var mu sync.Mutex
	var wg sync.WaitGroup

	sem := semaphore.NewWeighted(5) // Limit to 5 concurrent sends

	for _, v := range recipients {
		if alreadySentTo[v] {
			continue
		}

		wg.Add(1)
		go func(inbox string) {
			defer wg.Done()

			if err := sem.Acquire(context.Background(), 1); err != nil {
				app.Logger().Error(fmt.Sprintf("Semaphore acquire failed: %s", err))
				return
			}
			defer sem.Release(1)

			buf := bytes.NewBuffer(body)
			req, err := http.NewRequest(http.MethodPost, inbox, buf)
			if err != nil {
				app.Logger().Error(fmt.Sprintf("Request creation failed: %s", err))
				return
			}
			req.Header.Add("Content-Type", `application/ld+json; profile="https://www.w3.org/ns/activitystreams"`)
			req.Header.Add("Date", strings.ReplaceAll(time.Now().UTC().Format(time.RFC1123), "UTC", "GMT"))

			if err := signer.SignRequest(privateKey, pubID, req, body); err != nil {
				app.Logger().Error(fmt.Sprintf("Signing request failed: %s", err))
				return
			}

			resp, err := client.Do(req)
			if err != nil {
				app.Logger().Error(fmt.Sprintf("Error sending request to inbox %s: %s", inbox, err))
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
				body, _ := io.ReadAll(resp.Body)
				app.Logger().Error(fmt.Sprintf("Inbox %s responded with %d: %s", inbox, resp.StatusCode, body))
			}

			app.Logger().Info(fmt.Sprintf("Sent %s to %s", activity.Type, inbox))

			mu.Lock()
			alreadySentTo[inbox] = true
			mu.Unlock()
		}(v)
	}

	wg.Wait()
	return nil
}

func ProcessActivity(e *core.RequestEvent) error {

	body, err := io.ReadAll(e.Request.Body)
	if err != nil {
		return err
	}
	var activity pub.Activity
	activity.UnmarshalJSON(body)

	actor, err := e.App.FindFirstRecordByData("activitypub_actors", "iri", activity.Actor.GetID().String())
	if err != nil {
		if err == sql.ErrNoRows {
			actor, _, err = GetActorByIRI(e.App, activity.Actor.GetID().String(), false)
			if err != nil {
				return err
			}
		} else {
			return err

		}
	}

	verified, err := verifySignature(e.App, e.Request, actor.GetString("public_key"))
	if err != nil || !verified {
		return e.UnauthorizedError("Invalid http signature", err)
	}

	switch activity.Type {
	case pub.FollowType:
		ProcessFollowActivity(e.App, actor, activity)
	case pub.AcceptType:
		ProcessAcceptActivity(e.App, actor, activity)
	case pub.UndoType:
		ProcessUnfollowActivity(e.App, actor, activity)
	case pub.UpdateType:
		fallthrough
	case pub.CreateType:
		ProcessCreateOrUpdateActivity(e.App, actor, activity)
	case pub.DeleteType:
		ProcessDeleteActivity(e.App, actor, activity)
	}
	return e.JSON(http.StatusOK, nil)
}

func verifySignature(app core.App, req *http.Request, publicKeyPem string) (bool, error) {
	block, _ := pem.Decode([]byte(publicKeyPem))
	if block == nil || block.Type != "PUBLIC KEY" {
		return false, fmt.Errorf("could not decode publicKeyPem to PUBLIC KEY pem block type")
	}

	if reqHeadersBytes, err := json.Marshal(req.Header); err == nil {
		app.Logger().Info(string(reqHeadersBytes))
	}
	app.Logger().Info(publicKeyPem)

	req.URL = &url.URL{
		Path: req.Header.Get("X-Forwarded-Path"),
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false, err
	}

	v, err := httpsig.NewVerifier(req)
	if err != nil {
		return false, err
	}

	err = v.Verify(publicKey, httpsig.RSA_SHA256)
	if err != nil {
		return false, err
	}

	return true, nil
}
