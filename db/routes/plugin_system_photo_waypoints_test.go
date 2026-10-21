package routes

import (
	"testing"

	"pocketbase/pluginsystem"
)

func positionedPhoto(id string, lat float64) pluginsystem.Photo {
	lon := 11.0
	return pluginsystem.Photo{ExternalID: id, Lat: &lat, Lon: &lon}
}

func photoIDs(photos []pluginsystem.Photo) []string {
	ids := make([]string, len(photos))
	for i, photo := range photos {
		ids[i] = photo.ExternalID
	}
	return ids
}

func TestMovePhotosToWaypoints(t *testing.T) {
	item := pluginsystem.TrailImport{
		Waypoints: []pluginsystem.Waypoint{{Name: "Hut", Lat: 47.0, Lon: 11}},
		Photos: []pluginsystem.Photo{
			positionedPhoto("cover", 47.0),
			{ExternalID: "no-position"},
			positionedPhoto("bad-position", 95),
			positionedPhoto("at-hut", 47.0002),
			positionedPhoto("lake-1", 47.1),
			positionedPhoto("lake-2", 47.1002),
		},
	}

	movePhotosToWaypoints(nil, &item, nil)

	if got := photoIDs(item.Photos); len(got) != 3 || got[0] != "cover" || got[1] != "no-position" || got[2] != "bad-position" {
		t.Fatalf("trail photos = %v, want [cover no-position bad-position]", got)
	}
	if len(item.Waypoints) != 2 {
		t.Fatalf("got %d waypoints, want 2", len(item.Waypoints))
	}
	if got := photoIDs(item.Waypoints[0].Photos); len(got) != 1 || got[0] != "at-hut" {
		t.Fatalf("hut photos = %v, want [at-hut]", got)
	}
	lake := item.Waypoints[1]
	if got := photoIDs(lake.Photos); lake.Name != "" || len(got) != 2 || got[0] != "lake-1" || got[1] != "lake-2" {
		t.Fatalf("new waypoint = %+v, want an unnamed waypoint with both lake photos", lake)
	}
}

func TestGroupPhotosIntoWaypointsCapsNewWaypoints(t *testing.T) {
	item := pluginsystem.TrailImport{Photos: []pluginsystem.Photo{{ExternalID: "cover"}}}
	for i := range 30 {
		item.Photos = append(item.Photos, positionedPhoto("p", 47.0+float64(i)*0.01))
	}

	groupPhotosIntoWaypoints(&item, waypointMergeSettings{Enabled: false})

	if len(item.Waypoints) != maxPhotoWaypointsPerTrail {
		t.Fatalf("got %d waypoints, want %d", len(item.Waypoints), maxPhotoWaypointsPerTrail)
	}
	if len(item.Photos) != 6 {
		t.Fatalf("got %d trail photos, want the cover and the 5 over the cap", len(item.Photos))
	}
}

func TestGroupPhotosIntoWaypointsKeepsPhotosOverTheWaypointLimit(t *testing.T) {
	item := pluginsystem.TrailImport{Photos: []pluginsystem.Photo{{ExternalID: "cover"}}}
	for range 25 {
		item.Photos = append(item.Photos, positionedPhoto("p", 47.0))
	}

	groupPhotosIntoWaypoints(&item, waypointMergeSettings{Enabled: true, Radius: defaultWaypointMergeRadius})

	if len(item.Waypoints) != 1 {
		t.Fatalf("got %d waypoints, want 1", len(item.Waypoints))
	}
	if len(item.Waypoints[0].Photos) != 20 {
		t.Fatalf("got %d waypoint photos, want 20", len(item.Waypoints[0].Photos))
	}
	if len(item.Photos) != 6 {
		t.Fatalf("got %d trail photos, want the cover and the 5 over the limit", len(item.Photos))
	}
}
