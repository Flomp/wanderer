package migrations

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/security"
)

func init() {
	m.Register(func(app core.App) error {
		users, err := app.FindAllRecords("users")

		encryptionKey := os.Getenv("POCKETBASE_ENCRYPTION_KEY")
		if len(encryptionKey) == 0 {
			return errors.New("POCKETBASE_ENCRYPTION_KEY not set")
		}

		if err != nil {
			return err
		}

		for _, u := range users {
			collection, err := app.FindCollectionByNameOrId("activitypub_actors")
			if err != nil {
				return err
			}
			priv, pub, err := generateKeyPair()
			if err != nil {
				return err
			}
			privBytes := x509.MarshalPKCS1PrivateKey(priv)

			privEncrypted, err := security.Encrypt(privBytes, encryptionKey)
			if err != nil {
				return err
			}

			pubBytes, err := x509.MarshalPKIXPublicKey(pub)
			if err != nil {
				return err
			}
			pubPem := pem.EncodeToMemory(&pem.Block{
				Type:  "PUBLIC KEY",
				Bytes: pubBytes,
			})

			settings, err := app.FindFirstRecordByData("settings", "user", u.Id)
			if err != nil {
				return err
			}

			record := core.NewRecord(collection)

			origin := os.Getenv("ORIGIN")
			if origin == "" {
				return fmt.Errorf("ORIGIN environment variable not set")
			}
			id := fmt.Sprintf("%s/api/v1/activitypub/user/%s", origin, u.GetString("username"))

			url, err := url.Parse(origin)
			if err != nil {
				return err
			}
			domain := strings.TrimPrefix(url.Hostname(), "www.")

			record.Set("username", u.GetString("username"))
			record.Set("domain", domain)
			record.Set("summary", settings.GetString("bio"))
			record.Set("published", u.GetDateTime("created"))
			record.Set("iri", id)
			record.Set("icon", fmt.Sprintf("%s/api/v1/files/users/%s/%s", origin, u.Id, u.GetString("avatar")))
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
				return err
			}

		}

		return nil
	}, func(app core.App) error {
		// add down queries...

		return nil
	})
}

func generateKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	pub := &priv.PublicKey
	return priv, pub, nil
}
