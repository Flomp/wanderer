export default class Copyright {
  author: string;
  year: number;
  license: string;
  constructor(object: { author: string, year: number, license: string }) {
    this.author = object.author;
    this.year = object.year;
    this.license = object.license;
  }
}