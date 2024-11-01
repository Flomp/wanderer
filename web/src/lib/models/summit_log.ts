
class SummitLog {
  id?: string;
  date: string;
  text?: string;
  gpx?: string;
  _gpx: File | null;
  author?: string;

  expand: {
    gpx_data?: string;
  }

  constructor(date: string, params?: { id?: string, text?: string }) {
    this.date = date;
    this.id = params?.id;
    this.text = params?.text ?? "";
    this.expand = {}
    this._gpx = null
  }
}


export { SummitLog };
