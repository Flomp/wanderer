export default class Bounds {
  minlat: number;
  minlon: number;
  maxlat: number;
  maxlon: number;
  constructor(object: {minlat: number, minlon: number, maxlat: number, maxlon: number}) {
    this.minlat = object.minlat;
    this.minlon = object.minlon;
    this.maxlat = object.maxlat;
    this.maxlon = object.maxlon;
  }
}
