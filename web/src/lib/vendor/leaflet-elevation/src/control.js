import * as _      from './utils';
import { Options } from './options';

// "leaflet-i18n" fallback
if (!L._ || !L.i18n) {
	L._ = L.i18n = (string, data) => string;
}

export const Elevation = L.Control.Elevation = L.Control.extend({

	includes: L.Evented ? L.Evented.prototype : L.Mixin.Events,

	options: Options,
	__mileFactor:     0.621371, // 1 km = (0.621371 mi)
	__footFactor:     3.28084,  // 1 m  = (3.28084 ft)
	__D3:            'https://unpkg.com/d3@7.8.4/dist/d3.min.js',
	__TOGEOJSON:     'https://unpkg.com/@tmcw/togeojson@5.6.2/dist/togeojson.umd.js',
	__LGEOMUTIL:     'https://unpkg.com/leaflet-geometryutil@0.10.1/src/leaflet.geometryutil.js',
	__LALMOSTOVER:   'https://unpkg.com/leaflet-almostover@1.0.1/src/leaflet.almostover.js',
	__LHOTLINE:      '../libs/leaflet-hotline.min.js',
	__LDISTANCEM:    '../libs/leaflet-distance-marker.min.js',
	__LEDGESCALE:    '../libs/leaflet-edgescale.min.js',
	__LCHART:        '../src/components/chart.js',
	__LMARKER:       '../src/components/marker.js',
	__LSUMMARY:      '../src/components/summary.js',
	__modulesFolder: '../src/handlers/',
	__btnIcon:       '../images/elevation.svg',

	/*
	 * Add data to the diagram either from GPX or GeoJSON and update the axis domain and data
	 */
	addData(d, layer) {
		this.import(this.__D3)
			.then(() => {
				if (this._modulesLoaded) {
					layer = layer ?? (d.on && d);
					this._addData(d);
					this._addLayer(layer);
					this._fireEvt("eledata_added", { data: d, layer: layer, track_info: this.track_info });
				} else {
					this.once('modules_loaded', () => this.addData(d,layer));
				}
			});
	},

	/**
	 * Adds the control to the given map.
	 */
	addTo(map) {
		if (this.options.detached) {
			let parent = _.select(this.options.elevationDiv);
			let eleDiv = this.onAdd(map);
			parent ? _.append(parent, eleDiv) : _.insert(map.getContainer(), eleDiv, 'afterend');
		} else {
			L.Control.prototype.addTo.call(this, map);
		}
		return this;
	},

	/*
	 * Reset data and display
	 */
	clear() {
		if (this._marker)        this._marker.remove();
		if (this._chart)         this._clearChart();
		if (this._layers)        this._clearLayers(this._layers);
		if (this._markers)       this._clearLayers(this._markers);
		if (this._circleMarkers) this._circleMarkers.remove();
		if (this._hotline)       this._hotline.eachLayer(l => l.options.renderer.remove()); // hotfix for: https://github.com/Raruto/leaflet-elevation/issues/233
		if (this._hotline)       this._clearLayers(this._hotline);

		this._data      = [];
		this.track_info = {};

		this._fireEvt("eledata_clear");

		this._updateChart();
	},

	_clearChart() {
		if (this._events && this._events.elechart_updated) {
			this._events.elechart_updated.forEach(({fn, ctx}) => this.off('elechart_updated', fn, ctx));
		}
		if (this._chart && this._chart._container) {
			this._chart._container.selectAll('g.point .point').remove();
			this._chart.clear();
		}
	},

	_clearLayers(l) {
		l = l || this._layers;
		if (l && l.eachLayer) {
			l.eachLayer(f => f.remove())
			l.clearLayers();
		}
	},

	/**
	 * TODO: Create a base class to handle custom data attributes (heart rate, cadence, temperature, ...)
	 * 
	 * @link https://leafletjs.com/examples/extending/extending-3-controls.html#handlers
	 */
	// addHandler: function (name, HandlerClass) {
	// 	if (HandlerClass) {
	// 		let handler = this[name] = new HandlerClass(this);
	// 		this.handlers.push(handler);
	// 		if (this.options[name]) {
	// 			handler.enable();
	// 		}
	// 	}
	// 	return this;
	// },

	/**
	 * Disable chart brushing.
	 */
	disableBrush() {
		this._chart._brushEnabled = false;
		this._resetDrag();
	},

	/**
	 * Enable chart brushing.
	 */
	enableBrush() {
		this._chart._brushEnabled = true;
	},

	/**
	 * Disable chart zooming.
	 */
	disableZoom() {
		this._chart._zoomEnabled = false;
		this._chart._resetZoom();
	},

	/**
	 * Enable chart zooming.
	 */
	enableZoom() {
		this._chart._zoomEnabled = true;
	},

	/**
	 * Sets a map view that contains the given geographical bounds.
	 */
	fitBounds(bounds) {
		bounds = bounds || this.getBounds();
		if (this._map && bounds.isValid()) this._map.fitBounds(bounds);
	},

	getBounds(data) {
		return L.latLngBounds((data || this._data).map((d) => d.latlng));
	},

	/**
	 * Get default zoom level (followMarker: true).
	 */
	getZFollow() {
		return this.options.zFollow;
	},

	/**
	 * Hide current elevation chart profile.
	 */
	hide() {
		_.style(this._container, "display", "none");
	},

	/**
	 * Initialize chart control "options" and "container".
	 */
	initialize(opts) {

		// opts = L.setOptions(this, opts);

		// Fixes: https://github.com/Raruto/leaflet-elevation/pull/240
		opts = L.setOptions(this, L.extend({}, _.cloneDeep(Options), opts)); // "deep copy" nested objects (multiple charts)

		this._data           = [];
		this._layers         = L.featureGroup();
		this._markers        = L.featureGroup();
		this._hotline        = L.featureGroup();
		this._circleMarkers  = L.featureGroup();
		this._markedSegments = L.polyline([]);
		this._start          = L.circleMarker([0,0], (opts.trkStart || Options.trkStart));
		this._end            = L.circleMarker([0,0], (opts.trkEnd || Options.trkEnd));
		this._chartEnabled   = true;
		this._yCoordMax      = -Infinity;
		this.track_info      = {};
		//  this.handlers        = [];

		if (opts.followMarker)             this._setMapView = _.throttle(this._setMapView, 300, this);
		if (opts.legend)                   opts.margins.bottom += 30;
		if (opts.theme)                    opts.polylineSegments.className += ' ' + opts.theme;
		if (opts.wptIcons === true)        opts.wptIcons = Options.wptIcons;
		if (opts.distanceMarkers === true) opts.distanceMarkers = Options.distanceMarkers;
		if (opts.trkStart)                 this._start.addTo(this._circleMarkers);
		if (opts.trkEnd)                   this._end.addTo(this._circleMarkers);


		this._markedSegments.setStyle(opts.polylineSegments);

		// Leaflet canvas renderer colors
		L.extend(_.Colors, opts.colors || {});

		// Various stuff
		this._fixCanvasPaths();
		this._fixTooltipSize();

	},

	/**
	 * Javascript scripts downloader (lazy loader)
	 */
	import(src, condition) {
		if (Array.isArray(src)) {
			return Promise.all(src.map(m => this.import(m)));
		}
		switch(src) {
			case this.__D3:          condition = typeof d3 !== 'object'; break;
			case this.__TOGEOJSON:   condition = typeof toGeoJSON !== 'object'; break;
			case this.__LGEOMUTIL:   condition = typeof L.GeometryUtil !== 'object'; break;
			case this.__LALMOSTOVER: condition = typeof L.Handler.AlmostOver  !== 'function'; break;
			case this.__LDISTANCEM:  condition = typeof L.DistanceMarkers  !== 'function'; break;
			case this.__LEDGESCALE:  condition = typeof L.Control.EdgeScale !== 'function'; break;
			case this.__LHOTLINE:    condition = typeof L.Hotline  !== 'function'; break;
		}
		return condition !== false ? import(_.resolveURL(src, this.options.srcFolder)) : Promise.resolve();
	},

	/**
	 * Load elevation data (GPX, GeoJSON, KML or TCX).
	 */
	load(data) {
		this._parseFromString(data).then( geojson => geojson ? this._loadLayer(geojson) : this._loadFile(data));
	},

	/**
	 * Create container DOM element and related event listeners.
	 * Called on control.addTo(map).
	 */
	onAdd(map) {
		this._map = map;

		let container = this._container = _.create("div", "elevation-control " + this.options.theme + " " + (this.options.detached ? 'elevation-detached' : 'leaflet-control'), this.options.detached ? { id: 'elevation-' + _.randomId() } : {});

		if (!this.eleDiv) this.eleDiv = container;

		this._loadModules(this.options.handlers).then(() => { // Inject here required modules (data handlers)
			this._initChart(container);
			this._initButton(container);
			this._initSummary(container);
			this._initMarker(map);
			this._initLayer(map);
			this._modulesLoaded = true;
			this.fire('modules_loaded');
		});

		this.fire('add');

		return container;
	},

	/**
	 * Clean up control code and related event listeners.
	 * Called on control.remove().
	 */
	onRemove(map) {
		this._container = null;

		map
			.off('zoom viewreset zoomanim',      this._hideMarker,    this)
			.off('resize',                       this._resetView,     this)
			.off('resize',                       this._resizeChart,   this)
			.off('mousedown',                    this._resetDrag,     this);

		_.off(map.getContainer(), 'mousewheel',  this._resetDrag,     this);
		_.off(map.getContainer(), 'touchstart',  this._resetDrag,     this);
		_.off(document,           'keydown',     this._onKeyDown,     this);

		this
			.off('eledata_added eledata_loaded', this._updateChart,   this)
			.off('eledata_added eledata_loaded', this._updateSummary, this);

		this.fire('remove');
	},

	/**
	 * Redraws the chart control. Sometimes useful after screen resize.
	 */
	redraw() {
		this._resizeChart();
	},

	/**
	 * Set default zoom level (followMarker: true).
	 */
	setZFollow(zoom) {
		this.options.zFollow = zoom;
	},

	/**
	 * Hide current elevation chart profile.
	 */
	show() {
		_.style(this._container, "display", "block");
	},

	/*
	 * Parsing data either from GPX or GeoJSON and update the diagram data
	 */
	_addData(d) {
		if (!d) {
			return;
		}

		// Standard GeoJSON
		if (d.type === "FeatureCollection" ) {
			return _.each(d.features, feature => this._addData(feature));
		} else if (d.type === "Feature") {
			let geom = d.geometry;
			if (geom) {
				switch (geom.type) {
					case 'LineString':		return this._addGeoJSONData(geom.coordinates, d.properties);
					case 'MultiLineString':	return _.each(geom.coordinates, (coords, i) => this._addGeoJSONData(coords, d.properties, i));
					case 'Point':
					default:				return console.warn('Unsopperted GeoJSON feature geometry type:' + geom.type);
				}
			}
		}

		// Fallback for leaflet layers (eg. L.Gpx)
		if (d._latlngs) {
			return this._addGeoJSONData(d._latlngs, d.feature && d.feature.properties);
		}

	},

	/*
	 * Parsing of GeoJSON data lines and their elevation in z-coordinate
	 */
	_addGeoJSONData(coords, properties, nestingLevel) {

		// "coordinateProperties" property is generated inside "@tmcw/toGeoJSON"
		let props  = (properties && properties.coordinateProperties) || properties;

		coords.forEach((point, i) => {

			// GARMIN_EXTENSIONS = ["hr", "cad", "atemp", "wtemp", "depth", "course", "bearing"];
			point.meta = point.meta ?? { time: null, ele: null };
			
			point.prev = (attr) => (attr ? this._data[i > 0 ? i - 1 : 0][attr] : this._data[i > 0 ? i - 1 : 0]);

			this.fire("elepoint_init", { point: point, props: props, id: i, isMulti: nestingLevel });

			this._addPoint(
				point.lat ?? point[1], 
				point.lng ?? point[0],
				point.alt ?? point.meta.ele ?? point[2]
			);

			this.fire("elepoint_added", { point: point, index: this._data.length - 1 });

			if (this._yCoordMax < this._data[this._data.length - 1][this.options.yAttr]) this._yCoordMax = this._data[this._data.length - 1][this.options.yAttr];
		});

		this.fire("eletrack_added", { coords: coords, index: this._data.length - 1 });
	},

	/*
	 * Parse and push a single (x, y, z) point to current elevation profile.
	 */
	_addPoint(x, y, z) {
		if (this.options.reverseCoords) {
			[x, y] = [y, x];
		}

		this._data.push({
			x: x,
			y: y,
			z: z,
			latlng: L.latLng(x, y, z)
		});

		this.fire("eledata_updated", { index: this._data.length - 1 });
	},

	_addLayer(layer) {
		if (layer) this._layers.addLayer(layer)
		// Postpone adding the distance markers (lazy: true)
		if (layer && this.options.distanceMarkers && this.options.distanceMarkers.lazy) {
			layer.on('add remove', ({target, type}) => L.DistanceMarkers && target instanceof L.Polyline && target[type + 'DistanceMarkers']());
		}
		return layer;
	},

	_addMarker(marker) {
		if (marker) this._markers.addLayer(marker)
		return marker;
	},

	/**
	 * Initialize "L.AlmostOver" integration
	 */
	_initAlmostOverHandler(map, layer) {
		return (map && this.options.almostOver && !L.Browser.mobile) ? this.import([this.__LGEOMUTIL, this.__LALMOSTOVER])
			.then(() => {
				map.addHandler('almostOver', L.Handler.AlmostOver)
				if (L.GeometryUtil && map.almostOver && map.almostOver.enabled()) {
					map.almostOver.addLayer(layer);
					map
						.on('almost:move', this._onMouseMoveLayer, this)
						.on('almost:out', this._onMouseOut, this);
					this.once('eledata_clear', () => {
						map.almostOver.removeLayer(layer);
						map
						.off('almost:move', this._onMouseMoveLayer, this)
						.off('almost:out', this._onMouseOut, this);
					})
				}
			}) : Promise.resolve();
	},

	/**
	 * Initialize "L.DistanceMarkers" integration
	 */
	_initDistanceMarkers() {
		return this.options.distanceMarkers ? this.import([this.__LGEOMUTIL, this.__LDISTANCEM]) : Promise.resolve();
	},

	/**
	 * Initialize "L.Control.EdgeScale" integration
	 */
	_initEdgeScale(map) {
		return this.options.edgeScale ? this.import(this.__LEDGESCALE)
			.then(() => {
				map.edgeScaleControl = map.edgeScaleControl || L.control.edgeScale('boolean' !== typeof this.options.edgeScale ? this.options.edgeScale : {}).addTo(map);
			}) : Promise.resolve();
	},

	_initHotLine(layer) {
		let prop = typeof this.options.hotline == 'string' ? this.options.hotline : 'elevation';
		return this.options.hotline ? this.import(this.__LHOTLINE)
			.then(() => {
				layer.eachLayer((trkseg) => {
					if (trkseg.feature.geometry.type != "Point") {
						let geo = L.geoJson(trkseg.toGeoJSON(), { coordsToLatLng: (coords) => L.latLng(coords[0], coords[1], coords[2] * (this.options.altitudeFactor || 1))});
						let line = L.hotline(geo.toGeoJSON().features[0].geometry.coordinates, {
							renderer: L.Hotline.renderer(),
							min: isFinite(this.track_info[prop + '_min']) ? this.track_info[prop + '_min'] : 0,
							max: isFinite(this.track_info[prop + '_max']) ? this.track_info[prop + '_max'] : 1,
							palette: {
								0.0: '#008800',
								0.5: '#ffff00',
								1.0: '#ff0000'
							},
							weight: 5,
							outlineColor: '#000000',
							outlineWidth: 1
						}).addTo(this._hotline);
						let alpha = trkseg.options.style && trkseg.options.style.opacity || 1;
						trkseg.on('add remove', ({type}) => {
							trkseg.setStyle({opacity: (type == 'add' ? 0 : alpha)});
							line[(type == 'add' ? 'addTo' : 'removeFrom')](trkseg._map);
							if (line._renderer) line._renderer._container.parentElement.insertBefore(line._renderer._container, line._renderer._container.parentElement.firstChild);
						});
					}
				});
			}) : Promise.resolve();
	},

	/**
	 * Initialize "L.AlmostOver" and "L.DistanceMarkers"
	 */
	_initMapIntegrations(layer) {
		let map = this._map;
		if (map) {
			if (this._data.length) {
				this._start.setLatLng(this._data[0].latlng);
				this._end.setLatLng(this._data[this._data.length -1].latlng);
			}
			Promise.all([
				this._initHotLine(layer),
				this._initAlmostOverHandler(map, layer),
				this._initDistanceMarkers(),
				this._initEdgeScale(map),
			]).then(() => {
				if (this.options.polyline) {
					this._layers.addLayer(layer.addTo(map)); // hotfix for: https://github.com/Raruto/leaflet-elevation/issues/233
					this._circleMarkers.addTo(map);
				}
				if (this.options.autofitBounds) {
					this.fitBounds(layer.getBounds());
				}
				map.invalidateSize();
			});
		} else {
			this.once('add', () => this._initMapIntegrations(layer));
		}
	},

	/*
	 * Collapse current chart control.
	 */
	_collapse() {
		_.replaceClass(this._container, 'elevation-expanded', 'elevation-collapsed');
		if (this._map) this._map.invalidateSize();
	},

	/*
	 * Expand current chart control.
	 */
	_expand() {
		_.replaceClass(this._container, 'elevation-collapsed', 'elevation-expanded');
		if (this._map) this._map.invalidateSize();
	},

	/**
	 * Add some basic colors to leaflet canvas renderer (preferCanvas: true).
	 */
	_fixCanvasPaths() {
		let oldProto = L.Canvas.prototype._fillStroke;
		let control  = this;

		let theme      = this.options.theme.split(' ')[0].replace('-theme', '');
		let color      = _.Colors[theme] || {};

		L.Canvas.include({
			_fillStroke(ctx, layer) {
				if (control._layers.hasLayer(layer)) {

					let options    = layer.options;

					options.color  = color.line || color.area || theme;
					options.stroke = !!options.color;

					oldProto.call(this, ctx, layer);

					if (options.stroke && options.weight !== 0) {
						let oldVal                   = ctx.globalCompositeOperation || 'source-over';
						ctx.globalCompositeOperation = 'destination-over'
						ctx.strokeStyle              = color.outline || '#FFF';
						ctx.lineWidth                = options.weight * 1.75;
						ctx.stroke();
						ctx.globalCompositeOperation = oldVal;
					}

				} else {
					oldProto.call(this, ctx, layer);
				}
			}
		});
	},

	/**
	 * Partial fix for initial tooltip size
	 * 
	 * @link https://github.com/Raruto/leaflet-elevation/issues/81#issuecomment-713477050
	 */
	_fixTooltipSize() {
		this.on('elechart_init', () =>
			this.once('elechart_change elechart_hover', ({data, xCoord}) => {
				if (this._chartEnabled) {
					this._chart._showDiagramIndicator(data, xCoord);
					this._chart._showDiagramIndicator(data, xCoord);
				}
				this._updateMarker(data);
			})
		);
	},

	/*
	 * Finds a data entry for the given LatLng
	 */
	_findItemForLatLng(latlng) {
		return this._data[this._chart._findIndexForLatLng(latlng)];
	},

	/*
	 * Finds a data entry for the given xDiagCoord
	 */
	_findItemForX(x) {
		return this._data[this._chart._findIndexForXCoord(x)];
	},

	/**
	 * Fires an event of the specified type.
	 */
	_fireEvt(type, data, propagate) {
		if (this.fire) this.fire(type, data, propagate);
		if (this._map) this._map.fire(type, data, propagate);
	},

	/*
	 * Hides the position/height indicator marker drawn onto the map
	 */
	_hideMarker() {
		if (this.options.autohideMarker) {
			this._marker.remove();
		}
	},

	/**
	 * Generate "svg" chart (DOM element).
	 */
	_initChart(container) {
		let opts = this.options;
		let map  = this._map;

		if (opts.detached) {
			let { offsetWidth, offsetHeight}             = this.eleDiv;
			if (offsetWidth > 0)             opts.width  = offsetWidth;
			if (offsetHeight > 20)           opts.height = offsetHeight - 20; // 20 = horizontal scrollbar size.
		} else {
			let { clientWidth }                          = map.getContainer();
			opts._maxWidth                               = opts._maxWidth > opts.width ? opts._maxWidth : opts.width;
			this._container.style.maxWidth               = opts._maxWidth + 'px';
			if (opts._maxWidth > clientWidth) opts.width = clientWidth - 30;
		}

		this
			.import([this.__D3, this.__LCHART])
			.then((m) => {

			let chart = this._chart = new (m[1] || Elevation).Chart(opts, this);
	
			this._x     = this._chart._x;
			this._y     = this._chart._y;
	
			d3
				.select(container)
				.call(chart.render())
	
			chart
				.on('reset_drag',      this._hideMarker,     this)
				.on('mouse_enter',     this._onMouseEnter,   this)
				.on('dragged',         this._onDragEnd,      this)
				.on('mouse_move',      this._onMouseMove,    this)
				.on('mouse_out',       this._onMouseOut,     this)
				.on('ruler_filter',    this._onRulerFilter,  this)
				.on('zoom',            this._updateChart,    this)
				.on('elepath_toggle',  this._onToggleChart,  this)
				.on('margins_updated', this._resizeChart,    this);
	
	
			this.fire("elechart_init");

			map
				.on('zoom viewreset zoomanim',      this._hideMarker,    this)
				.on('resize',                       this._resetView,     this)
				.on('resize',                       this._resizeChart,   this)
				.on('rotate',                       this._rotateMarker,  this)
				.on('mousedown',                    this._resetDrag,     this);

			_.on(map.getContainer(), 'mousewheel',  this._resetDrag,     this);
			_.on(map.getContainer(), 'touchstart',  this._resetDrag,     this);
			_.on(document,           'keydown',     this._onKeyDown,     this);

			this
				.on('eledata_added eledata_loaded', this._updateChart,   this)
				.on('eledata_added eledata_loaded', this._updateSummary, this);

			this._updateChart();
			this._updateSummary();
		});


	},

	_initLayer() {
		this._layers
			.on('layeradd layerremove', ({layer, type}) => {
				let node  = layer.getElement && layer.getElement();
				_.toggleClass(node,  this.options.polyline.className + ' ' + this.options.theme, type == 'layeradd');
				_.toggleEvent(layer, "mousemove", this._onMouseMoveLayer.bind(this),             type == 'layeradd')
				_.toggleEvent(layer, "mouseout",  this._onMouseOut.bind(this),                   type == 'layeradd');
			});
	},

	_initMarker(map) {
		let pane                     = map.getPane('elevationPane');
		if (!pane) {
			pane = this._pane        = map.createPane('elevationPane', map.getPane('norotatePane') || map.getPane('mapPane'));
			pane.style.zIndex        = 625; // This pane is above markers but below popups.
			pane.style.pointerEvents = 'none';
		}

		if (this._renderer) this._renderer.remove()
		this._renderer               = L.svg({ pane: "elevationPane" }).addTo(this._map); // default leaflet svg renderer

		this.import([this.__D3, this.__LMARKER])
			.then((m) => {
				this._marker             = new (m[1] || Elevation).Marker(this.options, this);
				this.fire("elechart_marker");
			});
	},

	/**
	 * Inspired by L.Control.Layers
	 */
	_initButton(container) {
		L.DomEvent
			.disableClickPropagation(container)
			.disableScrollPropagation(container);

		this.options.collapsed ? this._collapse() : this._expand();

		if (this.options.autohide) {
			_.on(container, 'mouseover', this._expand,   this);
			_.on(container, 'mouseout',  this._collapse, this);
			this._map.on('click', this._collapse, this);
		}

		if (this.options.closeBtn) {
			let link = this._button = _.create('a', "elevation-toggle-icon", { href: '#', title: L._('Elevation'), }, container);
			_.on(link, 'click', L.DomEvent.stop);
			_.on(link, 'click', this._toggle, this);
			_.on(link, 'focus', this._toggle, this);
			fetch(_.resolveURL(this.__btnIcon, this.options.srcFolder)).then(r => r.ok && r.text().then(img => link.innerHTML = img));
		}
	},

	_initSummary(container) {
		this.import(this.__LSUMMARY).then((m)=>{
			this._summary = new (m || Elevation).Summary({ summary: this.options.summary }, this);

			this.on('elechart_init', () => {
				d3.select(container).call(this._summary.render());
			});
		});
	},

	/**
	 * Retrieve data from a remote url (HTTP).
	 */
	_loadFile(url) {
		fetch(url)
			.then((response) => response.text())
			.then((data)     => {
				this._downloadURL = url; // TODO: handle multiple urls?
				this._parseFromString(data)
					.then( geojson => geojson && this._loadLayer(geojson));
			}).catch((err) => console.warn(err));
	},

	/**
	 * Dynamically import only required javascript modules (code splitting)
	 */
	_loadModules(handlers) {
		// First map known classnames (eg. "Altitude" --> L.Control.Elevation.Altitude)
		handlers = handlers.map((h) => typeof h === 'string' && typeof Elevation[h] !== "undefined" ? Elevation[h] : h);
		// Then load optional classes and custom imports (eg. "Cadence" --> import('../src/handlers/cadence.js'))
		let modules = handlers.map(file => (typeof file === 'string' && this.import(this.__modulesFolder + file.toLowerCase() + '.js')) || (file instanceof Promise && file) || Promise.resolve());
		return Promise.all(modules).then((m) => {
			_.each(m, (exported, i) => {
				let fn = exported && Object.keys(exported)[0];
				if (fn) {
					handlers[i] = Elevation[fn] = (Elevation[fn] ?? exported[fn]);
				}
			});
			_.each(handlers, h => ["function", "object"].includes(typeof h) && this._registerHandler(h));
		});
	},

	/**
	 * Simple GeoJSON data loader (L.GeoJSON).
	 */
	_loadLayer(geojson) {
		let { polyline, theme, waypoints, wptIcons, wptLabels, distanceMarkers } = this.options;
		let style = L.extend({}, polyline);

		if (theme) {
			style.className += ' ' + theme;
		}

		if (geojson.name) {
			this.track_info.name = geojson.name;
		}

		let layer = L.geoJson(geojson, {
			distanceMarkers: distanceMarkers,
			style: style,
			pointToLayer: (feature, latlng) => {
				if (waypoints) {
					let { desc, name, sym } = feature.properties;
					desc = desc || '';
					name = name || '';
					// Handle chart waypoints (dots)
					if ([true, 'dots'].includes(waypoints)) {
						this._registerCheckPoint({
							latlng: latlng,
							label : ([true, 'dots'].includes(wptLabels) ? name : '')
						});
					}
					// Handle map waypoints (markers)
					if ([true, 'markers'].includes(waypoints) && wptIcons != false) {
						return this._registerMarker({
							latlng : latlng,
							sym    : (sym ?? name).replace(' ', '-').replace('"', '').replace("'", '').toLowerCase(),
							content: [true, 'markers'].includes(wptLabels) && (name || desc) && decodeURI("<b>" + name + "</b>" + (desc.length > 0 ? '<br>' + desc : ''))
						});
					}
				}
			},
			onEachFeature: (feature, layer) => feature.geometry && feature.geometry.type != 'Point' && this.addData(feature, layer),
		});

		this.import(this.__D3).then(() => {
			this._initMapIntegrations(layer);
			const event_data = { data: geojson, layer: layer, name: this.track_info.name, track_info: this.track_info };
			if (this._modulesLoaded) {
				this._fireEvt("eledata_loaded", event_data);
			} else {
				this.once('modules_loaded', () => this._fireEvt("eledata_loaded", event_data));
			}
		});

		return layer;
	},

	_onDragEnd({ dragstart, dragend}) {
		this._hideMarker();
		this.fitBounds(L.latLngBounds([dragstart.latlng, dragend.latlng]));

		this.fire("elechart_dragged");
	},

	_onKeyDown({key}) {
		if (!this.options.detached && key === "Escape"){
			this._collapse()
		};
	},

	/**
	 * Trigger mouseenter event.
	 */
	_onMouseEnter() {
		this.fire('elechart_enter');
	},

	/*
	 * Handles the moueseover the chart and displays distance and altitude level.
	 */
	_onMouseMove({xCoord}) {
		if (this._chartEnabled && this._data.length) {
			let item = this._findItemForX(xCoord);
			if (item) {
				if (this._chartEnabled) this._chart._showDiagramIndicator(item, xCoord);

				this._updateMarker(item);
				this._setMapView(item);

				if (this._map) {
					_.addClass(this._map.getContainer(), 'elechart-hover');
				}

				this.fire("elechart_change", { data: item, xCoord: xCoord });
				this.fire("elechart_hover",  { data: item, xCoord: xCoord });
			}
		}
	},

	/*
	 * Handles mouseover events of the data layers on the map.
	 */
	_onMouseMoveLayer({latlng}) {
		if (this._data.length) {
			let item = this._findItemForLatLng(latlng);
			if (item) {
				let xCoord = item.xDiagCoord;
	
				if (this._chartEnabled) this._chart._showDiagramIndicator(item, xCoord);
	
				this._updateMarker(item);
	
				this.fire("elechart_change", { data: item, xCoord: xCoord });
			}
		}
	},

	/*
	 * Handles the moueseout over the chart.
	 */
	_onMouseOut() {
		if (!this.options.detached) {
			this._hideMarker();
			this._chart._hideDiagramIndicator();
		}

		if (this._map) {
			_.removeClass(this._map.getContainer(), 'elechart-hover');
		}

		this.fire("elechart_leave");
	},

	/**
	 * Handles the drag event over the ruler filter.
	 */
	_onRulerFilter({coords}) {
		this._updateMapSegments(coords);
	},

	/**
	 * Toggle chart data on legend click
	 */
	_onToggleChart({ name, enabled }) {

		this._chartEnabled = this._chart._hasActiveLayers();

		// toggle layer visibility on empty chart
		this._layers.eachLayer(layer => _.toggleClass(layer.getElement && layer.getElement(), this.options.polyline.className + ' ' + this.options.theme, this._chartEnabled));

		// toggle option value (eg. altitude = { 'disabled' || 'enabled' })
		this.options[name] = !enabled && this.options[name] == 'disabled' ? 'enabled' : 'disabled';

		// remove marker on empty chart
		if (!this._chartEnabled) {
			this._chart._hideDiagramIndicator();
			this._marker.remove();
		}
	},

	/**
	 * Simple GeoJSON Parser
	 */
	_parseFromGeoJSONString(data) {
		try {
			return JSON.parse(data);
		} catch (e) { }
	},

	/**
	 * Attempt to parse raw response data (GeoJSON or XML > GeoJSON)
	 */
	_parseFromString(data) {
		return new Promise(resolve =>
			this.import(this.__TOGEOJSON).then(() => {
				let geojson;
				try {
					geojson = this._parseFromXMLString(data.trim());
				} catch (e) {
					geojson = this._parseFromGeoJSONString(data.toString());
				}
				if (geojson) {
					geojson.name = geojson.name || (this._downloadURL || '').split('/').pop().split('#')[0].split('?')[0];
				}
				resolve(geojson);
			})
		);
	},

	/**
	 * Simple XML Parser (GPX, KML, TCX)
	 */
	_parseFromXMLString(data) {
		if (data.indexOf("<") != 0) {
			throw 'Invalid XML';
		}
		let xml  = (new DOMParser()).parseFromString(data, "text/xml");
		let type = xml.documentElement.tagName.toLowerCase(); // "kml" or "gpx"
		let name = xml.getElementsByTagName('name');
		if (xml.getElementsByTagName('parsererror').length) {
			throw 'Invalid XML';
		}
		if (!(type in toGeoJSON)) {
			type = xml.documentElement.tagName == "TrainingCenterDatabase" ? 'tcx' : 'gpx';
		}
		let geojson  = toGeoJSON[type](xml);
		geojson.name = name.length > 0 ? (Array.from(name).find(tag => tag.parentElement.tagName == "trk") ?? name[0]).textContent : '';
		return geojson;
	},

	/**
	 * Add chart profile to diagram
	 */
	_registerAreaPath(props) {
		this.on("elechart_init", () => this._chart._registerAreaPath(props));
	},

	/**
	 * Add chart grid to diagram
	 */
	_registerAxisGrid(props) {
		this.on("elechart_axis", () => this._chart._registerAxisGrid(props));
	},

	/**
	 * Add chart axis to diagram
	 */
	_registerAxisScale(props) {
		this.on("elechart_axis", () => this._chart._registerAxisScale(props));
	},

	/**
	 * Add a point of interest over the diagram
	 */
	_registerCheckPoint(props) {
		const cb = () => this._chart._registerCheckPoint(props);
		this
			.on("elechart_updated", cb)
			.once("eledata_clear", () => this.off("elechart_updated", cb));
	},

	/**
	 * Base handler for iterative track statistics (dist, time, z, slope, speed, acceleration, ...)
	 */
	 _registerDataAttribute(props) {

		// parse of "coordinateProperties" for later usage
		if (props.coordPropsToMeta) {
			this.on("elepoint_init", (e) => props.coordPropsToMeta.call(this, e));
		}

		// prevent excessive variabile instanstations
		let i, curr, prev, attr = props.attr || props.name;

		// save here a reference to last used point
		let lastValid = {};

		// iteration
		this.on("elepoint_added", ({index, point}) => {
			i = index;

			prev = curr ?? this._data[i]; // same as: this._data[i > 0 ? i - 1 : i]
			curr = this._data[i];

			// retrieve point value
			curr[attr] = props.pointToAttr.call(this, point, i);

			// check and fix missing data on last added point
			if (i > 0 && isNaN(prev[attr])) {
				if (!isNaN(lastValid[attr]) && !isNaN(curr[attr])) {
					prev[attr]   = (lastValid[attr] + curr[attr]) / 2;
				} else if (!isNaN(lastValid[attr])) {
					prev[attr]   = lastValid[attr];
				} else if (!isNaN(curr[attr])) {
					prev[attr]   = curr[attr];
				}
				// update "yAttr" and "xAttr"
				if (props.meta) {
					prev[props.meta] = prev[attr];
				}
			}

			// skip to next iteration for invalid or missing data (eg. i == 0)
			if (isNaN(curr[attr])) {
				return;
			}

			// update reference to last used point
			lastValid[attr] = curr[attr];

			// Limit "crazy" delta values.
			if (props.deltaMax) {
				curr[attr] =_.wrapDelta(curr[attr], prev[attr], props.deltaMax);
			}
			
			// Range of acceptable values.
			if (props.clampRange) {
				curr[attr] = _.clamp(curr[attr], props.clampRange);
			}

			// Limit floating point precision.
			if (!isNaN(props.decimals)) {
				curr[attr] = _.round(curr[attr], props.decimals);
			}

			// update "track_info" stats (min, max, avg, ...)
			if (props.stats) {
				for (const key in props.stats) {
					let sname = (props.statsName || attr) + (key != '' ? '_' : '');
					this.track_info[sname + key] = props.stats[key].call(this, curr[attr], this.track_info[sname + key], this._data.length);
				}
			}

			// update here some mixins (eg. complex "track_info" stuff)
			if (props.onPointAdded) props.onPointAdded.call(this, curr[attr], i, point);
		});
	},

	/**
	 * Parse a module definition and attach related function listeners
	 */
	_registerHandler(props) {

		// eg. L.Control.Altitude
		if (typeof props === "function") {
			return this._registerHandler(props.call(this));
		}

		let {
			name,
			attr,
			required,
			deltaMax,
			clampRange,
			decimals,
			meta,
			unit,
			coordinateProperties,
			coordPropsToMeta,
			pointToAttr,
			onPointAdded,
			stats,
			statsName,
			grid,
			scale,
			path,
			tooltip,
			summary
		} = props;

		// eg. "altitude" == true
		if (this.options[name] || required) {

			this._registerDataAttribute({
				name,
				attr,
				meta,
				deltaMax,
				clampRange,
				decimals,
				coordPropsToMeta: _.coordPropsToMeta(coordinateProperties, meta || name, coordPropsToMeta),
				pointToAttr,
				onPointAdded,
				stats,
				statsName,
			});

			if (grid) {
				this._registerAxisGrid(L.extend({ name }, grid));
			}

			if (this.options[name] !== "summary") {
				if (scale) this._registerAxisScale(L.extend({ name, label: unit }, scale));
				if (path)  this._registerAreaPath(L.extend({ name }, path));
			}

			if (tooltip || props.tooltips) {
				_.each([tooltip, ...(props.tooltips || [])], t => t && this._registerTooltip(L.extend({ name }, t)));
			}

			if (summary) {
				_.each(summary, (s, k) => summary[k] = L.extend({ unit }, s));
				this._registerSummary(summary);
			}
		}
	},

	_registerMarker({latlng, sym, content}) {
		let { wptIcons } = this.options;
		// generate and cache appropriate icon symbol
		if (!wptIcons.hasOwnProperty(sym)) {
			wptIcons[sym] = L.divIcon(L.extend({}, wptIcons[""].options, { html: '<i class="elevation-waypoint-icon ' + sym + '"></i>' } ));
		}
		let marker = L.marker(latlng, { icon: wptIcons[sym] });
		if (content) {
			marker.bindPopup(content, { className: 'elevation-popup', keepInView: true }).openPopup();
			marker.bindTooltip(content, { className: 'elevation-tooltip', direction: 'auto', sticky: true, opacity: 1 }).openTooltip();
		}
		return this._addMarker(marker)
	},
	
	/**
	 * Add chart or marker tooltip info
	 */
	_registerTooltip(props) {
		props.chart  && this.on("elechart_init",   () => this._chart._registerTooltip(L.extend({}, props, { value: props.chart })));
		props.marker && this.on("elechart_marker", () => this._marker._registerTooltip(L.extend({}, props, { value: props.marker })));
	},

	/**
	 * Add summary info to diagram
	 */
	_registerSummary(props) {
		this.on('elechart_summary',  () => this._summary._registerSummary(props));
	},

	/*
	 * Removes the drag rectangle and zoms back to the total extent of the data.
	 */
	_resetDrag() {
		this._chart._resetDrag();
		this._hideMarker();
	},

	/**
	 * Resets drag, marker and bounds.
	 */
	_resetView() {
		if (this._map && this._map._isFullscreen) return;
		this._resetDrag();
		this._hideMarker();
		if (this.options.autofitBounds) {
			this.fitBounds();
		}
	},

	/**
	 * Hacky way for handling chart resize. Deletes it and redraw chart.
	 */
	_resizeChart() {
		if (this._container && _.style(this._container, "display") != "none") {
			let opts = this.options;
			let newWidth = opts.detached ? (this.eleDiv || this._container).offsetWidth : _.clamp(opts._maxWidth, [0, this._map.getContainer().clientWidth - 30]);
			if (newWidth) {
				opts.width = newWidth;
				if (this._chart && this._chart._chart) {
					this._chart._chart._resize(opts);
					this._updateChart();
				}
			}
			this._updateMapSegments();
		}
	},

	/**
	 * Collapse or Expand chart control.
	 */
	_toggle() {
		_.hasClass(this._container, "elevation-expanded") ? this._collapse() : this._expand();
	},

	/**
	 * Update map center and zoom (followMarker: true)
	 */
	_setMapView(item) {
		if (this._map && this.options.followMarker) {
			let zoom = this._map.getZoom();
			let z    = this.options.zFollow;
			if (typeof z === "number") {
				this._map.setView(item.latlng, (zoom < z ? z : zoom), { animate: true, duration: 0.25 });
			} else if (!this._map.getBounds().contains(item.latlng)) {
				this._map.setView(item.latlng, zoom, { animate: true, duration: 0.25 });
			}
		}
	},

	/**
	 * Calculates [x, y] domain and then update chart.
	 */
	_updateChart() {
		if (this._chart && this._container) {
			this.fire("elechart_axis");
	
			this._chart.update({ data: this._data, options: this.options });
	
			this._x     = this._chart._x;
			this._y     = this._chart._y;
	
			this.fire('elechart_updated');		
		}
	},

	/*
	 * Update the position/height indicator marker drawn onto the map
	 */
	_updateMarker(item) {
		if (this._marker) {
			this._marker.update({
				map         : this._map,
				item        : item,
				yCoordMax   : this._yCoordMax || 0,
				options     : this.options
			});
		}
	},

	/**
	 * Fix marker rotation on rotated maps
	 */
	_rotateMarker() {
		if (this._marker) {
			this._marker.update();
		}
	},

	/**
	 * Highlight track segments on the map.
	 */
	_updateMapSegments(coords) {
		this._markedSegments.setLatLngs(coords || []);
		if (coords && this._map && !this._map.hasLayer(this._markedSegments)) {
			this._markedSegments.addTo(this._map);
		}
	},

	/**
	 * Update chart summary.
	 */
	_updateSummary() {
		if (this._summary) {
			this._summary.reset();
			if (this.options.summary) {
				this.fire("elechart_summary");
				this._summary.update();
			}
			if (this.options.downloadLink && this._downloadURL) { // TODO: generate dynamically file content instead of using static file urls.
				this._summary._container.innerHTML += '<span class="download"><a href="#">' + L._('Download') + '</a></span>'
				_.select('.download a', this._summary._container).onclick = (e) => {
					e.preventDefault();
					let event = { downloadLink: this.options.downloadLink, confirm: _.saveFile.bind(this, this._downloadURL) };
					if (this.options.downloadLink == 'modal' && typeof CustomEvent === "function") {
						document.dispatchEvent(new CustomEvent("eletrack_download", { detail: event }));
					} else if (this.options.downloadLink == 'link' || this.options.downloadLink === true) {
						event.confirm();
					}
					this.fire('eletrack_download', event);
				};
			}
		}
	},


	/**
	 * Calculates chart width.
	 */
	_width() {
		if (this._chart) return this._chart._width();
		const { width, margins } = this.options;
		return width - margins.left - margins.right;
	},

	/**
	 * Calculates chart height.
	 */
	_height() {
		if (this._chart) return this._chart._height();
		const { height, margins } = this.options;
		return height - margins.top - margins.bottom;
	},

});
