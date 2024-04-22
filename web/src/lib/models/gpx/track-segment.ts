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
}