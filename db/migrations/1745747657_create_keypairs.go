package migrations

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"

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
			collection, err := app.FindCollectionByNameOrId("activitypub")
			if err != nil {
				return err
			}
			priv, pub, err := generateKeyPair()
			if err != nil {
				return err
			}
			privEncrypted, err := security.Encrypt(x509.MarshalPKCS1PrivateKey(priv), encryptionKey)
			if err != nil {
				return err
			}
			pubBytes, err := x509.MarshalPKIXPublicKey(pub)
			if err != nil {
				return err
			}

			record := core.NewRecord(collection)

			pemBlock := pem.EncodeToMemory(&pem.Block{
				Type:  "PUBLIC KEY",
				Bytes: pubBytes,
			})

			record.Set("public_key", string(pemBlock))
			record.Set("private_key", privEncrypted)
			record.Set("user", u.Id)

			app.Save(record)

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
