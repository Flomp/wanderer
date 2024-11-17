package migrations

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db)
		settings, err := dao.FindSettings()

		if err != nil {
			return err
		}

		settings.Meta.AppName = "wanderer"
		settings.Meta.ResetPasswordTemplate.ActionUrl = "{APP_URL}/auth/confirm-reset/{TOKEN}"
		settings.Meta.VerificationTemplate.ActionUrl = "{APP_URL}/auth/confirm-verification/{TOKEN}"
		settings.Meta.ConfirmEmailChangeTemplate.ActionUrl = "{APP_URL}/auth/confirm-email-change/{TOKEN}"

		//   settings.meta.verificationTemplate.subject = 'your default subject';
		//   settings.meta.verificationTemplate.body = '<p>Hello world - {ACTION_URL}!</p>';
		//   settings.meta.verificationTemplate.actionUrl = '{APP_URL}/_/#/auth/confirm-verification/{TOKEN}';

		// you can also change similarly any other template or setting
		//
		// all setting fields are identical to the json response so you can
		// use the https://pocketbase.io/docs/api-settings/ > Update settings
		// "Body parameters" section as a reference

		return dao.SaveSettings(settings)

	}, func(db dbx.Builder) error {
		// add down queries...

		return nil
	})
}
