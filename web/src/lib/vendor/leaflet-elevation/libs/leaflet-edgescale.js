/*
 * Copyright (c) 2023, GPL-3.0+ Project, Raruto
 *
 *  This file is free software: you may copy, redistribute and/or modify it
 *  under the terms of the GNU General Public License as published by the
 *  Free Software Foundation, either version 2 of the License, or (at your
 *  option) any later version.
 *
 *  This file is distributed in the hope that it will be useful, but
 *  WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 *  General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with this program.  If not, see .
 *
 * This file incorporates work covered by the following copyright and
 * permission notice:
 * 
 *     Copyright (c) GPL-3.0+ Project  - 2018- Dražen Tutić - https://github.com/dtutic/Leaflet.EdgeScaleBar
 *     Copyright (c) MIT License (MIT) - 2015- Xisco Guaita - https://github.com/xguaita/Leaflet.MapCenterCoord
 */

/**
 * Original source: https://github.com/xguaita/Leaflet.MapCenterCoord
 */
L.Control.EdgeScale = L.Control.extend({
  
  // Defaults
  options: {
    position: 'bottomleft',
    icon: true,
    coords: true,
    bar: true,
    onMove: true,
    template: '{y} | {x}', // https://en.wikipedia.org/wiki/ISO_6709
    projected: false,
    formatProjected: '#.##0,000',
    latlngFormat: 'DD', // DD, DM, DMS
    latlngDesignators: true,
    latLngFormatter: undefined,
    iconStyle: {
      background: `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' xml:space='preserve' viewBox='0 0 100 100'%3E%3Cg stroke='%23fff'%3E%3Ccircle cx='50' cy='50.2' r='3.9' stroke-width='2' /%3E%3Cpath stroke-width='3' d='M5 54h32a4 4 0 1 0 0-8H5a4 4 0 1 0 0 8z M54 5a4 4 0 1 0-8 0v32a4 4 0 1 0 8 0V5z M99 50c0-2-2-4-4-4H63a4 4 0 1 0 0 8h32c2 0 4-1 4-4zM46 95a4 4 0 1 0 8 0V64a4 4 0 1 0-8 0v31z'/%3E%3C/g%3E%3C/svg%3E%0A")`,
      width:      '24px',
      height:     '24px',
      left:       'calc(50% - 12px)',
      top:        'calc(50% - 12px)',
      content:    '',
      display:    'block',
      position:   'absolute',
      zIndex:     999,
      pointerEvents: 'none',
    },
    containerStyle: {
      backgroundColor: 'rgba(255, 255, 255, 0.7)',
      boxShadow:       '0 0 5px #bbb',
      borderRadius:    '3px',
      padding:         '3px 2px',
      color:           '#333',
      font:            '11px/1.5 Consolas, monaco, monospace',
      writingMode:     'vertical-lr',
    },
  },

  initialize: function(options) {
    L.setOptions(this, options);
  },

  onAdd: function (map) {

    if (this.options.bar) {
      this._scaleBar = (new L.Control.EdgeScale.Layer(true === this.options.bar ? {} : this.options.bar)).addTo(map);
    }

    // create a DOM element and put it into overlayPane
    if (this.options.icon) {
      this._icon = L.DomUtil.create('div', 'leaflet-crosshair');
      Object.assign(this._icon.style, this.options.iconStyle);
      map.getContainer().insertBefore(this._icon, map.getContainer().firstChild);
    }

    // Control container
    this._container = L.DomUtil.create('div', 'leaflet-control-mapcentercoord');

    Object.assign(this._container.style, this.options.containerStyle);

    if (!this.options.coords) {
      this._container.style.display = 'none';
    }

    L.DomEvent.disableClickPropagation(this._container);

    this._container.innerHTML = this._getMapCenterCoord();

    // Add events listeners for updating coordinates & icon's position
    map.on('move', this._onMapMove, this);
    map.on('moveend', this._onMapMove, this);

    return this._container;
  },

  onRemove: function (map) {
    if (this.options.bar) {
      this._scaleBar.remove();
    }

    // remove icon's DOM elements and listeners
    if (this.options.icon) {
      map.getContainer().removeChild(this._icon);
    }
    map.off('move', this._onMapMove, this);
    map.off('moveend', this._onMapMove, this);
  },

  // update coordinates
  _onMapMove: function (e) {
    if (this.options.onMove || 'moveend' === e.type) {
      this._container.innerHTML = this._getMapCenterCoord();
    }
  },

  _getMapCenterCoord: function () {
    const center = this._map.getCenter();
    return this.options.projected
      ? this._getProjectedCoord(this._map.options.crs.project(center))
      : this._getLatLngCoord(center);
  },

  _getProjectedCoord: function (center) {
    return L.Util.template(
      this.options.template,
      {
        x: this._format(this.options.formatProjected, center.x),
        y: this._format(this.options.formatProjected, center.y)
      }
    );
  },

  _getLatLngCoord: function (latLng) {

    const { latLngFormatter, latlngFormat, latlngDesignators: designators } = this.options;

    if (undefined !== latLngFormatter ) {
      return latLngFormatter(latLng.lat, latLng.lng);
    }

    let lat, lng, deg, min;

    // make a copy of center so we aren't affecting leaflet's internal state
    let center = {
      lat: latLng.lat,
      lng: latLng.lng,
      lng_neg: latLng.lng < 0,
      lat_neg: latLng.lat < 0, 
    };

    // 180 degrees & negative
    if (center.lng < 0) {
      center.lng = Math.abs(center.lng);
    }
    if (center.lng > 180) {
      center.lng = 360 - center.lng;
      center.lng_neg = !center.lng_neg;
    }
    if (center.lat < 0) {
      center.lat = Math.abs(center.lat);
    }

    // format
    if ('DM' === latlngFormat) {
      deg = parseInt(center.lng);
      lng = deg + 'º ' + this._format('00.000', (center.lng - deg) * 60) + "'";
      deg = parseInt(center.lat);
      lat = deg + 'º ' + this._format('00.000', (center.lat - deg) * 60) + "'";
    } else if ('DMS' === latlngFormat) {
      deg = parseInt(center.lng);
      min = (center.lng - deg) * 60;
      lng = deg + 'º ' + this._format('00', parseInt(min)) + "' " + this._format('00.0', (min - parseInt(min)) * 60) + "''";
      deg = parseInt(center.lat);
      min = (center.lat - deg) * 60;
      lat = deg + 'º ' + this._format('00', parseInt(min)) + "' " + this._format('00.0', (min - parseInt(min)) * 60) + "''";
    } else { // 'DD'
      lng = this._format('#0.00000', center.lng) + 'º';
      lat = this._format('##0.00000', center.lat) + 'º';
    }

    return L.Util.template(this.options.template, {
      x: (!designators && center.lng_neg ? '-' : '') + lng + (designators ? (center.lng_neg ? ' W' : ' E') : ''),
      y: (!designators && center.lat_neg ? '-' : '') + lat + (designators ? (center.lat_neg ? ' S' : ' N') : '')
    });
  },

  /**
   * IntegraXor Web SCADA - JavaScript Number Formatter
   * 
   * @see https://code.google.com/p/javascript-number-formatter
   * @authors KPL, KHL 
   */
  _format: function (m, v) {
    if (!m || isNaN(+v)) {
      return v;                                                    // return as it is.
    }

    v = m.charAt(0) == '-' ? -v : +v;                              // convert any string to number according to formation sign.
    let isNegative = v < 0 ? v = -v : 0;                           // process only abs(), and turn on flag.

    let result = m.match(/[^\d\-\+#]/g);                           // search for separator for grp & decimal, anything not digit, not +/- sign, not #.
    let Decimal = (result && result[result.length - 1]) || '.';    // treat the right most symbol as decimal
    let Group = (result && result[1] && result[0]) || ',';         // treat the left most symbol as group separator

    m = m.split(Decimal);                                          // split the decimal for the format string if any.
    v = v.toFixed(m[1] && m[1].length);                            // Fix the decimal first, toFixed will auto fill trailing zero.
    v = +(v) + '';                                                 // convert number to string to trim off *all* trailing decimal zero(es)

    let pos_trail_zero = m[1] && m[1].lastIndexOf('0');            // fill back any trailing zero according to format (look for last zero in format)
    let part = v.split('.');

    if (!part[1] || part[1] && part[1].length <= pos_trail_zero) { // integer will get !part[1]
      v = (+v).toFixed(pos_trail_zero + 1);
    }
    let szSep = m[0].split(Group);                                 // look for separator
    m[0] = szSep.join('');                                         // join back without separator for counting the pos of any leading 0.

    let pos_lead_zero = m[0] && m[0].indexOf('0');
    if (pos_lead_zero > -1) {
      while (part[0].length < (m[0].length - pos_lead_zero)) {
        part[0] = '0' + part[0];
      }
    } else if (+part[0] == 0) {
      part[0] = '';
    }
    v = v.split('.');
    v[0] = part[0];

    var pos_separator = (szSep[1] && szSep[szSep.length - 1].length); // process the first group separator from decimal (.) only, the rest ignore. Get the length of the last slice of split result.
    if (pos_separator) {
      let integer = v[0];
      let str = '';
      let offset = integer.length % pos_separator;
      for (let i = 0, l = integer.length; i < l; i++) {
        str += integer.charAt(i);                                  // ie6 only support charAt for sz.
        if (
          !((i - offset + 1) % pos_separator) &&
          i < l - pos_separator                                    // -pos_separator so that won't trail separator on full length
        ) {
          str += Group;
        }
      }
      v[0] = str;
    }

    v[1] = (m[1] && v[1]) ? Decimal + v[1] : "";
    return (isNegative ? '-' : '') + v[0] + v[1];                  // put back any negation and combine integer and fraction.
  }

});

/**
 * Original Source: https://github.com/dtutic/Leaflet.EdgeScaleBar
 * 
 * Draws the metric scale bars in Web Mercator map along top and right edges. 
 * Authors: Dražen Tutić (dtutic@geof.hr), Ana Kuveždić Divjak (akuvezdic@geof.hr)
 * University of Zagreb, Faculty of Geodesy, GEOF-OSGL Lab
 *  Inspired by LatLonGraticule Leaflet plugin by: lanwei@cloudybay.com.tw
 */
L.Control.EdgeScale.Layer = L.Layer.extend({

  includes: L.Evented ? L.Evented.prototype : L.Mixin.Events,

  options: {
      opacity: 1,
      weight: 0.8,
      gradient: {
        size: 10,
        opacity: 0.5,
      }, 
      color: '#000',
      font: '11px Arial',
      zoomInterval: [
          {start: 0, end: 2, interval: 5000000},
          {start: 3, end: 3, interval: 2000000},
          {start: 4, end: 4, interval: 1000000},
          {start: 5, end: 5, interval: 500000},
          {start: 6, end: 7, interval: 200000},
          {start: 8, end: 8, interval: 100000},
          {start: 9, end: 9, interval: 50000},
          {start: 10, end: 10, interval: 20000},
          {start: 11, end: 11, interval: 10000},
          {start: 12, end: 12, interval: 5000},
          {start: 13, end: 13, interval: 2000},
          {start: 14, end: 14, interval: 1000},
          {start: 15, end: 15, interval: 500},
          {start: 16, end: 16, interval: 200},
          {start: 17, end: 17, interval: 100},
          {start: 18, end: 18, interval: 50},
          {start: 19, end: 19, interval: 20},
          {start: 20, end: 20, interval: 10}
      ],
      pane: 'edgescalePane'
  },

  initialize: function (options) {

    L.setOptions(this, options);

    // Constants of the WGS84 ellipsoid needed to calculate meridian length or latitute
    const a   = this._a   = 6378137.0;
    const b   = this._b   = 6356752.3142;
    const n   = this._n   = (a - b)/(a + b);
    const a2  = a * a;
    const b2  = b * b;
    const n2  = n * n;
    const n3  = n2 * n;
    const n4  = n3 * n;
    const n5  = n4 * n;
    this._A   = a * (1.0 - n) * (1.0 - n2) * (1.0 + 9.0/4.0 * n2 + 225.0/64.0 * n4);
    this._e2  = (a2 - b2) / a2;
    this._ic1 = 1.5 * n - 29.0/12.0 * n3 + 553.0/80.0 * n5;
    this._ic2 = 21.0/8.0 * n2 - 1537.0/128.0 * n4;
    this._ic3 = 151.0/24.0 * n3 - 32373.0/640.0 * n5;
    this._ic4 = 1097.0/64.0 * n4;
    this._ic5 = 8011.0/150.0 * n5;
    this._c1  = -1.5 * n + 31.0/24.0 * n3 - 669.0/640.0 * n5;
    this._c2  = 15.0/18.0 * n2 - 435.0/128.0 * n4;
    this._c3  = -35.0/12.0 * n3 + 651.0/80.0 * n5;
    this._c4  = 315.0/64.0 * n4;
    this._c5  = -693.0/80.0 * n5;

    // Latitude limit of the Web Mercator projection
    this._LIMIT_PHI = 1.484419982;
  },

  onAdd: function (map) {
    this._map = map;

    let pane                   = map.getPane(this.options.pane);
    if (!pane) {
      pane = this._pane        = map.createPane('edgescalePane', map.getPane('norotatePane') || map.getPane('mapPane'));
      pane.style.zIndex        = 625; // This pane is above markers but below popups.
      pane.style.pointerEvents = 'none';
    }

    this._pane = pane;

    // if (this._renderer) this._renderer.remove()
    // this._renderer               = L.canvas({ pane: "edgescalePane" }).addTo(this._map); // default leaflet svg renderer

    if (!this._canvas) {
      this._initCanvas();
    }

    this._pane.appendChild(this._canvas);

    map.on('viewreset', this._reset, this);
    map.on('move', this._reset, this);
    map.on('moveend', this._reset, this);
    map.on('rotate', this._reset, this);

    this._reset();
  },

  onRemove: function (map) {
    this._pane.removeChild(this._canvas);

    map.off('viewreset', this._reset, this);
    map.off('move', this._reset, this);
    map.off('moveend', this._reset, this);
  },

  addTo: function (map) {
    map.addLayer(this);
    return this;
  },

  setOpacity: function (opacity) {
    this.options.opacity = opacity;
    L.DomUtil.setOpacity(this._canvas, this.options.opacity);
    return this;
  },

  bringToFront: function () {
    if (this._canvas) {
        this._pane.appendChild(this._canvas);
    }
    return this;
  },

  bringToBack: function () {
    if (this._canvas) {
      this._pane.insertBefore(this._canvas, pane.firstChild);
    }
    return this;
  },

  _initCanvas: function () {
    this._canvas = L.DomUtil.create('canvas', '');
    this._ctx = this._canvas.getContext('2d');

    this.setOpacity();

    L.extend(this._canvas, {
      onselectstart: L.Util.falseFn,
      onmousemove: L.Util.falseFn,
      onload: L.bind(this._onCanvasLoad, this)
    });
  },

  _reset: function () {
    var canvas = this._canvas,
        size = this._map.getSize();

    this._setCanvasPosition();

    canvas.width  = size.x;
    canvas.height = size.y;
    canvas.style.width  = size.x + 'px';
    canvas.style.height = size.y + 'px';

    /**
     * @TODO add support for "leaflet-rotate"
     */
    if (this._map._bearing) {
      return;
    }

    const { gradient } = this.options;

    // horizontal gradient
    if (!this._hor_gradient) {
      this._hor_gradient = this._ctx.createLinearGradient(0, 0, 0, gradient.size);
      this._hor_gradient.addColorStop(0,"rgba(255, 255, 255, " + gradient.opacity + ")");
      this._hor_gradient.addColorStop(1,"rgba(255, 255, 255, 0)");
    }
    this._ctx.fillStyle = this._hor_gradient;
    this._ctx.fillRect(0, 0, size.x, gradient.size);

    // vertical gradient
    if (!this._vert_gradient) {
      this._vert_gradient = this._ctx.createLinearGradient(0, 0, gradient.size, 0);
      this._vert_gradient.addColorStop(0,"rgba(255, 255, 255, " + gradient.opacity + ")");
      this._vert_gradient.addColorStop(1,"rgba(255, 255, 255, 0)");
    }
    this._ctx.fillStyle = this._vert_gradient;
    this._ctx.fillRect(0, 0, gradient.size, size.y);

    this._ctx.beginPath();
    this._ctx.moveTo(0,0);
    this._ctx.lineTo(size.x,0);
    this._ctx.lineTo(size.x,size.y);
    this._ctx.stroke();

    this._calcInterval();
    this._draw();
  },

  _onCanvasLoad: function () {
    this.fire('load');
  },

  _calcInterval: function() {
    const { zoomInterval } = this.options;
    const zoom = this._map.getZoom();

    if (undefined !== zoomInterval) {
      // Manually set scale using a custom this.options.zoomInterval object
      for (const idx in zoomInterval) {
        const dict = zoomInterval[idx];
        if (dict.start <= zoom && dict.end && dict.end >= zoom) {
          this._interval = dict.interval;
          break;
        }
      }
    } else {
      // Autamatically get current scale using L.Control.Scale
      // Source: https://gis.stackexchange.com/a/198444
      this._interval = L.Control.Scale.prototype._getRoundNum(
        this._map
          .containerPointToLatLng([0, this._map.getSize().y / 2 ])
          .distanceTo(
            this._map.containerPointToLatLng([L.Control.Scale.prototype.options.maxWidth, this._map.getSize().y / 2 ]
          )
        )
      );
    }

    this._currZoom = zoom;

  },

  _draw: function() {
    this._ctx.strokeStyle = this.options.color;    
    this._create_lat_ticks();
    this._create_lon_ticks();

    this._ctx.fillStyle = this.options.color;
    this._ctx.font      = this.options.font;

    const size = this._map.getSize();
    const text = this._interval >= 1000 ? (this._interval / 1000 + ' km') : this._interval + ' m';    

    this._ctx.textAlign    = 'left';
    this._ctx.textBaseline = 'middle'; 
    this._ctx.fillText(text, +12, size.y / 2);

    this._ctx.textAlign    = 'center';
    this._ctx.textBaseline = 'top'; 
    this._ctx.fillText(text, size.x / 2, 12);
  },

  _create_lat_ticks: function() {
    const { weight } = this.options;
    const size       = this._map.getSize();
    const to_rad     = Math.PI/180.0;
    const center     = this._merLength(this._map.containerPointToLatLng(L.point(0, size.y / 2)).lat * to_rad);
    const top        = this._merLength(this._map.containerPointToLatLng(L.point(0,0)).lat * to_rad);
    const bottom     = this._merLength(this._map.containerPointToLatLng(L.point(0, size.y)).lat * to_rad);
        
    // draw major ticks    
    for (let i = center + this._interval / 2; i < top; i = i + this._interval) {
      const phi = this._invmerLength(i);			
      if ((phi < this._LIMIT_PHI) && (phi > -this._LIMIT_PHI)) {
        this._draw_lat_tick(phi, 10, weight * 1.5);
      }
    }
    for (let i = center - this._interval / 2; i > bottom; i = i - this._interval) {
      const phi = this._invmerLength(i);			
      if ((phi > -this._LIMIT_PHI) && (phi < this._LIMIT_PHI)) {
        this._draw_lat_tick(phi, 10, weight * 1.5);
      }
    }

    // draw minor ticks    
    for (let i = center; i < top; i = i + this._interval / 10.0) {
      const phi = this._invmerLength(i);			
      if ((phi < this._LIMIT_PHI) && (phi > -this._LIMIT_PHI)) {
        this._draw_lat_tick(phi, 4, weight);
      }
    }
    for (let i = center - this._interval / 10; i > bottom; i = i - this._interval / 10.0) {
      const phi = this._invmerLength(i);			
      if ((phi > -this._LIMIT_PHI) && (phi < this._LIMIT_PHI)) {
        this._draw_lat_tick(phi, 4, weight);
      }
    }
  },

  _create_lon_ticks: function() {
    const { weight } = this.options;
    const size       = this._map.getSize();
    const to_rad     = Math.PI/180.0;
    const to_deg     = 180.0/Math.PI;
    const center     = this._map.containerPointToLatLng(L.point(size.x / 2, 0));
    const left       = this._map.containerPointToLatLng(L.point(0, 0));
    const right      = this._map.containerPointToLatLng(L.point(size.x, 0));
    const sinPhi2    = Math.pow(Math.sin(center.lat * to_rad), 2);
    const N          = this._a / Math.sqrt(1.0 - this._e2 * sinPhi2);
    const dl         = this._interval / (N * Math.cos(center.lat * to_rad)) * to_deg;
        
    // draw major ticks
    for (let i = center.lng + dl / 2; i < right.lng; i = i + dl) this._draw_lon_tick(i, 10, weight * 1.5);
    for (let i = center.lng - dl / 2; i > left.lng;  i = i - dl) this._draw_lon_tick(i, 10, weight * 1.5);
    
    // draw minor ticks
    for (let i = center.lng;           i < right.lng; i = i + dl / 10) this._draw_lon_tick(i, 4, weight);
    for (let i = center.lng - dl / 10; i > left.lng;  i = i - dl / 10) this._draw_lon_tick(i, 4, weight);
  },

  _setCanvasPosition: function() {
    let lt = this._map.containerPointToLayerPoint([0, 0]);
    /**
     * @TODO add support for "leaflet-rotate"
     */
    if (this._map._bearing) {
      lt = this._map.rotatedPointToMapPanePoint(
        this._map.containerPointToLayerPoint(L.point(this._map._container.getBoundingClientRect()))
      );
    }
    L.DomUtil.setPosition(this._canvas, lt);
  },

  _latLngToCanvasPoint: function (latlng) {
    return L.point(
      this._map
        .project(L.latLng(latlng))
        ._subtract(this._map.getPixelOrigin())
      ).add(this._map._getMapPanePos()); 
  },

  _draw_lat_tick: function (phi, lenght, weight) {
    const to_deg = 180.0/Math.PI;
    const size   = this._map.getSize();
    const y      = this._latLngToCanvasPoint(L.latLng(phi * to_deg, 0.0)).y;
    this._ctx.lineWidth = weight;
    this._ctx.beginPath();
    this._ctx.moveTo(0, y);
    this._ctx.lineTo(+ lenght, y);
    this._ctx.stroke();
  },

  _draw_lon_tick: function(lam, lenght, weight) {
    const x = this._latLngToCanvasPoint(L.latLng(0.0, lam)).x;
    this._ctx.lineWidth = weight;
    this._ctx.beginPath();
    this._ctx.moveTo(x, 0);
    this._ctx.lineTo(x, lenght);
    this._ctx.stroke();
  },

  _merLength: function(phi) {
    const cos2 = Math.cos(2.0 * phi);
    const sin2 = Math.sin(2.0 * phi); 
    return this._A * (phi + sin2 * (this._c1 + (this._c2 + (this._c3 + (this._c4 + this._c5 * cos2) * cos2) *cos2) * cos2));
  },

  _invmerLength: function(s) {
    const psi = s/this._A;
    const cos2 = Math.cos(2.0 * psi);
    const sin2 = Math.sin(2.0 * psi); 
    return psi + sin2 * (this._ic1 + (this._ic2 + (this._ic3 + (this._ic4 + this._ic5 * cos2) * cos2) * cos2) * cos2);
  },
});

L.control.edgeScale = function (options) {
  return new L.Control.EdgeScale(options);
};

L.Map.mergeOptions({
  edgeScaleControl: false
});

L.Map.addInitHook(function () {
  if (this.options.edgeScaleControl) {
    this.edgeScaleControl = new L.Control.EdgeScale();
    this.addControl(this.edgeScaleControl);
  }
});