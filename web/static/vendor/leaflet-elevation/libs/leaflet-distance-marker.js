/*
 * Copyright (c) 2022, GPL-3.0+ Project, Raruto
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
 *     Copyright (c) 2014- Doroszlai Attila, 2016- Phil Whitehurst
 *
 *     Permission to use, copy, modify, and/or distribute this software
 *     for any purpose with or without fee is hereby granted, provided
 *     that the above copyright notice and this permission notice appear
 *     in all copies.
 *
 *     THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL
 *     WARRANTIES WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED
 *     WARRANTIES OF MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE
 *     AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT, INDIRECT, OR
 *     CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS
 *     OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT,
 *     NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF OR IN
 *     CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

// TODO: the "L.DistanceMarker" (canvas marker) class could be alternatively provided by "leaflet-rotate"? 

L.DistanceMarker = L.CircleMarker.extend({
	_updatePath: function () {
		let ctx = this._renderer._ctx;
		let p = this._point;

		// Calculate image direction (rotation)
		this.options.rotation = this.options.rotation || 0;

		// Draw circle marker (canvas point)
		if (this.options.radius && this._renderer._updateCircle) {
			this._renderer._updateCircle(this);
		}

		// Draw image over circle (distance marker)
		if (this.options.icon && this.options.icon.url) {
			if (!this.options.icon.element) {
				const icon = document.createElement('img');
				this.options.icon = L.extend({ rotate: 0, size: [40, 40], offset: { x: 0, y: 0 } }, this.options.icon);
				this.options.icon.rotate += this.options.rotation;
				this.options.icon.element = icon;
				icon.src = this.options.icon.url;
				icon.onload = () => this.redraw();
				icon.onerror = () => this.options.icon = null;
			} else {
				const icon = this.options.icon;
				let cx = p.x + icon.offset.x;
				let cy = p.y + icon.offset.y;
				ctx.save();
				if (icon.rotate) {
					ctx.translate(p.x, p.y);
					ctx.rotate(icon.rotate);
					cx = 0;
					cy = 0;
				}
				ctx.drawImage(icon.element, cx - icon.size[0] / 2, cy - icon.size[1] / 2, icon.size[0], icon.size[1]);
				ctx.restore();
			}
		}

		// Add a label inside the circle (distance marker)
		if (this.options.label) {
			let cx = p.x, cy = p.y;
			ctx.save();

			ctx.font = this.options.font || 'normal 7pt "Helvetica Neue", Arial, Helvetica, sans-serif';
			ctx.textAlign = "center";
			ctx.textBaseline = "middle";
			ctx.fillStyle = this.options.fillStyle || 'black';

			// TODO rescale circle to fit text
			// let fontSize  = Number(/[0-9\.]+/.exec(ctx.font)[0]);
			// let fontWidth = ctx.measureText(this.options.html).width;

			if (this.options.rotation) {
				ctx.translate(p.x, p.y);
				ctx.rotate(this.options.rotation);
				cx = 0;
				cy = 0;
			}

			// Temporary fix to prevent stroke blurs at higher zoom levels
			if (this._map.getZoom() > 17) {
				ctx.fillStyle = this.options.strokeStyle || 'black';
			}

			ctx.fillText(this.options.label, cx, cy);

			if (this.options.strokeStyle && this._map.getZoom() <= 17) {
				ctx.strokeStyle = this.options.strokeStyle;
				ctx.strokeText(this.options.label, cx, cy);
			}

			ctx.restore();
		}
	}
});

L.DistanceMarkers = L.LayerGroup.extend({
	options: {
		cssClass: 'dist-marker',
		iconSize: [12, 12],
		arrowSize: [10, 10],
		arrowUrl: "data:image/svg+xml,%3Csvg transform='rotate(90)' xmlns='http://www.w3.org/2000/svg' width='560px' height='560px' viewBox='0 0 560 560'%3E%3Cpath stroke-width='35' fill='%23000' stroke='%23FFF' d='M280,40L522,525L280,420L38,525z'/%3E%3C/svg%3E",
		offset: 1000,
		showAll: 12,
		textFunction: (distance, i, offset) => i,
		distance: true,
		direction: true,
	},
	initialize: function (line, map, options) {

		this._layers = {};
		this._zoomLayers = {};

		options = L.setOptions(this, options);

		let preferCanvas = map.options.preferCanvas;
		let showAll = Math.min(map.getMaxZoom(), options.showAll);

		// You should use "leaflet-rotate" to show rotated arrow markers (preferCanvas: false)
		if (!preferCanvas && !map.options.rotate) {
			console.warn('Missing dependency: "leaflet-rotate"');
		}

		// Get line coords as an array
		let coords = typeof line.getLatLngs == 'function' ? line.getLatLngs() : line;

		// Handle "MultiLineString" features
		coords = L.LineUtil.isFlat(coords) ? [coords] : coords;

		coords.forEach(latlngs => {
			// Get accumulated line lengths as well as overall length
			let accumulated = L.GeometryUtil.accumulatedLengths(latlngs);
			let length = accumulated.length > 0 ? accumulated[accumulated.length - 1] : 0;

			// count = Number of distance markers to be added
			// j = Position in accumulated line length array
			for (let i = 1, count = Math.floor(length / options.offset), j = 0; i <= count; ++i) {

				let distance = options.offset * i;

				// Find the first accumulated distance that is greater than the distance of this marker
				while (j < accumulated.length - 1 && accumulated[j] < distance) ++j;

				// Grab two nearest points either side marker position
				let p1 = latlngs[j - 1];
				let p2 = latlngs[j];
				let m_line = L.polyline([p1, p2]);

				// and create a simple line to interpolate on
				let ratio = (distance - accumulated[j - 1]) / (accumulated[j] - accumulated[j - 1]);
				let position = L.GeometryUtil.interpolateOnLine(map, m_line, ratio);
				let delta = map.project(p2).subtract(map.project(p1));
				let angle = Math.atan2(delta.y, delta.x);

				// Generate distance marker label
				let text = options.textFunction.call(this, distance, i, options.offset);

				// Grouping layer of visible layers at zoom level (arrow + distance)
				let zoom = this._minimumZoomLevelForItem(i, showAll);
				let markers = this._zoomLayers[zoom] = this._zoomLayers[zoom] || L.layerGroup()

				// create arrow markers
				if (options.direction && ((options.distance && i % 2 == 1) || !options.distance)) {
					if (preferCanvas) {
						markers.addLayer(
							new L.DistanceMarker(p1, {
								radius: 0,
								icon: {
									url: options.arrowUrl,   //image link
									size: options.arrowSize, //image size ( default [40, 40] )
									rotate: 0,               //image base rotate ( default 0 )
									offset: { x: 0, y: 0 },  //image offset ( default { x: 0, y: 0 } )
								},
								rotation: angle,
								interactive: false,
								// label: '⮞', //'➜',
								// font: 'normal 20pt "Helvetica Neue", Arial, Helvetica, sans-serif',
								// fillStyle: 'white',//'#3366CC',
								// strokeStyle: 'black',
							})
						);
					} else {
						markers.addLayer(
							L.marker(position.latLng, {
								icon: L.icon({
									iconUrl: options.arrowUrl,
									iconSize: options.arrowSize,
								}),
								// NB the following option is added by "leaflet-rotate"
								rotation: angle,
								interactive: false,
							})
						);
					}
				}

				// create distance markers
				if (options.distance && i % 2 == 0) {
					if (preferCanvas) {
						markers.addLayer(
							new L.DistanceMarker(position.latLng, {
								label: text, // TODO: handle text rotation (leaflet-rotate)
								radius: 7,
								fillColor: '#fff',
								fillOpacity: 1,
								fillStyle: 'black',
								color: '#777',
								weight: 1,
								interactive: false,
							})
						);
					} else {
						markers.addLayer(
							L.marker(position.latLng, {
								title: text,
								icon: L.divIcon({
									className: options.cssClass,
									html: text,
									iconSize: options.iconSize
								}),
								interactive: false,
							})
						);
					}
				}
			}
		});

		const updateMarkerVisibility = () => {
			let oldZoom = this._lastZoomLevel || 0;
			let newZoom = map.getZoom();
			if (newZoom > oldZoom) {
				for (let i = oldZoom + 1; i <= newZoom; ++i) {
					if (this._zoomLayers[i] !== undefined) {
						this.addLayer(this._zoomLayers[i]);
					}
				}
			} else if (newZoom < oldZoom) {
				for (let i = oldZoom; i > newZoom; --i) {
					if (this._zoomLayers[i] !== undefined) {
						this.removeLayer(this._zoomLayers[i]);
					}
				}
			}
			this._lastZoomLevel = newZoom;
		};
		map.on('zoomend', updateMarkerVisibility);
		updateMarkerVisibility();
	},

	_minimumZoomLevelForItem: function (i, zoom) {
		while (i > 0 && i % 2 === 0) {
			--zoom;
			i = Math.floor(i / 2);
		}
		return zoom;
	},

});

L.Polyline.include({

	_originalOnAdd: L.Polyline.prototype.onAdd,
	_originalOnRemove: L.Polyline.prototype.onRemove,

	addDistanceMarkers: function () {
		if (this._map && this._distanceMarkers) {
			this._map.addLayer(this._distanceMarkers);
		}
	},

	removeDistanceMarkers: function () {
		if (this._map && this._distanceMarkers) {
			this._map.removeLayer(this._distanceMarkers);
		}
	},

	onAdd: function (map) {
		this._originalOnAdd(map);

		let opts = this.options.distanceMarkers || {};
		if (this.options.distanceMarkers) {
			this._distanceMarkers = this._distanceMarkers || new L.DistanceMarkers(this, map, opts);
		}
		if (opts.lazy === undefined || opts.lazy === false) {
			this.addDistanceMarkers();
		}
	},

	onRemove: function (map) {
		this.removeDistanceMarkers();
		this._originalOnRemove(map);
	}

});
