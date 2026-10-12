package routes

import (
	"net/http"
	"pocketbase/util"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type mapViewerPrefsRequest struct {
	Base     *string                      `json:"base"`
	Overlays *map[string]bool             `json:"overlays"`
	Pois     *map[string]map[string]bool `json:"pois"`
}

func mapViewerPrefsResponse(prefs util.MapViewerPrefs) map[string]any {
	body := map[string]any{}
	if prefs.Base != nil {
		body["base"] = *prefs.Base
	}
	if prefs.Overlays != nil {
		body["overlays"] = *prefs.Overlays
	}
	if prefs.Pois != nil {
		body["pois"] = *prefs.Pois
	}
	return body
}

// MapViewerPrefsAccountGet serves the stored account prefs. Unset fields are
// omitted; a fresh account reads as an empty object.
func MapViewerPrefsAccountGet(e *core.RequestEvent) error {
	if e.Auth == nil {
		return apis.NewUnauthorizedError("authentication required", nil)
	}
	prefs, err := util.StoredMapViewerPrefs(e.App, e.Auth.Id, util.MapViewerPrefsScopeAccount, "")
	if err != nil {
		return apis.NewBadRequestError("failed to read viewer prefs", err)
	}
	return e.JSON(http.StatusOK, mapViewerPrefsResponse(prefs))
}

// MapViewerPrefsAccountPut partially updates the stored account prefs. Absent
// JSON fields leave the stored value unchanged.
func MapViewerPrefsAccountPut(e *core.RequestEvent) error {
	if e.Auth == nil {
		return apis.NewUnauthorizedError("authentication required", nil)
	}
	var request mapViewerPrefsRequest
	if err := e.BindBody(&request); err != nil {
		return apis.NewBadRequestError("failed to read request data", err)
	}
	prefs, err := util.PutMapViewerPrefs(e.App, e.Auth.Id, util.MapViewerPrefsScopeAccount, "", util.MapViewerPrefsPatch{
		Base:     request.Base,
		Overlays: request.Overlays,
		Pois:     request.Pois,
	})
	if err != nil {
		return apis.NewBadRequestError("failed to store viewer prefs", err)
	}
	return e.JSON(http.StatusOK, mapViewerPrefsResponse(prefs))
}

// MapViewerPrefsTrailGet serves one trail's effective prefs across all three
// levels. With ?list={listId} each setting resolves trail, then list, then
// account; without it the list level is skipped.
func MapViewerPrefsTrailGet(e *core.RequestEvent) error {
	if e.Auth == nil {
		return apis.NewUnauthorizedError("authentication required", nil)
	}
	trailID := e.Request.PathValue("id")
	if trailID == "" {
		return apis.NewBadRequestError("missing trail id", nil)
	}
	prefs, err := util.EffectiveMapViewerPrefs(e.App, e.Auth.Id, trailID, e.Request.URL.Query().Get("list"))
	if err != nil {
		return apis.NewBadRequestError("failed to read viewer prefs", err)
	}
	return e.JSON(http.StatusOK, mapViewerPrefsResponse(prefs))
}

// MapViewerPrefsTrailPut partially updates one trail's stored prefs. Absent
// JSON fields leave the stored value unchanged; other trails are untouched.
func MapViewerPrefsTrailPut(e *core.RequestEvent) error {
	if e.Auth == nil {
		return apis.NewUnauthorizedError("authentication required", nil)
	}
	trailID := e.Request.PathValue("id")
	if trailID == "" {
		return apis.NewBadRequestError("missing trail id", nil)
	}
	var request mapViewerPrefsRequest
	if err := e.BindBody(&request); err != nil {
		return apis.NewBadRequestError("failed to read request data", err)
	}
	prefs, err := util.PutMapViewerPrefs(e.App, e.Auth.Id, util.MapViewerPrefsScopeTrail, trailID, util.MapViewerPrefsPatch{
		Base:     request.Base,
		Overlays: request.Overlays,
		Pois:     request.Pois,
	})
	if err != nil {
		return apis.NewBadRequestError("failed to store viewer prefs", err)
	}
	return e.JSON(http.StatusOK, mapViewerPrefsResponse(prefs))
}

// MapViewerPrefsListGet serves one list's effective prefs: each setting comes
// from the list record when stored there and otherwise falls back to the
// account record. Reading a list with no stored prefs and non-empty trails
// seeds it once from its first trail.
func MapViewerPrefsListGet(e *core.RequestEvent) error {
	if e.Auth == nil {
		return apis.NewUnauthorizedError("authentication required", nil)
	}
	listID := e.Request.PathValue("id")
	if listID == "" {
		return apis.NewBadRequestError("missing list id", nil)
	}
	if err := util.SeedListPrefsOnce(e.App, e.Auth.Id, listID); err != nil {
		return apis.NewBadRequestError("failed to read viewer prefs", err)
	}
	prefs, err := util.EffectiveListPrefs(e.App, e.Auth.Id, listID)
	if err != nil {
		return apis.NewBadRequestError("failed to read viewer prefs", err)
	}
	return e.JSON(http.StatusOK, mapViewerPrefsResponse(prefs))
}

// MapViewerPrefsListPut partially updates one list's stored prefs. Absent
// JSON fields leave the stored value unchanged; other lists are untouched.
func MapViewerPrefsListPut(e *core.RequestEvent) error {
	if e.Auth == nil {
		return apis.NewUnauthorizedError("authentication required", nil)
	}
	listID := e.Request.PathValue("id")
	if listID == "" {
		return apis.NewBadRequestError("missing list id", nil)
	}
	var request mapViewerPrefsRequest
	if err := e.BindBody(&request); err != nil {
		return apis.NewBadRequestError("failed to read request data", err)
	}
	prefs, err := util.PutMapViewerPrefs(e.App, e.Auth.Id, util.MapViewerPrefsScopeList, listID, util.MapViewerPrefsPatch{
		Base:     request.Base,
		Overlays: request.Overlays,
		Pois:     request.Pois,
	})
	if err != nil {
		return apis.NewBadRequestError("failed to store viewer prefs", err)
	}
	return e.JSON(http.StatusOK, mapViewerPrefsResponse(prefs))
}
