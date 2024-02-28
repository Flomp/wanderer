import * as D3 from './d3.js';

const _ = L.Control.Elevation.Utils;

export var Marker = L.Class.extend({

	initialize(options, control) {
		this.options = options;
		this.control = control

		switch(this.options.marker) {
			case 'elevation-line':
				// this._container = d3.create("g").attr("class", "height-focus-group");
			break;
			case 'position-marker':
				// this._marker   = L.circleMarker([0, 0], { pane: 'overlayPane', radius: 6, fillColor: '#fff', fillOpacity:1, color: '#000', weight:1, interactive: false });
				this._marker      = L.marker([0, 0], { icon: this.options.markerIcon, zIndexOffset: 1000000, interactive: false });
			break;
		}

		this._labels = {};

		return this;
	},

	addTo(map) {
		this._map = map;
		switch(this.options.marker) {
			case 'elevation-line':  this._container = d3.select(map.getPane('elevationPane')).select("svg > g").call(D3.PositionMarker({})); break;
			case 'position-marker': this._marker.addTo(map, { pane: 'overlayPane' }); break;
		}
		return this;
	},

	/**
	 * Update position marker ("leaflet-marker").
	 */
	update(props) {
		if (props) this._props = props;
		else props = this._props;

		if (!props) return;

		if (props.options) this.options = props.options;
		if (!this._map) this.addTo(props.map);

		this._latlng = props.item.latlng;

		switch(this.options.marker) {
			case 'elevation-line':
				if (this._container) {
					let point = this._map.latLngToLayerPoint(this._latlng);
					point     = L.extend({}, props.item, this._map._rotate ? this._map.rotatedPointToMapPanePoint(point) : point);

					let yMax = (this.control._height() / props.yCoordMax * point[this.options.yAttr]);

					if (!isFinite(yMax) || isNaN(yMax)) yMax = 0;

					this._container.classed("leaflet-hidden", false);
					this._container.call(D3.PositionMarker({
						theme : this.options.theme,
						xCoord: point.x,
						yCoord: point.y,
						length: point.y - yMax, // normalized Y
						labels: this._labels,
						item: point,
					}));
				}
			break;
			case 'position-marker':
				_.removeClass(this._marker.getElement(), 'leaflet-hidden');
				this._marker.setLatLng(this._latlng);
			break;
		}
	},

	/*
	 * Hides the position/height indicator marker drawn onto the map
	 */
	remove() {
		this._props = null;
		switch(this.options.marker) {
			case 'elevation-line':  this._container && this._container.classed("leaflet-hidden", true); break;
			case 'position-marker': _.addClass(this._marker.getElement(), 'leaflet-hidden');            break;
		}
	},

	getLatLng() {
		return this._latlng;
	},

	_registerTooltip(props) {
		this._labels[props.name] = props;
	}

});
