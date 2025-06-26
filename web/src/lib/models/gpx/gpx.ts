import * as xml2js from 'isomorphic-xml2js';
import Metadata from './metadata';
import Route from './route';
import Track from './track';
import { allDatesToISOString, haversineDistance, removeEmpty } from './utils';
import Waypoint from './waypoint';
import GpxMetricsComputation from './gpx-metrics-computation';
//@ts-ignore
import geohash from "ngeohash"
import { encodePolyline } from '$lib/util/polyline_util';
import { APIError } from '$lib/util/api_util';
import type { ValhallaHeightResponse } from '../valhalla';

const defaultAttributes = {
  version: '1.1',
  creator: 'wanderer',
  xmlns: 'http://www.topografix.com/GPX/1/1',
  'xmlns:xsi': 'http://www.w3.org/2001/XMLSchema-instance',
  'xsi:schemaLocation':
    'http://www.topografix.com/GPX/1/1 http://www.topografix.com/GPX/1/1/gpx.xsd'
}


type GPXFeature = {
  centroid: { lat: number; lon: number };
  boundingBox: { minLat: number; maxLat: number; minLon: number; maxLon: number };
  distance: number;
  elevationGain?: number;
  elevationLoss?: number;
  duration: number;
  hash: string; // MinHash or Geohash for track shape
};

export default class GPX {
  $: {
    version: string;
    creator: string;
    xmlns: string;
    'xmlns:xsi': string;
    'xsi:schemaLocation': string;
  }
  extensions?: string;
  metadata?: Metadata;
  wpt?: Waypoint[];
  rte?: Route[];
  trk?: Track[];
  features: GPXFeature

  constructor(object: {
    $?: {
      version: string,
      creator: string,
      xmlns: string,
      'xmlns:xsi': string,
      'xsi:schemaLocation': string,
    }
    extensions?: string,
    metadata?: Metadata,
    wpt?: Waypoint[] | Waypoint,
    rte?: Route[] | Route,
    trk?: Track[],
  }) {
    this.$ = Object.assign({}, defaultAttributes, object.$ || {});
    if (object.extensions) {
      this.extensions = object.extensions;
    }

    if (object.metadata) {
      this.metadata = object.metadata;
    }
    if (object.wpt) {
      if (!Array.isArray(object.wpt)) {
        object.wpt = [object.wpt];
      }
      this.wpt = object.wpt.filter(trk => typeof trk === 'object').map(wpt => new Waypoint(wpt))
    }
    if (object.rte) {
      if (!Array.isArray(object.rte)) {
        object.rte = [object.rte];
      }
      this.rte = object.rte.filter(trk => typeof trk === 'object').map(rte => new Route(rte))
    }
    if (object.trk) {
      if (!Array.isArray(object.trk)) {
        object.trk = [object.trk];
      }
      this.trk = object.trk.filter(trk => typeof trk === 'object').map(trk => new Track(trk))
    }

    this.features = this.getTotals();

    removeEmpty(this);
  }

  getTotals(): GPXFeature {
    let totalElevationGain = 0;
    let totalElevationLoss = 0;
    let totalDuration = 0;
    let totalDistance = 0;
    let totalLat = 0
    let totalLon = 0

    const metrics = new GpxMetricsComputation(5, 5);

    let minLat = Infinity, maxLat = -Infinity, minLon = Infinity, maxLon = -Infinity;

    const allPoints: Waypoint[] = []
    for (const track of this.trk ?? []) {
      for (const segment of track.trkseg ?? []) {
        const points = segment.trkpt ?? [];
        allPoints.push(...points);

        if (points.length >= 2) {
          const startTime = points[0].time;
          const endTime = points[points.length - 1].time

          if (startTime && endTime) {
            totalDuration += endTime.getTime() - startTime.getTime();
          }
        }

        const pointLength = points.length
        for (let i = 1; i < pointLength; i++) {
          const point = points[i];
          metrics.addAndFilter(point)

          totalLat += point.$.lat ?? 0;
          totalLon += point.$.lon ?? 0;

          minLat = Math.min(minLat, point.$.lat ?? Infinity);
          maxLat = Math.max(maxLat, point.$.lat ?? -Infinity);
          minLon = Math.min(minLon, point.$.lon ?? Infinity);
          maxLon = Math.max(maxLon, point.$.lon ?? -Infinity);
        }
      }
    }

    totalElevationGain = metrics.totalElevationGainSmoothed;
    totalElevationLoss = metrics.totalElevationLossSmoothed;
    totalDistance = metrics.totalDistance;

    const boundingBox = { minLat, maxLat, minLon, maxLon };
    const centroid = { lat: totalLat / allPoints.length, lon: totalLon / allPoints.length };

    return {
      centroid,
      boundingBox,
      distance: totalDistance,
      elevationGain: totalElevationGain,
      elevationLoss: totalElevationLoss,
      duration: totalDuration,
      hash: this.generateMinHash(allPoints)
    }
  }

