package util

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/pocketbase/pocketbase/core"
)

// Map viewer preference scopes. Each (user, scope, ref) triple carries one
// stored record; ref is empty for the account scope and holds the trail or
// list id otherwise.
const (
	MapViewerPrefsScopeAccount = "account"
	MapViewerPrefsScopeTrail   = "trail"
	MapViewerPrefsScopeList    = "list"
)

// MapViewerPrefsPatch is a partial viewer-prefs update. Nil fields are absent
// and leave the stored value unchanged.
type MapViewerPrefsPatch struct {
	Base     *string
	Overlays *map[string]bool
	Pois     *map[string]map[string]bool
}

// MapViewerPrefs is the stored viewer-prefs record. Nil fields are unset.
type MapViewerPrefs struct {
	Base     *string
	Overlays *map[string]bool
	Pois     *map[string]map[string]bool
}

// EnsureMapViewerPrefsSchema creates the prefs table when it is missing so
// fresh and upgraded databases behave the same.
func EnsureMapViewerPrefsSchema(app core.App) error {
	_, err := app.DB().NewQuery(`CREATE TABLE IF NOT EXISTS map_viewer_prefs (
		user TEXT NOT NULL,
		scope TEXT NOT NULL,
		ref TEXT NOT NULL DEFAULT '',
		base TEXT,
		overlays TEXT,
		pois TEXT,
		PRIMARY KEY (user, scope, ref)
	)`).Execute()
	return err
}

func validMapViewerScope(scope string) bool {
	return scope == MapViewerPrefsScopeAccount ||
		scope == MapViewerPrefsScopeTrail ||
		scope == MapViewerPrefsScopeList
}

// ValidateMapViewerPrefsPatch rejects structurally invalid prefs. Base must
// be a non-empty style name when present; overlays and pois are validated by
// JSON shape (flat and nested boolean maps) at decode time.
func ValidateMapViewerPrefsPatch(patch MapViewerPrefsPatch) error {
	if patch.Base != nil && *patch.Base == "" {
		return fmt.Errorf("base must be a non-empty style name")
	}
	return nil
}

type mapViewerPrefsRow struct {
	Base     *string `db:"base"`
	Overlays *string `db:"overlays"`
	Pois     *string `db:"pois"`
}

// StoredMapViewerPrefs returns only the fields stored on one record.
func StoredMapViewerPrefs(app core.App, userID, scope, ref string) (MapViewerPrefs, error) {
	var prefs MapViewerPrefs
	if userID == "" || !validMapViewerScope(scope) {
		return prefs, fmt.Errorf("invalid prefs identity")
	}
	if err := EnsureMapViewerPrefsSchema(app); err != nil {
		return prefs, err
	}
	var row mapViewerPrefsRow
	err := app.DB().NewQuery(`SELECT base, overlays, pois FROM map_viewer_prefs
		WHERE user = {:user} AND scope = {:scope} AND ref = {:ref}`).
		Bind(map[string]any{"user": userID, "scope": scope, "ref": ref}).
		One(&row)
	if err == sql.ErrNoRows {
		return prefs, nil
	}
	if err != nil {
		return prefs, err
	}
	prefs.Base = row.Base
	if row.Overlays != nil {
		var overlays map[string]bool
		if err := json.Unmarshal([]byte(*row.Overlays), &overlays); err != nil {
			return prefs, err
		}
		prefs.Overlays = &overlays
	}
	if row.Pois != nil {
		var pois map[string]map[string]bool
		if err := json.Unmarshal([]byte(*row.Pois), &pois); err != nil {
			return prefs, err
		}
		prefs.Pois = &pois
	}
	return prefs, nil
}

func mergeViewerOverlays(levels ...*map[string]bool) *map[string]bool {
	merged := map[string]bool{}
	hasAny := false
	for _, level := range levels {
		if level == nil {
			continue
		}
		hasAny = true
		for key, value := range *level {
			merged[key] = value
		}
	}
	if !hasAny {
		return nil
	}
	return &merged
}

func mergeViewerPois(levels ...*map[string]map[string]bool) *map[string]map[string]bool {
	merged := map[string]map[string]bool{}
	hasAny := false
	for _, level := range levels {
		if level == nil {
			continue
		}
		hasAny = true
		for category, toggles := range *level {
			if merged[category] == nil {
				merged[category] = map[string]bool{}
			}
			for key, value := range toggles {
				merged[category][key] = value
			}
		}
	}
	if !hasAny {
		return nil
	}
	return &merged
}

// EffectiveMapViewerPrefs resolves one trail's prefs across all three
// levels: each setting comes from the trail record when stored there, then
// the list record when a list context is given and stored there, then the
// account record. Unset settings stay unset for the viewer defaults.
func EffectiveMapViewerPrefs(app core.App, userID, trailID, listID string) (MapViewerPrefs, error) {
	var prefs MapViewerPrefs
	if trailID == "" {
		return prefs, fmt.Errorf("invalid trail identity")
	}
	account, err := StoredMapViewerPrefs(app, userID, MapViewerPrefsScopeAccount, "")
	if err != nil {
		return prefs, err
	}
	trail, err := StoredMapViewerPrefs(app, userID, MapViewerPrefsScopeTrail, trailID)
	if err != nil {
		return prefs, err
	}
	var list MapViewerPrefs
	if listID != "" {
		list, err = StoredMapViewerPrefs(app, userID, MapViewerPrefsScopeList, listID)
		if err != nil {
			return prefs, err
		}
	}
	prefs.Base = account.Base
	if list.Base != nil {
		prefs.Base = list.Base
	}
	if trail.Base != nil {
		prefs.Base = trail.Base
	}
	prefs.Overlays = mergeViewerOverlays(account.Overlays, list.Overlays, trail.Overlays)
	prefs.Pois = mergeViewerPois(account.Pois, list.Pois, trail.Pois)
	return prefs, nil
}

