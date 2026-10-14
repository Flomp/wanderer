import 'package:flutter_test/flutter_test.dart';
import 'package:gpx/gpx.dart';
import 'package:wanderer/util/gpx/conversion.dart';
import 'package:wanderer/util/gpx/gpx.dart';
import 'package:wanderer/util/route/planner_handoff.dart';

/// A 3-leg loop as the web's `trail2gpx` exports it since #1370: one
/// segment, the two inner leg boundaries marked with `wanderer:anchor`.
const _joinedLoop =
    '<?xml version="1.0" encoding="UTF-8"?>'
    '<gpx version="1.1" creator="wanderer" xmlns="http://www.topografix.com/GPX/1/1" '
    'xmlns:wanderer="https://wanderer.to/xmlns/gpx/1"><trk><trkseg>'
    '<trkpt lat="47.0" lon="11.0"/>'
    '<trkpt lat="47.1" lon="11.1"/>'
    '<trkpt lat="47.2" lon="11.2"><extensions><wanderer:anchor/></extensions></trkpt>'
    '<trkpt lat="47.3" lon="11.3"><extensions><wanderer:anchor></wanderer:anchor></extensions></trkpt>'
    '<trkpt lat="47.4" lon="11.4"/>'
    '<trkpt lat="47.0" lon="11.0"/>'
    '</trkseg></trk></gpx>';

List<List<List<double>>> _coordinates(Gpx gpx) => [
  for (final seg in gpx.trks.single.trksegs)
    [
      for (final p in seg.trkpts) [p.lat!, p.lon!],
    ],
];

void main() {
  group('parseGpxSafely with route anchor marks', () {
    test('restores the legs of a joined web export', () {
      final gpx = parseGpxSafely(_joinedLoop);

      expect(_coordinates(gpx), [
        [
          [47.0, 11.0],
          [47.1, 11.1],
        ],
        [
          [47.2, 11.2],
        ],
        [
          [47.3, 11.3],
          [47.4, 11.4],
          [47.0, 11.0],
        ],
      ]);
    });

    test('recovers every drawn anchor for the route planner', () {
      final gpx = parseGpxSafely(_joinedLoop);
      final anchors = anchorsFromTrack(gpx);

      expect(
        [
          for (final a in anchors) [a.lat, a.lon],
        ],
        [
          [47.0, 11.0],
          [47.2, 11.2],
          [47.3, 11.3],
          [47.0, 11.0],
        ],
      );
      expect(segmentPolylinesFromTrack(gpx, anchors).map((l) => l.length), [
        3,
        2,
        3,
      ]);
    });

    test('keeps the flattened point stream identical to the file', () {
      final gpx = parseGpxSafely(_joinedLoop);

      expect(
        [
          for (final p in gpx.allPoints) [p.lat, p.lon],
        ],
        [
          [47.0, 11.0],
          [47.1, 11.1],
          [47.2, 11.2],
          [47.3, 11.3],
          [47.4, 11.4],
          [47.0, 11.0],
        ],
      );
    });

    test('strips the marks so GpxWriter output stays well-formed', () {
      final xml = GpxWriter().asString(parseGpxSafely(_joinedLoop));

      expect(xml, isNot(contains('wanderer:')));
      expect(() => parseGpxSafely(xml), returnsNormally);
    });

    test('keeps other extensions on a marked point', () {
      final gpx = parseGpxSafely(
        '<gpx><trk><trkseg>'
        '<trkpt lat="1" lon="1"/>'
        '<trkpt lat="2" lon="2"><extensions><ns3:hr>90</ns3:hr><wanderer:anchor/></extensions></trkpt>'
        '<trkpt lat="3" lon="3"/>'
        '</trkseg></trk></gpx>',
      );

      expect(gpx.trks.single.trksegs, hasLength(2));
      expect(gpx.trks.single.trksegs[1].trkpts.first.extensions, {
        'ns3:hr': '90',
      });
    });

    test('ignores marks on the ends of a segment but still strips them', () {
      final gpx = parseGpxSafely(
        '<gpx><trk><trkseg>'
        '<trkpt lat="1" lon="1"><extensions><wanderer:anchor/></extensions></trkpt>'
        '<trkpt lat="2" lon="2"/>'
        '<trkpt lat="3" lon="3"><extensions><wanderer:anchor/></extensions></trkpt>'
        '</trkseg></trk></gpx>',
      );

      expect(_coordinates(gpx), [
        [
          [1.0, 1.0],
          [2.0, 2.0],
          [3.0, 3.0],
        ],
      ]);
      expect(
        gpx.trks.single.trksegs.single.trkpts.every(
          (p) => p.extensions.isEmpty,
        ),
        isTrue,
      );
    });

    test('leaves an unmarked multi-segment track unchanged', () {
      final gpx = parseGpxSafely(
        '<gpx><trk>'
        '<trkseg><trkpt lat="1" lon="1"/><trkpt lat="2" lon="2"/></trkseg>'
        '<trkseg><trkpt lat="2" lon="2"/><trkpt lat="3" lon="3"/></trkseg>'
        '</trk></gpx>',
      );

      expect(_coordinates(gpx), [
        [
          [1.0, 1.0],
          [2.0, 2.0],
        ],
        [
          [2.0, 2.0],
          [3.0, 3.0],
        ],
      ]);
    });
  });
}
