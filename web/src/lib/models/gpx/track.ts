import TrackSegment from './track-segment';
import Link from './link';

export default class Track {
  name?: string;
  cmt?: string;
  desc?: string;
  src?: string;
  number?: number;
  type?: string;
  extensions?: string;
  link?: Link[];
  trkseg?: TrackSegment[];
  constructor(object: {
    name?: string,
    cmt?: string,
    desc?: string,
    src?: string,
    number?: number,
    type?: string,
    extensions?: string,
    link?: Link | Link[],
    trkseg?: TrackSegment | TrackSegment[]
  }) {
    this.name = object.name;
    this.cmt = object.cmt;
    this.desc = object.desc;
    this.src = object.src;
    this.number = object.number;
    this.type = object.type;
    this.extensions = object.extensions;
    if (object.link) {
      if (!Array.isArray(object.link)) {
        object.link = [object.link];
      }
      this.link = object.link.map(l => new Link(l));
    }
    if (object.trkseg) {
      if (!Array.isArray(object.trkseg)) {
        object.trkseg = [object.trkseg];
      }
      this.trkseg = object.trkseg.map(trkseg => new TrackSegment(trkseg));;
    }
  }
}