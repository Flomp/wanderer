package hooks

import (
	"fmt"
	"net/url"

	"pocketbase/federation"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

// announceTrailMadePrivate retracts a trail that just went from public to
// private, with the same Delete a deletion sends. A current share holder
// keeps their copy, so every inbox on that holder's host is left out.
func announceTrailMadePrivate(e *core.RecordEvent) error {
	return announceMadePrivate(e, "trail", federation.TrailDeleteRecipients, federation.CreateTrailDeleteActivity)
}

// announceListMadePrivate is announceTrailMadePrivate for a list.
func announceListMadePrivate(e *core.RecordEvent) error {
	return announceMadePrivate(e, "list", federation.ListDeleteRecipients, federation.CreateListDeleteActivity)
}

func announceMadePrivate(
	e *core.RecordEvent,
	kind string,
	recipients func(core.App, *core.Record, bool) (federation.DeleteAudience, error),
	send func(core.App, *core.Record, federation.DeleteAudience) error,
) error {
	record := e.Record
	// public=true: a trail that just went private still has to reach the
	// followers it was public to. TrailDeleteRecipients documents that.
	if !record.Original().GetBool("public") || record.GetBool("public") {
		return nil
	}

	audience, err := recipients(e.App, record, true)
	if err != nil {
		e.App.Logger().Error(
			"could not collect recipients to announce that a "+kind+" went private",
			kind, record.Id, "error", err,
		)
		return nil
	}

	// A share still grants access. processDeleteTrailActivity and
	// processDeleteListActivity delete the record for the whole server, so a
	// Delete to any inbox on the share holder's host would drop their copy.
	// If we cannot see who holds a share, skip the announcement rather than
	// guess.
	shared, err := shareHolderInboxes(e.App, kind, record.Id)
	if err != nil {
		e.App.Logger().Error(
			"could not collect recipients to announce that a "+kind+" went private",
			kind, record.Id, "error", err,
		)
		return nil
	}

	return send(e.App, record, withoutShareHolders(audience, shared))
}

func shareHolderInboxes(app core.App, kind, id string) ([]string, error) {
	var query string
	switch kind {
	case "trail":
		query = `
			SELECT aa.inbox
			FROM trail_share ts
			INNER JOIN activitypub_actors aa ON ts.actor = aa.id
			WHERE ts.trail = {:id} AND aa.inbox != ''`
	case "list":
		query = `
			SELECT aa.inbox
			FROM list_share ls
			INNER JOIN activitypub_actors aa ON ls.actor = aa.id
			WHERE ls.list = {:id} AND aa.inbox != ''`
	default:
		return nil, fmt.Errorf("unknown share subject %q", kind)
	}

	rows, err := app.DB().NewQuery(query).Bind(dbx.Params{"id": id}).Rows()
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

// withoutShareHolders drops every inbox whose host matches a current share
// holder's inbox. The host is url.Parse(inbox).Host (host:port). An inbox
// that does not parse, or that has no host, is kept or dropped by comparing
// the raw string.
func withoutShareHolders(audience federation.DeleteAudience, shared []string) federation.DeleteAudience {
	if len(shared) == 0 {
		return audience
	}
	hosts := make(map[string]struct{}, len(shared))
	raw := make(map[string]struct{})
	for _, inbox := range shared {
		if host, ok := inboxHost(inbox); ok {
			hosts[host] = struct{}{}
			continue
		}
		raw[inbox] = struct{}{}
	}
	audience.Followers = omitInboxes(audience.Followers, hosts, raw)
	audience.Holders = omitInboxes(audience.Holders, hosts, raw)
	return audience
}

func inboxHost(inbox string) (string, bool) {
	parsed, err := url.Parse(inbox)
	if err != nil || parsed.Host == "" {
		return "", false
	}
	return parsed.Host, true
}

func omitInboxes(inboxes []string, hosts, raw map[string]struct{}) []string {
	kept := make([]string, 0, len(inboxes))
	for _, inbox := range inboxes {
		if host, ok := inboxHost(inbox); ok {
			if _, found := hosts[host]; found {
				continue
			}
		} else if _, found := raw[inbox]; found {
			continue
		}
		kept = append(kept, inbox)
	}
	return kept
}
