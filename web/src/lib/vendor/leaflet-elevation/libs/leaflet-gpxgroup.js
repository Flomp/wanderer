/*
 * https://github.com/adoroszlai/joebed/tree/gh-pages
 *
 * The MIT License (MIT)
 *
 * Copyright (c) 2014- Doroszlai Attila, 2019- Raruto
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in
 * the Software without restriction, including without limitation the rights to
 * use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
 * the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

L.Mixin.Selectable = {
  includes: L.Mixin.Events,

  setSelected: function (s) {
    var selected = !!s;
    if (this._selected !== selected) {
      this._selected = selected;
      this.fire('selected');
    }
  },

  isSelected: function () {
    return !!this._selected;
  },
};

L.Mixin.Selection = {
  includes: L.Mixin.Events,

  getSelection: function () {
    return this._selected;
  },

  setSelection: function (item) {
    if (this._selected === item) {
      if (item !== null) {
        item.setSelected(!item.isSelected());
        if (!item.isSelected()) {
          this._selected = null;
        }
      }
    } else {
      if (this._selected) {
        this._selected.setSelected(false);
      }
      this._selected = item;
      if (this._selected) {
        this._selected.setSelected(true);
      }
    }
    this.fire('selection_changed', { polyline: item });
  },
};

L.GeoJSON.include(L.Mixin.Selectable);

export const GpxGroup = L.GpxGroup = L.Class.extend({
  options: {
    highlight: {
      opacity: 1,
      weight: 6,
    },
    points: [],
    points_options: {
      icon: {
        iconUrl: '../images/elevation-poi.png',
        iconSize: [12, 12],
      }
    },
    flyToBounds: true,
    elevation: true,
    elevation_options: {
      theme: 'lightblue-theme',
      detached: true,
      elevationDiv: '#elevation',
    },
    distanceMarkers: true,
    distanceMarkers_options: {
      lazy: true,
      distance: false,
      direction: true,
      offset: 1000,
    },
  },

  initialize: function (tracks, options) {

    L.Util.setOptions(this, options);

    this._count = 0;
    this._loadedCount = 0;
    this._tracks = tracks;
    this._layers = L.featureGroup();
    this._markers = L.featureGroup();
    this._hotline = L.featureGroup();
    this._elevation = L.control.elevation(this.options.elevation_options);

    this._routes = [];

    this.options.points.forEach((poi) =>
      L
        .marker(poi.latlng, { icon: L.icon(this.options.points_options.icon) })
        .bindTooltip(poi.name, { direction: 'auto' }).addTo(this._markers)
    );

  },

  getBounds: function () {
    return this._layers.getBounds();
  },

  addTo: function (map) {
    this._hotline.addTo(map);
    this._layers.addTo(map);
    this._markers.addTo(map);

    this._map = map;

    this.on('selection_changed', this._onSelectionChanged, this);
    this.addTracks();
  },

  addTracks() {
    this._tracks.forEach(this._addTrack, this);

  },

  highlightTrack(id) {
    const route = this._routes.find(r => r.options.id == id)
    if (!route) {
      return;
    }
    route.eachLayer((l) => {
      this.highlight(route, l)
    })
  },

  unHighlightTrack(id) {
    const route = this._routes.find(r => r.options.id == id)
    if (!route) {
      return;
    }
    route.eachLayer((l) => {
      this.unhighlight(route, l)
    })
  },

  select(id) {
    const route = this._routes.find(r => r.options.id == id)
    if (!route) {
      return;
    }
    this.setSelection(route)
  },

  resetSelection() {
    this.setSelection(this._selected)
    if (this.options.flyToBounds) {
      this._map.flyToBounds(this.getBounds(), { duration: 0.25, easeLinearity: 0.25, noMoveStart: true });
    }
  },

  openPopup(id) {
    const route = this._routes.find(r => r.options.id == id)
    if (!route) {
      return;
    }
    route.openPopup();
  },

  closePopup(id) {
    const route = this._routes.find(r => r.options.id == id)
    if (!route) {
      return;
    }
    route.closePopup();
  },

  _addTrack: function (track) {
    this._elevation._parseFromString(track.gpx)
      .then(geojson => this._loadGeoJSON(geojson, track.id, track.gpx.split('/').pop().split('#')[0].split('?')[0]))

  },

  clear: function () {
    this._elevation.clear()
    this._elevation.remove();
    this._clearLayers();
    this._clearLayers(this._markers);
    this._clearLayers(this._hotline)
    this._count = 0;
    this._loadedCount = 0;
    this._tracks = []
    this._routes = []

    this.fire('clear')
  },

  _clearLayers(l) {
    l = l || this._layers;
    if (l && l.eachLayer) {
      l.eachLayer(f => f.remove())
      l.clearLayers();
    }
  },

  _loadGeoJSON: function (geojson, id, fallbackName) {
    if (geojson) {
      geojson.id = id
      geojson.name = geojson.name || (geojson[0] && geojson[0].properties.name) || fallbackName;
      this._loadRoute(geojson);
    }
  },

  _loadRoute: function (data) {
    if (!data) return;
    var line_style = {
      color: this._stringToColor(data.id),
      opacity: 0.75,
      weight: 5,
      distanceMarkers: this.options.distanceMarkers_options,
    };

    var route = L.geoJson(data, {
      name: data.name || '',
      style: (feature) => line_style,
      distanceMarkers: line_style.distanceMarkers,
      originalStyle: line_style,
      isGroupLayer: true,
      id: data.id,
      filter: feature => feature.geometry.type != "Point",
    });


    this._elevation.import([this._elevation.__LGEOMUTIL, this._elevation.__LDISTANCEM]).then(() => {
      route.addTo(this._layers);

      route.eachLayer((layer) => this._onEachRouteLayer(route, layer));


      this._onEachRouteLoaded(route);
    });

  },

  _onEachRouteLayer: function (route, layer) {
    var polyline = layer;

    this._routes.push(route)

    route.on('selected', L.bind(this._onRouteSelected, this, route, polyline));

    polyline.on('mouseover', L.bind(this._onRouteMouseOver, this, route, polyline));
    polyline.on('mouseout', L.bind(this._onRouteMouseOut, this, route, polyline));
    polyline.on('click', L.bind(this._onRouteClick, this, route, polyline));

    const startIcon = L.divIcon({
      html: '<i class="p-2 text-white bg-gray-500 rounded-full fa fa-bullseye -translate-x-1/2"></i>',
      className: 'start-icon'
    });
    const endIcon = L.divIcon({
      html: '<i class="p-2 text-white bg-gray-500 rounded-full fa fa-flag-checkered -translate-x-1/2"></i>',
      className: 'end-icon'
    });
    const latlngs = polyline.getLatLngs().flat(1);

    polyline.setLatLngs(latlngs);

    if (this._loadedCount == 0 || !this.options.itinerary) {
      L.marker(latlngs[0], { icon: startIcon }).addTo(this._markers)
    }

    L.marker(latlngs[latlngs.length - 1], { icon: endIcon }).addTo(this._markers)
  },

  _onEachRouteLoaded: function (route) {
    this.fire('route_loaded', { route: route });

    if (++this._loadedCount === this._tracks.length) {
      this.fire('loaded');
      if (this.options.flyToBounds) {
        this._map.flyToBounds(this.getBounds(), { duration: 0.25, easeLinearity: 0.25, noMoveStart: true });
      }
    }
  },

  highlight: function (route, polyline) {
    polyline.setStyle(this.options.highlight);
    polyline.options.highlighted = true
    if (this.options.distanceMarkers) {
      polyline.addDistanceMarkers();
    }
  },

  unhighlight: function (route, polyline) {
    polyline.setStyle(route.options.originalStyle);
    polyline.options.highlighted = false
    if (this.options.distanceMarkers) {
      polyline.removeDistanceMarkers();
    }
  },

  _onRouteMouseOver: function (route, polyline) {
    if (!route.isSelected()) {
      this.highlight(route, polyline);
    }
    this.fire('route_mouseover', { route: route, polyline: polyline });
  },

  _onRouteMouseOut: function (route, polyline) {
    if (!route.isSelected()) {
      this.unhighlight(route, polyline);
    }
    this.fire('route_mouseout', { route: route, polyline: polyline });
  },

  _onRouteClick: function (route, polyline) {
    this.highlight(route, polyline)
    this.setSelection(route);
  },

  _onRouteSelected: function (route, polyline) {
    if (!route.isSelected()) {
      this.unhighlight(route, polyline);
    }
  },

  _onSelectionChanged: async function (e) {
    var elevation = this._elevation;
    var eleDiv = elevation.getContainer();
    var route = this.getSelection();
    var hotline = this._hotline;

    hotline.clearLayers();

    if (route && route.isSelected()) {
      elevation.clear();

      if (!eleDiv) {
        elevation.addTo(this._map);
        this._map.invalidateSize()
      }
      for (const layer of route.getLayers()) {
        if (layer instanceof L.Polyline) {
          await elevation.addData(layer, false);
        }
      }
      await elevation._initHotLine(route, this._hotline)

      
      this._map.flyToBounds(route.getBounds(), { duration: 0.25, easeLinearity: 0.25, noMoveStart: true });


    } else {
      if (eleDiv) {
        elevation.remove();
        this._map.invalidateSize()
      }
    }
  },

  _stringToColor(input) {
    // Define the maximum brightness to ensure the color is not too light
    const maxBrightness = 200;

    // Divide the string into 3 parts (5 characters each)
    const redPart = input.slice(0, 5);
    const greenPart = input.slice(5, 10);
    const bluePart = input.slice(10, 15);

    // Calculate the red, green, and blue values by summing the char codes of the parts
    const red = Math.floor((redPart.split('').reduce((sum, char) => sum + char.charCodeAt(0), 0) % 256) * (maxBrightness / 255));
    const greenRaw = greenPart.split('').reduce((sum, char) => sum + char.charCodeAt(0), 0) % 256;
    const green = Math.floor((greenRaw * 0.5) * (maxBrightness / 255)); // Reduce green to avoid greenish colors
    const blue = Math.floor((bluePart.split('').reduce((sum, char) => sum + char.charCodeAt(0), 0) % 256) * (maxBrightness / 255));

    // Format as HEX color
    return `#${(1 << 24 | red << 16 | green << 8 | blue).toString(16).slice(1).toUpperCase()}`;
  },

  removeFrom: function (map) {
    this._layers.removeFrom(map);
  },

});

L.GpxGroup.include(L.Mixin.Events);
L.GpxGroup.include(L.Mixin.Selection);

L.gpxGroup = (tracks, options) => new L.GpxGroup(tracks, options);