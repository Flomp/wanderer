package federation

import (
	"bytes"
	"crypto/x509"
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
	"github.com/valyala/fastjson"

	"github.com/go-ap/jsonld"
	"github.com/go-fed/httpsig"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/security"
)

func PostActivity(app core.App, actor *core.Record, activity *pub.Activity, recipients []string) error {
	encryptionKey := os.Getenv("POCKETBASE_ENCRYPTION_KEY")
	if len(encryptionKey) == 0 {
		return fmt.Errorf("POCKETBASE_ENCRYPTION_KEY not set")
	}

	algs := []httpsig.Algorithm{httpsig.RSA_SHA256}
	postHeaders := []string{"(request-target)", "Date", "Digest"}
	expiresIn := 60

	// create HTTP signer
	signer, _, err := httpsig.NewSigner(algs, httpsig.DigestSha256, postHeaders, httpsig.Signature, int64(expiresIn))
	if err != nil {
		return err
	}

	// marshal activity
	body, err := jsonld.WithContext(
		jsonld.IRI(pub.ActivityBaseURI),
		jsonld.IRI(pub.SecurityContextURI),
	).Marshal(activity)
	if err != nil {
		return err
	}

	client := &http.Client{}

	// for each recipient
	for _, v := range recipients {
		buf := bytes.NewBuffer(body)
		req, err := http.NewRequest(http.MethodPost, v, buf)
		if err != nil {
			return err
		}
		req.Header.Add("Content-Type", `application/ld+json; profile="https://www.w3.org/ns/activitystreams"`)

		currentTime := strings.ReplaceAll(time.Now().UTC().Format(time.RFC1123), "UTC", "GMT")
		req.Header.Add("Date", currentTime)

		// sign request header
		encryptedPrivateKey := actor.GetString("private_key")
		decryptedPrivateKey, err := security.Decrypt(encryptedPrivateKey, encryptionKey)
		if err != nil {
			return err
		}

		privateKey, err := x509.ParsePKCS1PrivateKey(decryptedPrivateKey)
		if err != nil {
			return err
		}

		pubID := actor.GetString("iri") + "#main-key"
		err = signer.SignRequest(privateKey, pubID, req, body)
		if err != nil {
			return err
		}

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("error sending request to inbox %s (%d): %s", v, resp.StatusCode, string(body))
		}
	}
	return nil
}

func ProcessActivity(e *core.RequestEvent) error {
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
	var activity pub.Activity
	if err := e.BindBody(&activity); err != nil {
		return e.BadRequestError("Failed to read request data", err)
	}

	actor, err := e.App.FindFirstRecordByData("activitypub_actors", "iri", activity.Actor)
	if err != nil {
		return err
	}

	verified, err := verifySignature(e.Request, actor.GetString("public_key"))
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

func verifySignature(req *http.Request, publicKeyPem string) (bool, error) {
	block, _ := pem.Decode([]byte(publicKeyPem))
	if block == nil || block.Type != "PUBLIC KEY" {
		return false, fmt.Errorf("could not decode publicKeyPem to PUBLIC KEY pem block type")
	}

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
