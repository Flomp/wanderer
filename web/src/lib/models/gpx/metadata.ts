import Copyright from './copyright';
import Link from './link';
import Person from './person';
import Bounds from './bounds';

export default class Metadata {
  name: string;
  desc: string;
  time: Date;
  keywords: string;
  extensions?: string;
  author?: Person;
  link?: Link[];
  bounds?: Bounds;
  copyright?: Copyright;
  constructor(object: {
    name: string,
    desc: string,
    time: string,
    keywords: string,
    extensions?: string,
    author?: Person,
    link?: Link | Link[],
    bounds?: Bounds,
    copyright?: Copyright
  }) {
    this.name = object.name;
    this.desc = object.desc;
    this.time = object.time ? new Date(object.time) : new Date();
    this.keywords = object.keywords;
    this.extensions = object.extensions;
    if (object.author) {
      this.author = new Person(object.author);
    }
    if (object.link) {
      if (!Array.isArray(object.link)) {
        object.link = [object.link];
      }
      this.link = object.link.map(l => new Link(l));
    }
    if (object.bounds) {
      this.bounds = new Bounds(object.bounds);
    }
    if (object.copyright) {
      this.copyright = new Copyright(object.copyright);
    }
  }
}