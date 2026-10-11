package routes

import (
	"context"
	"log/slog"
	"testing"
)

func TestResolveWaypointClusterNamesNamesOnlyNewClusters(t *testing.T) {
	clusters := []waypointPhotoCluster{
		{Photos: []string{"photo"}, Count: 1, Lat: 46.1, Lon: 8.2},
		{Waypoint: "existing", Name: "Existing name", Photos: []string{"photo"}, Count: 2, Lat: 46.2, Lon: 8.3},
		{Name: "Empty cluster", Lat: 46.3, Lon: 8.4},
	}
	calls := 0
	resolver := func(_ context.Context, _ *slog.Logger, lat float64, lon float64, radius float64) (string, error) {
		calls++
		if lat != 46.1 || lon != 8.2 || radius != 75 {
			t.Fatalf("resolver received lat=%v lon=%v radius=%v", lat, lon, radius)
		}
		return "Resolved name", nil
	}

	resolveWaypointClusterNames(context.Background(), nil, clusters, 75, resolver)

	if calls != 1 {
		t.Fatalf("resolver called %d times, want 1", calls)
	}
	if clusters[0].Name != "Resolved name" {
		t.Fatalf("new cluster name = %q, want Resolved name", clusters[0].Name)
	}
	if clusters[1].Name != "Existing name" {
		t.Fatalf("existing waypoint cluster name = %q, want Existing name", clusters[1].Name)
	}
	if clusters[2].Name != "Empty cluster" {
		t.Fatalf("empty cluster name = %q, want Empty cluster", clusters[2].Name)
	}
}

func clusterTestPhoto(id string, lat float64, minutes *int64) waypointClusterPhoto {
	photo := waypointClusterPhoto{ID: id, Lat: lat, Lon: 11}
	if minutes != nil {
		millis := *minutes * 60 * 1000
		photo.Time = &millis
	}
	return photo
}

func minutes(m int64) *int64 {
	return &m
}

func TestClusterWaypointPhotosSplitsPassesByTime(t *testing.T) {
	settings := waypointMergeSettings{Enabled: true, Radius: defaultWaypointMergeRadius}

	// Same spot about 20 m apart, once on the way out and once on the way back.
	photos := []waypointClusterPhoto{
		clusterTestPhoto("out", 47.0, minutes(8)),
		clusterTestPhoto("back", 47.0002, minutes(300)),
	}
	if clusters := clusterWaypointPhotos(photos, nil, settings); len(clusters) != 2 {
		t.Fatalf("got %d clusters, want 2", len(clusters))
	}
}

func TestClusterWaypointPhotosMergesNearbyPhotosCloseInTime(t *testing.T) {
	settings := waypointMergeSettings{Enabled: true, Radius: defaultWaypointMergeRadius}

	// A chain of photos each within the gap of the previous one stays together.
	photos := []waypointClusterPhoto{
		clusterTestPhoto("a", 47.0, minutes(0)),
		clusterTestPhoto("b", 47.0001, minutes(10)),
		clusterTestPhoto("c", 47.0, minutes(20)),
	}
	clusters := clusterWaypointPhotos(photos, nil, settings)
	if len(clusters) != 1 || len(clusters[0].Photos) != 3 {
		t.Fatalf("got %+v, want one cluster with 3 photos", clusters)
	}
}

func TestClusterWaypointPhotosMergesWithoutTimes(t *testing.T) {
	settings := waypointMergeSettings{Enabled: true, Radius: defaultWaypointMergeRadius}

	photos := []waypointClusterPhoto{
		clusterTestPhoto("timed", 47.0, minutes(8)),
		clusterTestPhoto("untimed", 47.0002, nil),
	}
	waypoints := []waypointClusterWaypoint{{ID: "wp", Lat: 47.0001, Lon: 11}}
	clusters := clusterWaypointPhotos(photos, waypoints, settings)
	if len(clusters) != 1 || len(clusters[0].Photos) != 2 {
		t.Fatalf("got %+v, want one cluster with 2 photos", clusters)
	}
}

func TestClusterWaypointPhotosMergesChainInAnyOrder(t *testing.T) {
	settings := waypointMergeSettings{Enabled: true, Radius: defaultWaypointMergeRadius}

	photos := []waypointClusterPhoto{
		clusterTestPhoto("a", 47.0, minutes(0)),
		clusterTestPhoto("c", 47.0, minutes(20)),
		clusterTestPhoto("b", 47.0001, minutes(10)),
	}
	clusters := clusterWaypointPhotos(photos, nil, settings)
	if len(clusters) != 1 || len(clusters[0].Photos) != 3 {
		t.Fatalf("got %+v, want one cluster with 3 photos", clusters)
	}
}

func TestClusterWaypointPhotosKeepsWaypointTimeApart(t *testing.T) {
	settings := waypointMergeSettings{Enabled: true, Radius: defaultWaypointMergeRadius}

	// A waypoint imported from a photo on the way out, then a photo of the same
	// spot on the way back.
	outTime := int64(8 * 60 * 1000)
	waypoints := []waypointClusterWaypoint{{ID: "out", Lat: 47.0, Lon: 11, Time: &outTime}}
	photos := []waypointClusterPhoto{clusterTestPhoto("back", 47.0002, minutes(300))}
	clusters := clusterWaypointPhotos(photos, waypoints, settings)
	if len(clusters) != 2 || len(clusters[0].Photos) != 0 {
		t.Fatalf("got %+v, want the photo in a cluster of its own", clusters)
	}
}