// EffectiveListPrefs resolves one list's prefs: each setting comes from the
// list record when stored there and otherwise falls back to the account
// record. Unset settings stay unset for the viewer defaults.
func EffectiveListPrefs(app core.App, userID, listID string) (MapViewerPrefs, error) {
	var prefs MapViewerPrefs
	if listID == "" {
		return prefs, fmt.Errorf("invalid list identity")
	}
	account, err := StoredMapViewerPrefs(app, userID, MapViewerPrefsScopeAccount, "")
	if err != nil {
		return prefs, err
	}
	list, err := StoredMapViewerPrefs(app, userID, MapViewerPrefsScopeList, listID)
	if err != nil {
		return prefs, err
	}
	prefs.Base = account.Base
	if list.Base != nil {
		prefs.Base = list.Base
	}
	prefs.Overlays = mergeViewerOverlays(account.Overlays, list.Overlays)
	prefs.Pois = mergeViewerPois(account.Pois, list.Pois)
	return prefs, nil
}

// SeedListPrefsOnce seeds a list's stored prefs from the effective prefs
// (trail over account, no list level) of the first trail on the list record.
// It is a no-op when the list already stores any pref or the list has no
// trails. Later additions never reseed.
func SeedListPrefsOnce(app core.App, userID, listID string) error {
	if listID == "" {
		return fmt.Errorf("invalid list identity")
	}
	stored, err := StoredMapViewerPrefs(app, userID, MapViewerPrefsScopeList, listID)
	if err != nil {
		return err
	}
	if stored.Base != nil || stored.Overlays != nil || stored.Pois != nil {
		return nil
	}
	record, err := app.FindRecordById("lists", listID)
	if err != nil {
		return nil
	}
	trails := record.GetStringSlice("trails")
	if len(trails) == 0 {
		return nil
	}
	effective, err := EffectiveMapViewerPrefs(app, userID, trails[0], "")
	if err != nil {
		return err
	}
	_, err = PutMapViewerPrefs(app, userID, MapViewerPrefsScopeList, listID, MapViewerPrefsPatch{
		Base:     effective.Base,
		Overlays: effective.Overlays,
		Pois:     effective.Pois,
	})
	return err
}

// PutMapViewerPrefs partially updates one stored record, creating it when
// missing, and returns the stored record.
func PutMapViewerPrefs(app core.App, userID, scope, ref string, patch MapViewerPrefsPatch) (MapViewerPrefs, error) {
	var prefs MapViewerPrefs
	if userID == "" || !validMapViewerScope(scope) {
		return prefs, fmt.Errorf("invalid prefs identity")
	}
	if err := ValidateMapViewerPrefsPatch(patch); err != nil {
		return prefs, err
	}
	if err := EnsureMapViewerPrefsSchema(app); err != nil {
		return prefs, err
	}
	_, err := app.DB().NewQuery(`INSERT INTO map_viewer_prefs(user, scope, ref)
		VALUES ({:user}, {:scope}, {:ref})
		ON CONFLICT(user, scope, ref) DO NOTHING`).
		Bind(map[string]any{"user": userID, "scope": scope, "ref": ref}).
		Execute()
	if err != nil {
		return prefs, err
	}
	if patch.Base != nil {
		if _, err := app.DB().NewQuery(`UPDATE map_viewer_prefs SET base = {:base}
			WHERE user = {:user} AND scope = {:scope} AND ref = {:ref}`).
			Bind(map[string]any{"base": *patch.Base, "user": userID, "scope": scope, "ref": ref}).
			Execute(); err != nil {
			return prefs, err
		}
	}
	if patch.Overlays != nil {
		raw, err := json.Marshal(*patch.Overlays)
		if err != nil {
			return prefs, err
		}
		if _, err := app.DB().NewQuery(`UPDATE map_viewer_prefs SET overlays = {:overlays}
			WHERE user = {:user} AND scope = {:scope} AND ref = {:ref}`).
			Bind(map[string]any{"overlays": string(raw), "user": userID, "scope": scope, "ref": ref}).
			Execute(); err != nil {
			return prefs, err
		}
	}
	if patch.Pois != nil {
		raw, err := json.Marshal(*patch.Pois)
		if err != nil {
			return prefs, err
		}
		if _, err := app.DB().NewQuery(`UPDATE map_viewer_prefs SET pois = {:pois}
			WHERE user = {:user} AND scope = {:scope} AND ref = {:ref}`).
			Bind(map[string]any{"pois": string(raw), "user": userID, "scope": scope, "ref": ref}).
			Execute(); err != nil {
			return prefs, err
		}
	}
	return StoredMapViewerPrefs(app, userID, scope, ref)
}
