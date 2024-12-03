import * as xml2js from 'isomorphic-xml2js';
import Metadata from './metadata';
import Route from './route';
import Track from './track';
import { allDatesToISOString, removeEmpty } from './utils';
import Waypoint from './waypoint';
import GpxMetricsComputation from './gpx-metrics-computation';

const defaultAttributes = {
  version: '1.1',
  creator: 'wanderer',
  xmlns: 'http://www.topografix.com/GPX/1/1',
  'xmlns:xsi': 'http://www.w3.org/2001/XMLSchema-instance',
  'xsi:schemaLocation':
    'http://www.topografix.com/GPX/1/1 http://www.topografix.com/GPX/1/1/gpx.xsd'
}

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
      this.wpt = object.wpt.map(wpt => new Waypoint(wpt))
    }
    if (object.rte) {
      if (!Array.isArray(object.rte)) {
        object.rte = [object.rte];
      }
      this.rte = object.rte.map(rte => new Route(rte))
    }
    if (object.trk) {
      if (!Array.isArray(object.trk)) {
        object.trk = [object.trk];
      }
      this.trk = object.trk.map(trk => new Track(trk))
    }

    removeEmpty(this);
  }

  getTotals() {
    let totalElevationGain = 0;
    let totalElevationLoss = 0;
    let totalDuration = 0;
    let totalDistance = 0;

    const filter = new GpxMetricsComputation(5, 10);

    for (const track of this.trk ?? []) {
      for (const segment of track.trkseg ?? []) {
        const points = segment.trkpt ?? [];

        if (points.length >= 2) {
          const startTime = points[0].time;
          const endTime = points[points.length - 1].time

          if (startTime && endTime) {
            totalDuration += endTime.getTime() - startTime.getTime();
          }
        }

        for (const point of points) {
          filter.addAndFilter(point);
        }
      }
    }

    totalElevationGain = filter.totalElevationGainSmoothed;
    totalElevationLoss = filter.totalElevationLossSmoothed;
    totalDistance = filter.totalDistance;

    return { distance: totalDistance, elevationGain: totalElevationGain, elevationLoss: totalElevationLoss, duration: totalDuration }
  }

  static parse(gpxString: string): Promise<GPX | Error> {
    return new Promise<GPX | Error>((resolve, reject) => xml2js.parseString(gpxString, {
      explicitArray: false,
      attrValueProcessors: [(str: string | number) => {
        if (!isNaN(Number(str))) {
          str = Number.isInteger(Number(str)) ? parseInt(String(str), 10) : parseFloat(String(str));
        }
        return str;
      }
      ]
    }, (err, xml) => {
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
  
    return xmlString;
  }
}
