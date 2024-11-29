import { getGraticule } from "./generator";
import * as M from "maplibre-gl";

export interface GraticuleConfig {
  minZoom?: number;
  maxZoom?: number;
  showLabels?: boolean;
  labelType?: "hdms" | "decimal";
  labelSize?: number;
  labelColor?: string;
  longitudePosition?: "top" | "bottom";
  latitudePosition?: "left" | "right";
  longitudeOffset?: number[];
  latitudeOffset?: number[];
  paint?: any;
}

type graticuleJson = {
  meridians: GeoJSON.Feature<GeoJSON.LineString>[];
  parallels: GeoJSON.Feature<GeoJSON.LineString>[];
};

export function randomString() {
  return Math.floor(Math.random() * 10e12).toString(36);
}

/**
 *
 * @param {Map} map
 * @returns
 */
function calculateResolution(map: M.Map) {
  const zoom = map.getZoom();
  const container = map.getContainer();
  const containerWidth = container.offsetWidth;
  const containerHeight = container.offsetHeight;

  const tileSize = 256;
  const metersPerPixel = (Math.PI * 2 * 6378137) / (tileSize * 2 ** zoom);

  const resolutionX = metersPerPixel * (containerWidth / tileSize);
  const resolutionY = metersPerPixel * (containerHeight / tileSize);

  return {
    x: resolutionX,
    y: resolutionY,
  };
}

export default class MaplibreGraticule {
  xid: string;
  yid: string;
  config: GraticuleConfig;
  updateBound: () => void;
  labelSize: any;
  map: any;
  constructor(config: GraticuleConfig) {
    this.xid = `graticule-meridains-${randomString()}`;
    this.yid = `graticule-parallels-${randomString()}`;
    this.config = config;

    this.updateBound = this.update.bind(this);
    this.labelSize = this.config.labelSize;
  }

  /**
   * @param {Map} map
   * @returns {HTMLElement}
   */
  onAdd(map: M.Map): HTMLElement {
    this.map = map;

    this.map.on("load", this.updateBound);
    this.map.on("move", this.updateBound);

    if (this.map.loaded()) {
      this.update();
    }

    return document.createElement("div");
  }

  /**
   * @returns {void}
   */
  onRemove(): void {
    if (!this.map) {
      return;
    }

    // Remove symbol layers
    const xSymbolLayer = this.map.getLayer(`symbols${this.xid}`);
    if (xSymbolLayer) {
      this.map.removeLayer(`symbols${this.xid}`);
    }
    const ySymbolLayer = this.map.getLayer(`symbols${this.yid}`);
    if (ySymbolLayer) {
      this.map.removeLayer(`symbols${this.yid}`);
    }

    const xsource = this.map.getSource(this.xid);
    if (xsource) {
      this.map.removeLayer(this.xid);
      this.map.removeSource(this.xid);
    }

    const ysource = this.map.getSource(this.yid);
    if (ysource) {
      this.map.removeLayer(this.yid);
      this.map.removeSource(this.yid);
    }

    this.map.off("load", this.updateBound);
    this.map.off("move", this.updateBound);

    this.map = undefined;
  }

  /**
   * @returns {void}
   */
  update(): void {
    if (!this.map) {
      return;
    }

    const longitudePosition = this.config.longitudePosition ?? "bottom";
    const latitudePosition = this.config.latitudePosition ?? "right";
    const labelType = this.config.labelType ?? "hdms";

    const resolution = calculateResolution(this.map);
    const xWidth = (resolution.x / 100) * 2;
    const yWidth = (resolution.y / 100) * 2;

    /** @type {graticuleJson} */
    let graticule: graticuleJson = {
      meridians: [],
      parallels: [],
    };
    if (this.active) {
      graticule = getGraticule(
        this.bbox,
        xWidth,
        yWidth,
        "kilometers",
        labelType,
        longitudePosition,
        latitudePosition
      );
    }

    const xsource: M.GeoJSONSource = this.map.getSource(this.xid);
    if (!xsource) {
      this.map.addSource(this.xid, {
        type: "geojson",
        data: { type: "FeatureCollection", features: graticule.meridians },
      });

      this.map.addLayer({
        id: this.xid,
        source: this.xid,
        type: "line",
        paint: this.config.paint ?? {},
      });

      if (this.config.showLabels) {
        this.map.addLayer({
          id: `symbols${this.xid}`,
          type: "symbol",
          source: this.xid,
          layout: {
            "symbol-placement": "point",
            "text-field": "{coord}",
            "text-size": this.config.labelSize ?? 12,
            "text-anchor": longitudePosition === "top" ? "top" : "bottom",
            "text-offset": this.config.longitudeOffset ?? [0, 0],
          },
          paint: {
            "text-color": this.config.labelColor ?? "#000000",
            "text-halo-blur": 1,
            "text-halo-color": "rgba(255,255,255,1)",
            "text-halo-width": 3,
          },
        });
      }
    } else {
      xsource.setData({
        type: "FeatureCollection",
        features: graticule.meridians,
      });
    }

    const ysource: M.GeoJSONSource = this.map.getSource(this.yid);
    if (!ysource) {
      this.map.addSource(this.yid, {
        type: "geojson",
        data: { type: "FeatureCollection", features: graticule.parallels },
      });

      this.map.addLayer({
        id: this.yid,
        source: this.yid,
        type: "line",
        paint: this.config.paint ?? {},
      });

      if (this.config.showLabels) {
        this.map.addLayer({
          id: `symbols${this.yid}`,
          type: "symbol",
          source: this.yid,
          layout: {
            "symbol-placement": "point",
            "text-field": "{coord}",
            "text-size": this.config.labelSize ?? 12,
            "text-anchor": latitudePosition === "left" ? "left" : "right",
            "text-offset": this.config.latitudeOffset ?? [0, 0],
          },
          paint: {
            "text-color": this.config.labelColor ?? "#000000",
            "text-halo-blur": 1,
            "text-halo-color": "rgba(255,255,255,1)",
            "text-halo-width": 3,
          },
        });
      }
    } else {
      ysource.setData({
        type: "FeatureCollection",
        features: graticule.parallels,
      });
    }
  }

  /**
   * @returns {boolean}
   */
  get active(): boolean {
    if (!this.map) {
      return false;
    }

    const minZoom = this.config.minZoom ?? 0;
    const maxZoom = this.config.maxZoom ?? 22;
    const zoom = this.map.getZoom();

    return minZoom <= zoom && zoom < maxZoom;
  }

  /**
   * @returns {GeoJSON.BBox}
   */
  get bbox(): GeoJSON.BBox {
    if (!this.map) {
      throw new Error("Invalid state");
    }

    const bounds = this.map.getBounds();
    if (bounds.getEast() - bounds.getWest() >= 360) {
      bounds.setNorthEast([bounds.getWest() + 360, bounds.getNorth()]);
    }

    const bbox = /** @type {GeoJSON.BBox} */ bounds.toArray().flat();
    return bbox;
  }
}