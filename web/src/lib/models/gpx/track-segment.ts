import type Track from './track';
import Waypoint from './waypoint';

export default class TrackSegment {
  trkpt?: Waypoint[];
  extensions?: string;
  constructor(object: { trkpt?: Waypoint[], extensions?: string }) {
    if (object.trkpt) {
      if (!Array.isArray(object.trkpt)) {
        object.trkpt = [object.trkpt];
      }
      this.trkpt = object.trkpt.map(trkpt => new Waypoint(trkpt))
    }
    this.extensions = object.extensions;
  }

  toGeoJSON(
    track: Track,
    segmentId: number,
    featureId: number
  ): GeoJSON.Feature {
    const coordinates = (this.trkpt || []).map(pt => [
      pt.$.lon ?? 0,
      pt.$.lat ?? 0,
      pt.ele ?? 0,
    ]);

    const times = (this.trkpt || []).map(pt => pt.time?.toISOString() ?? null);

    return {
      type: "Feature",
      geometry: {
        type: "LineString",
        coordinates,
      },
      properties: {
        name: track.name,
        desc: track.desc,
        type: track.type,
        number: track.number,
        featureId,
        segmentId,
        coordinateProperties: {
          times
        }
      }
    };
  }

}