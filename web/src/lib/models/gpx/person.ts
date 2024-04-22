import Link from './link';

export default class Person {
  name: string;
  email: string;
  link?: Link;
  constructor(object: { name: string, email: string, link?: Link }) {
    this.name = object.name;
    this.email = object.email;
    if (object.link) {
      this.link = new Link(object.link);
    }
  }
}