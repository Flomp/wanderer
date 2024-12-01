
class Settings {
  id?: string;
  unit?: "metric" | "imperial";
  language?: "en" | "de" | "fr" | "hu"| "it" | "nl" | "pl" | "pt" | "zh";
  mapFocus?: "trails" | "location";
  location?: { name: string, lat: number, lon: number };
  category?: string;
  tilesets?: {name: string, url: string}[]
  terrain?: string;
  user?: string;

  constructor(
    unit: "metric" | "imperial",
    language: "en" | "de" | "fr" | "hu"| "it" | "nl" | "pl" | "pt" | "zh",
    mapFocus: "trails" | "location",
    user: string,
    params?: {
      location?: { name: string, lat: number, lon: number }
      category?: string
      tilesets?: {name: string, url: string}[]
      terrain?: string;
    }
  ) {
    this.unit = unit;
    this.language = language;
    this.mapFocus = mapFocus;
    this.user = user;
    this.location = params?.location;
    this.category = params?.category;
    this.tilesets = params?.tilesets ?? [];
    this.terrain = params?.terrain;
  }
}


export { Settings };
