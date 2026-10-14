import 'package:gpx/gpx.dart';

/// The extension the web marks a joined leg boundary with on export
/// (`web/src/lib/util/gpx_segment_util.ts`, namespace
/// `https://wanderer.to/xmlns/gpx/1`). `<wanderer:anchor/>` reaches the
/// vendored reader as the extension key `wanderer:anchor`.
const routeAnchorExtension = 'wanderer:anchor';

/// Splits every `trkseg` of [gpx] at points marked with
/// [routeAnchorExtension] and removes the marks, in place.
///
/// The web exports a drawn route as ONE continuous segment (Garmin shows a
/// multi-segment track one segment at a time) and marks each former leg
/// boundary instead. The route planner recovers anchors from `trkseg`
/// boundaries (`anchorsFromTrack`), so without this split a re-imported
/// export collapses to a bare start/end pair.
///
/// The legs use the app's own layout (`buildFinalPlannedGpx`): the anchor
/// point opens the next leg and is not repeated as the previous leg's last
/// point, so the flattened point stream — and every metric computed from it
/// — is identical to the joined file's.
///
/// The marks are always stripped, including on a segment's first or last
/// point (which never splits). Left in place, `GpxWriter` would write them
/// back without the `xmlns:wanderer` declaration, producing XML with an
/// unbound prefix.
void splitTrackSegmentsAtAnchors(Gpx gpx) {
  for (final track in gpx.trks) {
    final split = <Trkseg>[];

    for (final segment in track.trksegs) {
      final points = segment.trkpts;
      var start = 0;

      for (var i = 0; i < points.length; i++) {
        if (points[i].extensions.remove(routeAnchorExtension) == null ||
            i == 0 ||
            i == points.length - 1) {
          continue;
        }
        split.add(
          Trkseg(
            trkpts: points.sublist(start, i),
            extensions: Map.of(segment.extensions),
          ),
        );
        start = i;
      }

      if (start == 0) {
        split.add(segment);
      } else {
        split.add(
          Trkseg(
            trkpts: points.sublist(start),
            extensions: Map.of(segment.extensions),
          ),
        );
      }
    }

    track.trksegs = split;
  }
}
