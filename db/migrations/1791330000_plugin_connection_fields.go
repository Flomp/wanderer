package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() { m.Register(addPluginConnectionFields1791330000, nil) }

// Both schema preparation and activation run before routes and schedulers start.
func addPluginConnectionFields1791330000(app core.App) error {
	collection, err := app.FindCollectionByNameOrId("plugin_instances")
	if err != nil {
		return err
	}
	for _, field := range []*core.TextField{
		{Name: "auth_context", Max: 256},
		{Name: "connection_generation", Max: 128},
		{Name: "oauth_token_revision", Max: 128, Hidden: true},
	} {
		if collection.Fields.GetByName(field.Name) == nil {
			collection.Fields.Add(field)
		}
	}
	return app.Save(collection)
}
