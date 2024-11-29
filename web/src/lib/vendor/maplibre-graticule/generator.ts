import distance from '@turf/distance';
import destination from '@turf/destination';

type graticuleJson = {
  meridians: GeoJSON.Feature<GeoJSON.LineString>[];
  parallels: GeoJSON.Feature<GeoJSON.LineString>[];
};

import { type Units } from '@turf/helpers';

/**
 * 
 * @param {number} n 
 * @param {number} decimals 
 * @returns {number}
 */
export function toFixed(n: number, decimals: number): number {
  const factor = Math.pow(10, decimals);
  return Math.round(n * factor) / factor;
}

/**
 * 
 * @param {number} a 
 * @param {number} b 
 * @returns {number}
 */
export function modulo(a: number, b: number): number {
  const r = a % b;
  return r * b < 0 ? r + b : r;
}

/**
* 
* @param {number} number 
* @param {number} width 
* @param {number | undefined} precision 
* @returns {string}
*/
export function padNumber(number: number, width: number, precision: number | undefined = undefined): string {
  const numberString =
    precision !== undefined ? number.toFixed(precision) : '' + number;
  let decimal = numberString.indexOf('.');
  decimal = decimal === -1 ? numberString.length : decimal;
  return decimal > width
    ? numberString
    : new Array(1 + width - decimal).join('0') + numberString;
}

/**
* 
* @param {string} hemispheres 
* @param {number} degrees 
* @param {number} fractionDigits 
* @returns {string}
*/
export function degreesToStringHDMS(hemispheres: string, degrees: number, fractionDigits: number): string {
  const normalizedDegrees = modulo(degrees + 180, 360) - 180;
  const x = Math.abs(3600 * normalizedDegrees);
  const decimals = fractionDigits || 0;

  let deg = Math.floor(x / 3600);
  let min = Math.floor((x - deg * 3600) / 60);
  let sec = toFixed(x - deg * 3600 - min * 60, decimals);

  if (sec >= 60) {
    sec = 0;
    min += 1;
  }

  if (min >= 60) {
    min = 0;
    deg += 1;
  }

  let hdms = deg + '\u00b0';
  if (min !== 0 || sec !== 0) {
    hdms += ' ' + padNumber(min, 2) + '\u2032';
  }
  if (sec !== 0) {
    hdms += ' ' + padNumber(sec, 2, decimals) + '\u2033';
  }
  if (normalizedDegrees !== 0) {
    hdms += ' ' + hemispheres;
  }

  return hdms;
}

/**
 * @param {GeoJSON.BBox} bbox
 * @param {number} graticuleWidth
 * @param {number} graticuleHeight
 * @param {Units} units
 * @param {string} labelType
 * @param {string} longitudePosition
 * @param {string} latitudePosition
 * @returns {graticuleJson}
 */
export function getGraticule(bbox: GeoJSON.BBox, graticuleWidth: number, graticuleHeight: number, units: Units, labelType: string, longitudePosition: string, latitudePosition: string): graticuleJson {

  const earthCircumference = Math.ceil(distance([0, 0], [180, 0], { units }) * 2);
  const maxColumns = Math.floor(earthCircumference / graticuleWidth);
  const fullDistance: (from: GeoJSON.Position, to: GeoJSON.Position, options: { units: Units; }) => number = (from, to, options): number => {
    const dist = distance(from, to, options);
    if (Math.abs(to[0] - from[0]) >= 180) {
      return earthCircumference - dist;
    }
    return dist;
  };

  const meridians: GeoJSON.Feature<GeoJSON.LineString>[] = [];
  const parallels = [];
  const west = bbox[0];
  const south = bbox[1];
  const east = bbox[2];
  const north = bbox[3];

  // calculate graticule start point
  const deltaX = (west < 0 ? -1 : 1) * fullDistance([0, 0], [west, 0], { units });
  const deltaY = (south < 0 ? -1 : 1) * fullDistance([0, 0], [0, south], { units });
  const startDeltaX = Math.ceil(deltaX / graticuleWidth) * graticuleWidth;
  const startDeltaY = Math.ceil(deltaY / graticuleHeight) * graticuleHeight;
  const startPoint: GeoJSON.Position = [
    destination([0, 0], startDeltaX, 90, { units }).geometry.coordinates[0],
    destination([0, 0], startDeltaY, 0, { units }).geometry.coordinates[1]
  ];

  // calculate graticule columns and rows count
  const width = fullDistance([west, 0], [east, 0], { units });
  const height = fullDistance([0, south], [0, north], { units });
  const columns = Math.min(Math.ceil(width / graticuleWidth), maxColumns);
  const rows = Math.ceil(height / graticuleHeight);

  let currentPoint: GeoJSON.Position;

  // meridians
  currentPoint = startPoint;
  for (let i = 0; i < columns; i++) {
    let coordinates: GeoJSON.Position[];
    if (longitudePosition === 'bottom') {
      coordinates = [
        [currentPoint[0], south],
        [currentPoint[0], north],
      ];
    } else {
      coordinates = [
        [currentPoint[0], north],
        [currentPoint[0], south],
      ];
    }
    let hemisphere;
    if (currentPoint[0] > 0) {
      hemisphere = 'E'
    } else {
      hemisphere = 'W'
    }
    const hdms = degreesToStringHDMS(hemisphere, currentPoint[0], 2)
    let feature: GeoJSON.Feature<GeoJSON.LineString>;
    if (labelType === 'hdms') {
      feature = { type: 'Feature', geometry: { type: 'LineString', coordinates }, properties: { "coord": hdms } };
    } else {
      feature = { type: 'Feature', geometry: { type: 'LineString', coordinates }, properties: { "coord": currentPoint[0].toFixed(3) + "°" } };
    }
    meridians.push(feature);

    currentPoint = [
      destination([currentPoint[0], 0], graticuleWidth, 90, { units }).geometry.coordinates[0],
      currentPoint[1]
    ];
  }

  // parallels
  currentPoint = startPoint;
  for (let i = 0; i < rows; i++) {
    let coordinates: GeoJSON.Position[];
    if (latitudePosition === 'right') {
      coordinates = [
        [east, currentPoint[1]],
        [west, currentPoint[1]],
      ];
    } else {
      coordinates = [
        [west, currentPoint[1]],
        [east, currentPoint[1]],
      ];
    }
    let hemisphere = 'S';
    if (currentPoint[1] > 0) {
      hemisphere = 'N'
    } else {
      hemisphere = 'S'
    }
    const hdms = degreesToStringHDMS(hemisphere, currentPoint[1], 2)
    let feature: GeoJSON.Feature<GeoJSON.LineString>;
    if (labelType === 'hdms') {
      feature = { type: 'Feature', geometry: { type: 'LineString', coordinates }, properties: { "coord": hdms } };
    } else {
      feature = { type: 'Feature', geometry: { type: 'LineString', coordinates }, properties: { "coord": currentPoint[1].toFixed(3) + "°" } };
    }
    parallels.push(feature);

    currentPoint = [
      currentPoint[0],
      destination([0, currentPoint[1]], graticuleHeight, 0, { units }).geometry.coordinates[1]
    ];
  }

  return {
    "meridians": meridians,
    "parallels": parallels
  };
}
