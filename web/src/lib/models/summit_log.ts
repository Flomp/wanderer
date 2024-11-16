import type { Trail } from "./trail";
import type { UserAnonymous } from "./user";

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
    author?: UserAnonymous
  }

  constructor(date: string, params?: { id?: string, text?: string, distance?: number, elevation_loss?: number, elevation_gain?: number, duration?: number }) {
    this.date = date;
    this.id = params?.id;
    this.text = params?.text ?? "";
    this.expand = {}
    this.distance = params?.distance
    this.elevation_gain = params?.elevation_gain
    this.elevation_loss = params?.elevation_loss
    this.duration = params?.duration
    this._gpx = null
  }
}

interface SummitLogFilter {
  category: string[],
  startDate?: string;
  endDate?: string;
}

export { SummitLog, type SummitLogFilter };
