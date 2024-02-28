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

  setSelected: function(s) {
    var selected = !!s;
    if (this._selected !== selected) {
      this._selected = selected;
      this.fire('selected');
    }
  },

  isSelected: function() {
    return !!this._selected;
  },
};

L.Mixin.Selection = {
  includes: L.Mixin.Events,

  getSelection: function() {
    return this._selected;
  },

  setSelection: function(item) {
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
    this.fire('selection_changed');
  },
};

L.Control.LayersLegend = L.Control.Layers.extend({
  _onInputClick: function() {
    this._handlingClick = true;

    this._layerControlInputs.reduceRight((_,input) => {
      if (input.checked) {
        this._map.fireEvent("legend_selected", {
          layer: this._getLayer(input.layerId).layer,
          input: input,
        }, true);
      return input;
     }
    }, 0);

    this._handlingClick = false;

    this._refocusOnMap();
  }
});

L.control.layersLegend = (baseLayers, overlays, options) => new L.Control.LayersLegend(baseLayers, overlays, options);

L.GeoJSON.include(L.Mixin.Selectable);

L.GpxGroup = L.Class.extend({
  options: {
    highlight: {
      color: '#ff0',
      opacity: 1,
    },
    points: [],
    points_options: {
      icon: {
        iconUrl: '../images/elevation-poi.png',
        iconSize: [12, 12],
      }
    },
    flyToBounds: true,
    legend: false,
    legend_options: {
      position: "topright",
      collapsed: false,
    },
    elevation: true,
    elevation_options: {
      theme: 'yellow-theme',
      detached: true,
      elevationDiv: '#elevation-div',
    },
    distanceMarkers: true,
    distanceMarkers_options: {
      lazy: true
    },
  },

  initialize: function(tracks, options) {

    L.Util.setOptions(this, options);

    this._count       = 0;
    this._loadedCount = 0;
    this._tracks      = tracks;
    this._layers      = L.featureGroup();
    this._markers     = L.featureGroup();
    this._elevation   = L.control.elevation(this.options.elevation_options);
    this._legend      = L.control.layersLegend(null, null, this.options.legend_options);

    this.options.points.forEach((poi) => 
      L
        .marker(poi.latlng, { icon: L.icon(this.options.points_options.icon) })
        .bindTooltip(poi.name, { direction: 'auto' }).addTo(this._markers)
    );

  },

  getBounds: function() {
    return this._layers.getBounds();
  },

  addTo: function(map) {
    this._layers.addTo(map);
    this._markers.addTo(map);

    this._map = map;

    this.on('selection_changed', this._onSelectionChanged, this);
    this._map.on('legend_selected', this._onLegendSelected, this);
    this._tracks.forEach(this.addTrack, this);

  },

  addTrack: function(track) {
    if (track instanceof Object) {
      this._loadGeoJSON(track);
    } else {
      fetch(track)
        .then(response => response.ok && response.text())
        .then(text => this._elevation._parseFromString(text))
        .then(geojson => this._loadGeoJSON(geojson, track.split('/').pop().split('#')[0].split('?')[0]));
    }
  },

  _loadGeoJSON: function(geojson, fallbackName) {
    if (geojson) {
      geojson.name = geojson.name || (geojson[0] && geojson[0].properties.name) || fallbackName;
      this._loadRoute(geojson);
    }
  },

  _loadRoute: function(data) {
    if (!data) return;

    var line_style = {
      color: this._uniqueColors(this._tracks.length)[this._count++],
      opacity: 0.75,
      weight: 5,
      distanceMarkers: this.options.distanceMarkers_options,
    };

    var route = L.geoJson(data, {
      name: data.name || '',
      style: (feature) => line_style,
      distanceMarkers: line_style.distanceMarkers,
      originalStyle: line_style,
      filter: feature => feature.geometry.type != "Point",
    });

    
    this._elevation.import(this._elevation.__LGEOMUTIL).then(() => {
      route.addTo(this._layers);

      route.eachLayer((layer) => this._onEachRouteLayer(route, layer));
      this._onEachRouteLoaded(route);
    });

  },

  _onEachRouteLayer: function(route, layer) {
    var polyline = layer;

    route.on('selected', L.bind(this._onRouteSelected, this, route, polyline));

    polyline.on('mouseover', L.bind(this._onRouteMouseOver, this, route, polyline));
    polyline.on('mouseout', L.bind(this._onRouteMouseOut, this, route, polyline));
    polyline.on('click', L.bind(this._onRouteClick, this, route, polyline));

    polyline.bindTooltip(route.options.name, { direction: 'auto', sticky: true, });
  },

  _onEachRouteLoaded: function(route) {
    if (this.options.legend) {
      this._legend.addBaseLayer(route, '<svg id="legend_' + route._leaflet_id + '" width="25" height="10" version="1.1" xmlns="http://www.w3.org/2000/svg">' + '<line x1="0" x2="50" y1="5" y2="5" stroke="' + route.options.originalStyle.color + '" fill="transparent" stroke-width="5" /></svg>' + ' ' + route.options.name);
    }
    
    this.fire('route_loaded', { route: route });

    if (++this._loadedCount === this._tracks.length) {
      this.fire('loaded');
      if (this.options.flyToBounds) {
        this._map.flyToBounds(this.getBounds(), { duration: 0.25, easeLinearity: 0.25, noMoveStart: true });
      }
      if (this.options.legend) {
        this._legend.addTo(this._map);
      }
    }
  },

  highlight: function(route, polyline) {
    polyline.setStyle(this.options.highlight);
    if (this.options.distanceMarkers) {
      polyline.addDistanceMarkers();
    }
  },

  unhighlight: function(route, polyline) {
    polyline.setStyle(route.options.originalStyle);
    if (this.options.distanceMarkers) {
      polyline.removeDistanceMarkers();
    }
  },

  _onRouteMouseOver: function(route, polyline) {
    if (!route.isSelected()) {
      this.highlight(route, polyline);
      if (this.options.legend) {
        this.setSelection(route);
        L.DomUtil.get('legend_' + route._leaflet_id).parentNode.previousSibling.click();
      }
    }
    this.fire('route_mouseover', { route: route, polyline: polyline });
  },

  _onRouteMouseOut: function(route, polyline) {
    if (!route.isSelected()) {
      this.unhighlight(route, polyline);
    }
    this.fire('route_mouseout', { route: route, polyline: polyline });
  },

  _onRouteClick: function(route, polyline) {
    this.setSelection(route);
  },

  _onRouteSelected: function(route, polyline) {
    if (!route.isSelected()) {
      this.unhighlight(route, polyline);
    }
  },

  _onSelectionChanged: function(e) {
    var elevation = this._elevation;
    var eleDiv = elevation.getContainer();
    var route = this.getSelection();

    elevation.clear();

    if (route && route.isSelected()) {
      if (!eleDiv) {
        elevation.addTo(this._map);
      }
      route.getLayers().forEach(function(layer) {
        if (layer instanceof L.Polyline) {
          elevation.addData(layer, false);
          layer.bringToFront();
        }
      });
    } else {
      if (eleDiv) {
        elevation.remove();
      }
    }
  },

  _onLegendSelected: function(e) {
    var parent = e.input.closest('.leaflet-control-layers-list');
    var route = e.layer;

    if (!route.isSelected()) {
      this.setSelection(route);
      for (var i in route._layers) {
        this.highlight(route, route._layers[i]);
      }
      this._map.flyToBounds(e.layer.getBounds());
    }

    parent.scroll({ top: (e.input.offsetTop - parent.offsetTop) || 0, behavior: 'smooth' });

    this._layers.eachLayer(layer => {
      var legend = L.DomUtil.get('legend_' + layer._leaflet_id);
      legend.querySelector("line").style.stroke = layer.isSelected() ? this.options.highlight.color : "";
      legend.parentNode.style.fontWeight = layer.isSelected() ? "700" : "";
    });
  },

  _uniqueColors: function(count) {
    return count === 1 ? ['#0000ff'] : new Array(count).fill(null).map((_,i) => this._hsvToHex(i * (1 / count), 1, 0.7));
  },

  _hsvToHex: function(h, s, v) {
    var i = Math.floor(h * 6);
    var f = h * 6 - i;
    var p = v * (1 - s);
    var q = v * (1 - f * s);
    var t = v * (1 - (1 - f) * s);
    var rgb = { 0: [v, t, p], 1: [q, v, p], 2: [p, v, t], 3: [p, q, v], 4: [t, p, v], 5: [v, p, q] }[i % 6];
    return rgb.map(d => d * 255).reduce((hex, byte) => hex + ((byte >> 4) & 0x0F).toString(16) + (byte & 0x0F).toString(16), "#");
  },

  removeFrom: function(map) {
    this._layers.removeFrom(map);
  },

});

L.GpxGroup.include(L.Mixin.Events);
L.GpxGroup.include(L.Mixin.Selection);

L.gpxGroup = (tracks, options) => new L.GpxGroup(tracks, options);