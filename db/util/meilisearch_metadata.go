package util

import (
	"errors"
	"fmt"
	"maps"

	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

// Only denormalized metadata is patched. In particular an actor rename must not
// rebuild remote list aggregates or overwrite independently updated grants.
func UpdateActorReferences(app core.App, actorID string, client meilisearch.ServiceManager) error {
	actor, err := app.FindRecordById("activitypub_actors", actorID)
	if err != nil {
		return err
	}
	domain := ""
	if !actor.GetBool("is_local") {
		domain = actor.GetString("domain")
	}
	for _, index := range []string{"trails", "lists"} {
		values := map[string]any{"author_name": actor.GetString("preferred_username"), "author_avatar": actor.GetString("icon"), "domain": domain}
		if index == "trails" {
			values["is_federated"] = !actor.GetBool("is_local")
		}
		err := updateRelatedMetadata(app, index, "author = {:actor}", dbx.Params{"actor": actorID}, client, func(*core.Record) (map[string]any, error) {
			return maps.Clone(values), nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateTagReferences(app core.App, tagID string, client meilisearch.ServiceManager) error {
	return updateRelatedMetadata(app, "trails", "tags.id ?= {:tag}", dbx.Params{"tag": tagID}, client, func(trail *core.Record) (map[string]any, error) {
		if errs := app.ExpandRecord(trail, []string{"tags"}, nil); len(errs) != 0 {
			return nil, fmt.Errorf("expand tags for trail %s: %v", trail.Id, errs)
		}
		tags := []string{}
		for _, tag := range trail.ExpandedAll("tags") {
			tags = append(tags, tag.GetString("name"))
		}
		return map[string]any{"tags": tags}, nil
	})
}

func UpdateCategoryReferences(app core.App, categoryID string, client meilisearch.ServiceManager) error {
	category, err := app.FindRecordById("categories", categoryID)
	if err != nil {
		return err
	}
	return updateRelatedMetadata(app, "trails", "category = {:category}", dbx.Params{"category": categoryID}, client, func(*core.Record) (map[string]any, error) {
		return map[string]any{"category": category.GetString("name"), "category_icon": category.GetString("icon")}, nil
	})
}

func updateRelatedMetadata(app core.App, index, filter string, params dbx.Params, client meilisearch.ServiceManager, project func(*core.Record) (map[string]any, error)) error {
	const batchSize = 200
	params["cursor"] = ""
	for {
		records, err := app.FindRecordsByFilter(index, "("+filter+") && id > {:cursor}", "id", batchSize, 0, params)
		if err != nil {
			return err
		}
		if len(records) == 0 {
			return nil
		}
		documents := make([]map[string]any, len(records))
		for i, record := range records {
			patch, err := project(record)
			if err != nil {
				return err
			}
			var existing map[string]any
			err = client.Index(index).GetDocument(record.Id, &meilisearch.DocumentQuery{Fields: []string{"id"}}, &existing)
			if err != nil {
				var apiError *meilisearch.Error
				if !errors.As(err, &apiError) || apiError.MeilisearchApiError.Code != "document_not_found" {
					return fmt.Errorf("read %s/%s before metadata update: %w", index, record.Id, err)
				}
				patch, err = missingMetadataDocument(app, index, record)
				if err != nil {
					return err
				}
			}
			patch["id"] = record.Id
			documents[i] = patch
		}
		if _, err := client.Index(index).UpdateDocuments(documents, nil); err != nil {
			return err
		}
		if len(records) < batchSize {
			return nil
		}
		params["cursor"] = records[len(records)-1].Id
	}
}

// A metadata patch is an upsert in Meilisearch. Never create a partial search
// hit when its previous document is absent. Local source data can reconstruct a
// complete document; remote list aggregates cannot be inferred here.
func missingMetadataDocument(app core.App, index string, record *core.Record) (map[string]any, error) {
	if index == "trails" {
		documents, err := trailSearchDocuments(app, []*core.Record{record})
		if err != nil {
			return nil, err
		}
		return documents[0], nil
	}
	if errs := app.ExpandRecord(record, []string{"author", "trails", "list_share_via_list"}, nil); len(errs) != 0 {
		return nil, fmt.Errorf("expand missing list %s: %v", record.Id, errs)
	}
	author := record.ExpandedOne("author")
	if author == nil {
		return nil, fmt.Errorf("missing list %s has no author", record.Id)
	}
	if record.GetString("iri") != "" && !author.GetBool("is_local") {
		return nil, fmt.Errorf("missing remote list index document %s requires complete materialization before metadata updates", record.Id)
	}
	return documentFromListRecord(record, author, true)
}
