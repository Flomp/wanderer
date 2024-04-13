
class Settings {
  id?: string;
  unit?: "metric" | "imperial";
  language?: "en" | "de" | "fr" | "hu" | "nl" | "pl" | "pt" | "zh";
  mapFocus?: "trails" | "location";
  location?: { name: string, lat: number, lon: number };
  user?: string;

  constructor(
    unit: "metric" | "imperial",
    language: "en" | "de" | "fr" | "hu" | "nl" | "pl" | "pt" | "zh",
    mapFocus: "trails" | "location",
    user: string,
    params?: {
      location: { name: string, lat: number, lon: number }
    }
  ) {
    this.unit = unit;
    this.language = language;
    this.mapFocus = mapFocus;
    this.user = user;
    this.location = params?.location;
  }
}


export { Settings };
