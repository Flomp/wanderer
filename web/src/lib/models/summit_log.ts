import type { Trail } from "./trail";

class SummitLog {
  id?: string;
  date: string;
  text?: string;
  gpx?: string;
  _gpx: File | null;
  distance?: number
  elevation_gain?: number
  elevation_loss?: number
  duration?: number
  author?: string;

  expand: {
    gpx_data?: string;
    trails_via_summit_logs?: Trail[];
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