  private generateMinHash(points: Waypoint[]): string {
    const hashes = points.map(pt => geohash.encode(pt.$.lat, pt.$.lon));
    return hashes.sort().join('').slice(0, 10);
  }

  async correctElevation(f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {
    let coordinates: number[][] = []
    for (const track of this.trk ?? []) {
      for (const segment of track.trkseg ?? []) {
        if (!segment.trkpt) {
          continue
        }
        coordinates = coordinates.concat(segment.trkpt.map(pt => [pt.$.lat ?? 0, pt.$.lon ?? 0]))
      }
    }

    const shape = encodePolyline(coordinates);
    const r2 = await f("/api/v1/valhalla/height", { method: "POST", body: JSON.stringify({ encoded_polyline: shape }) })

    if (!r2.ok) {
      const response = await r2.json();
      throw new APIError(r2.status, response.message, response.detail)
    }

    const heightResponse: ValhallaHeightResponse = await r2.json()

    let heightIndex = 0;
    for (const track of this.trk ?? []) {
      for (const segment of track.trkseg ?? []) {
        if (!segment.trkpt) {
          continue
        }
        segment.trkpt.forEach((pt) => {
          pt.ele = heightResponse.height[heightIndex]
          heightIndex++;
        })
      }
    }
    this.features = this.getTotals()
  }

  static parse(gpxString: string): Promise<GPX | Error> {
    const sanitizedGPX = gpxString.replace(/\sxmlns=""/g, '').replace(/<!--[\s\S]*?-->/g, '');

    return new Promise<GPX | Error>((resolve, reject) => xml2js.parseString(sanitizedGPX, {
      explicitArray: false,
      attrValueProcessors: [(str: string) => {
        if (str.length && !isNaN(Number(str))) {
          return Number.isInteger(Number(str)) ? parseInt(String(str), 10) : parseFloat(String(str));
        }
        return str;
      }
      ]
    }, async (err, xml) => {
      if (err) {
        reject(err);
      }
      const gpx = new GPX({
        $: xml.gpx.$,
        metadata: xml.gpx.metadata,
        wpt: xml.gpx.wpt,
        rte: xml.gpx.rte,
        trk: xml.gpx.trk
      });
      resolve(gpx)
    }));
  }

  toString(options?: xml2js.BuilderOptions) {
    options = options || {};
    options.rootName = 'gpx';

    const builder = new xml2js.Builder(options)
    const gpx = new GPX(this);
    allDatesToISOString(gpx);

    let xmlString = builder.buildObject(gpx);

    // Ensure xmlns is present in the root element for Firefox
    if (!xmlString.includes(`xmlns="${defaultAttributes["xmlns"]}"`)) {
      xmlString = xmlString.replace('<gpx', `<gpx xmlns="${defaultAttributes["xmlns"]}"`);
    }

    // Safari for some ungodly reason adds empty xmlns attributes to every tag...
    xmlString = xmlString.replace(/ xmlns=""/g, '');

    return xmlString;
  }
}
