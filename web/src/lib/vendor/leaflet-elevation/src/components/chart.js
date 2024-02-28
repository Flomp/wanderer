import * as D3 from './d3.js';

const _ = L.Control.Elevation.Utils;

export var Chart = L.Control.Elevation.Chart = L.Class.extend({

	includes: L.Evented ? L.Evented.prototype : L.Mixin.Events,

	initialize(opts, control) {
		this.options       = opts;
		this.control       = control;

		this._data         = control._data || [];

		// cache registered components
		this._props        = {
			scales      : {},
			paths       : {},
			areas       : {},
			grids       : {},
			axes        : {},
			legendItems : {},
			tooltipItems: {},
		};

		this._scales       = {};
		this._domains      = {};
		this._ranges       = {};
		this._paths        = {};

		this._brushEnabled = opts.dragging;
		this._zoomEnabled  = opts.zooming;

		opts.xTicks = this._xTicks();
		opts.yTicks = this._yTicks();
		
		let chart       = this._chart = D3.Chart(opts);

		// SVG Container
		this._container = chart.svg;

		// Panes
		this._grid      = chart.pane('grid');
		this._area      = chart.pane('area');
		this._point     = chart.pane('point');
		this._axis      = chart.pane('axis');
		this._legend    = chart.pane('legend');
		this._tooltip   = chart.pane('tooltip');
		this._ruler     = chart.pane('ruler');

		// Scales
		this._initScale();

		// Helpers
		this._mask      = chart.get('mask');
		this._context   = chart.get('context');
		this._brush     = chart.get('brush');

		this._zoom      = d3.zoom();
		this._drag      = d3.drag();

		// Interactions
		this._initInteractions();

		// svg.on('resize', (e)=>console.log(e.detail));

		// Handle multi-track segments
		this._maskGaps = [];
		control.on('eletrack_added', ({index}) => {
			this._maskGaps.push(index);
			control.once('elepoint_added', ({index}) => this._maskGaps.push(index));
		});

	},

	update(props) {
		if (props) {
			if (props.data) this._data = props.data;
			if (props.options) this.options = props.options;
		}

		this.options.xTicks = this._xTicks();
		this.options.yTicks = this._yTicks();

		this._updateScale();
		this._updateAxis();
		this._updateMargins();
		this._updateLegend();
		this._updateClipper();
		this._updateArea();

		return this;
	},

	render() {
		return container => container.append(() => this._container.node());
	},

	clear() {
		this._resetDrag();
		this._hideDiagramIndicator()
		this._area.selectAll('path').attr("d", "M0 0");
		this._context.clearRect(0, 0, this._width(), this._height());
		this._mask.selectAll(".gap").remove();
		this._maskGaps = [];

		// if (this._path) {
			// this._x.domain([0, 1]);
			// this._y.domain([0, 1]);
		// }
	},

	_drawPath(name) {
		let path   = this._paths[name];
		let area   = this._props.areas[name];

		path.datum(this._data).attr("d",
			D3.Area(
				L.extend({}, area, {
					width        : this._width(),
					height       : this._height(),
					scaleX       : this._scales[area.scaleX],
					scaleY       : this._scales[area.scaleY]
				})
			)
		);

		if (!path.classed('leaflet-hidden')) {
			this.options.preferCanvas
				? _.drawCanvas(this._context, path)
				: _.append(this._area.node(), path.node());
		}
	},

	_hasActiveLayers() {
		const paths = this._paths;
		for (var i in paths) {
			if (!paths[i].classed('leaflet-hidden')) {
				return true;
			}
		}
		return false;
	},

	/**
	 * Initialize "d3-brush".
	 */
	_initBrush(e) {
		const brush   = ({selection}) => {
			if (selection && this._data.length) {
				let start = this._findIndexForXCoord(selection[0]);
				let end   = this._findIndexForXCoord(selection[1]);
				this.fire('dragged', { dragstart: this._data[start], dragend: this._data[end] });
			}
		}

		const focus   = (e) => {
			if (this._data.length && (e.type != 'brush' /*|| e.sourceEvent*/)) {
				let rect    = this._chart.panes.brush.select('.overlay').node();
				let coords  = d3.pointers(e, rect)[0];
				let xCoord  = coords[0];
				let item    = this._data[this._findIndexForXCoord(xCoord)];

				this.fire("mouse_move", { item: item, xCoord: xCoord });
			}
		};

		this._brush
			.filter(({shiftKey, button}) => !shiftKey && !button && this._brushEnabled)
			.on("end.update", brush)
			.on("brush.update", focus);

		this._chart.panes.brush
			.on("mouseenter.focus touchstart.focus", this.fire.bind(this, "mouse_enter"))
			.on("mouseout.focus touchend.focus",     this.fire.bind(this, "mouse_out")  )
			.on("mousemove.focus touchmove.focus",   focus                              );

	},

	/**
	 * Initialize "d3-zoom"
	 */
	_initClipper() {
		let svg       = this._container;
		let margin    = this.options.margins;

		const zoom    = this._zoom;

		const onStart = ({transform, sourceEvent}) => {
			if (sourceEvent && sourceEvent.type == "mousedown") svg.style('cursor', 'grabbing');
			if (transform.k == 1 && transform.x == 0) {
				this._container.classed('zoomed', true);
				// Apply d3-zoom and bind <mask>
				if (this._mask) {
					this._point.attr('mask', 'url(#' + this._mask.attr('id') + ')');
				}
			}
			this.zooming = true;
		};

		const onEnd  = ({transform}) => {
			if (transform.k ==1 && transform.x == 0) {
				this._container.classed('zoomed', false);
				// Reset d3-zoom and remove <mask>
				if (this._mask) {
					this._point.attr('mask', null);
				}
			}
			this.zooming = false;
			svg.style('cursor', '');
		};

		const onZoom = ({transform, sourceEvent}) => {
			// TODO: find a faster way to redraw the chart.
			this.zooming = false;
			this._updateScale(); // hacky way for restoring x scale when zooming out
			this.zooming = true;
			this._scales.distance = this._x = transform.rescaleX(this._x); // calculate x scale at zoom level
			if (this._scales.time) this._scales.time = transform.rescaleX(this._scales.time); // calculate x scale at zoom level
			this._resetDrag();
			if (sourceEvent && sourceEvent.type == "mousemove") {
				this._hideDiagramIndicator();
			}
			this.fire('zoom');
		};

		zoom
			.scaleExtent([1, 10])
			.extent([
				[margin.left, 0],
				[this._width() - margin.right, this._height()]
			])
			.translateExtent([
				[margin.left, -Infinity],
				[this._width() - margin.right, Infinity]
			])
			.filter(({shiftKey, buttons}) => (shiftKey || buttons == 4) && this._zoomEnabled)
			.on("start", onStart)
			.on("end", onEnd)
			.on("zoom", onZoom);

		svg.call(zoom); // add zoom functionality to "svg" group

		// d3.select("body").on("keydown.grabzoom keyup.grabzoom", (e) => svg.style('cursor', e.shiftKey ? 'move' : ''));

	},

	_initInteractions() {
		this._initBrush();
		this._initRuler();
		this._initClipper();
		this._initLegend();
	},

	/**
	 * Toggle chart data on legend click
	 */
	_initLegend() {
		this._container.on('legend_clicked', ({detail}) => {
			let { path, legend, name, enabled } = detail;
			if (path) {
				let label = _.select('text', legend);
				let rect  = _.select('rect', legend);
				_.toggleStyle(label, 'text-decoration-line', 'line-through', enabled);
				_.toggleStyle(rect,  'fill-opacity',         '0',            enabled);
				_.toggleClass(path,  'leaflet-hidden',                       enabled);
				this._updateArea();
				this.fire("elepath_toggle", { path, name, legend, enabled })
			}
		});
	},

	/**
	 * Initialize "ruler".
	 */
	_initRuler() {
		if (!this.options.ruler) return;

		// const yMax      = this._height();
		const formatNum = d3.format(".0f");
		const drag      = this._drag;

		const label     = (e, d) => {
			let yMax      = this._height();
			let y         = this._ruler.data()[0].y;
			if (y >= yMax || y <= 0) this._ruler.select(".horizontal-drag-label").text('');
			this._hideDiagramIndicator();
		};

		const position  = (e, d) => {
			let yCoord    = d3.pointers(e, this._area.node())[0][1];
			let yMax      = this._height();
			let y         = _.clamp(yCoord, [0, yMax]);

			this._ruler
				.data([L.extend(this._ruler.data()[0], { y: y })])
				.attr("transform", d => "translate(" + d.x + "," + d.y + ")")
				.classed('active', y < yMax);

			this._container
				.select(".horizontal-drag-label")
				.text(formatNum(this._y.invert(y)) + " " + (this.options.imperial ? 'ft' : 'm'));

			this.fire('ruler_filter', { coords: yCoord < yMax && yCoord > 0 ? this._findCoordsForY(yCoord) : [] });
		}

		drag
			.on("start end", label)
			.on("drag", position);

		this._ruler.call(drag);

	},

	/**
	 * Initialize x and y scales
	 */
	_initScale() {
		let opts = this.options;

		this._registerAxisScale({
			axis    : 'x',
			position: 'bottom',
			attr    : opts.xAttr,
			min     : opts.xAxisMin,
			max     : opts.xAxisMax,
			name    : 'distance'
		});

		this._registerAxisScale({
			axis     : 'y',
			position : 'left',
			attr     : opts.yAttr,
			min      : opts.yAxisMin,
			max      : opts.yAxisMax,
			name     : 'altitude'
		});

		this._x = this._scales.distance;
		this._y = this._scales.altitude;
	},

	_registerAreaPath(props) {
		if (props.scale == 'y') props.scale = this._y;
		else if (props.scale == 'x') props.scale = this._x;

		let opts = this.options;

		if (!props.xAttr) props.xAttr = opts.xAttr;
		if (!props.yAttr) props.yAttr = opts.yAttr;
		if (typeof props.preferCanvas === "undefined") props.preferCanvas = opts.preferCanvas;

		// Save paths in memory for latter usage
		this._paths[props.name]       = D3.Path(props);
		this._props.areas[props.name] = props;

		if (opts.legend) {
			this._props.legendItems[props.name] = {
				name     : props.name,
				label    : props.label,
				color    : props.color,
				className: props.className,
				path     : this._paths[props.name]
			};
		}

	},

	_registerAxisGrid(props) {
		if (props.scale == 'y') props.scale = this._y;
		else if (props.scale == 'x') props.scale = this._x;

		this._props.grids[props.name || props.axis] = props;
	},

	_registerAxisScale(props) {
		if (props.scale == 'y') props.scale = this._y;
		else if (props.scale == 'x') props.scale = this._x;

		let opts  = this.options;
		let scale = props.scale;

		if (typeof this._scales[props.name] === 'function') {
			props.scale = this._scales[props.name]; // retrieve cached scale
		} else if (typeof scale !== 'function') {
			scale       = L.extend({
				data       : this._data,
				forceBounds: opts.forceAxisBounds
			}, scale);

			scale.attr   = scale.attr || props.name;

			let domain   = this._domains[props.name] = D3.Domain(props);
			let range    = this._ranges[props.name]  = D3.Range(props);
			scale.range  = scale.range  || range(this._width(),this._height());
			scale.domain = scale.domain || domain(this._data);

			this._props.scales[props.name] = scale;

			props.scale = this._scales[props.name] = D3.Scale(scale);
		}

		if (!props.ticks) {
			if (props.axis == 'x')      props.ticks = this._xTicks.bind(this);
			else if (props.axis == 'y') props.ticks = this._yTicks.bind(this);
		}

		this._props.axes[props.name] = props;

		if (props.name == this.options.yScale) this._y = scale;
		if (props.name == this.options.xScale) this._x = scale;

		return scale;
	},

	/**
	 * Add a point of interest over the chart
	 */
	_registerCheckPoint(point) {
		if (!this._data.length) return;

		const {xAttr, yAttr} = this.options;
		let item, x, y;

		if (point.latlng) {
			item = this._data[this._findIndexForLatLng(point.latlng)];
			x    = this._x(item[xAttr]);
			y    = this._y(item[yAttr]);
		} else if (!isNaN(point[xAttr])) {
			x    = this._x(point[xAttr]);
			item = this._data[this._findIndexForXCoord(x)]
			y    = this._y(item[yAttr]);
		} 

		this._point.call(D3.CheckPoint({
			point: point,
			width: this._width(),
			height: this._height(),
			x: x,
			y: y,
		}));
	},

	_registerTooltip(props) {
		props.order = props.order ?? 1000;
		this._props.tooltipItems[props.name] = props;
	},

	_updateArea() {
		// Reset and update chart profiles
		this._context.clearRect(0, 0, this._width(), this._height());
		_.each(this._paths, (path, i) => !path.classed('leaflet-hidden') && this._drawPath(i));
	},

	_updateAxis() {

		let opts = this.options;

		const gridOpts = {
			width     : this._width(),
			height    : this._height(),
			tickFormat: "",
		};

		const axesOpts  = {
			width  : this._width(),
			height : this._height(),
		};

		// Reset grids
		this._grid.selectAll('g').remove();
		_.each(this._props.grids, (grid, i) => {
			if (opts[i] !== false && opts[i] !== 'summary') {
				this._grid.call(D3.Grid(L.extend({}, gridOpts, grid)))
			}
		});

		// Rest axis
		this._axis.selectAll('g').remove();
		_.each(this._props.axes,  (axis, i) => {
			if (opts[i] !== false && opts[i] !== 'summary') {
				this._axis.call(D3.Axis(L.extend({}, axesOpts, axis)));
			}
		});

		// Adjust axis scale positions
		this._axis
			.selectAll('.y.axis.right')
			.each((d, i, n) => {

				let axis      = d3.select(n[i]);
				let transform = axis.attr('transform');
				let translate = transform.substring(transform.indexOf("(") + 1, transform.indexOf(")")).split(",");

				axis.attr('transform', 'translate(' + (+translate[0] + (i * 40)) + ',' + translate[1] + ')')

				if (i > 0) {
					axis.select(':scope > path')         .attr('opacity', 0.25);
					axis.selectAll(':scope > .tick line').attr('opacity', 0.75);
				}

			});

	},

	_updateClipper() {
		const { xAttr, margins } = this.options;
		const data = this._data;

		this._zoom
			.scaleExtent([1, 10])
			.extent([
				[margins.left, 0],
				[this._width() - margins.right, this._height()]
			])
			.translateExtent([
				[margins.left, -Infinity],
				[this._width() - margins.right, Infinity]
			]);

		// Apply svg mask on multi-track segments
		this._mask.selectAll(".gap").remove()
		this._maskGaps.forEach((d, i) => {
			if (i >= this._maskGaps.length - 2) return;
			let x1 = this._x(data[this._findIndexForLatLng(data[this._maskGaps[i]].latlng)][xAttr]);
			let x2 = this._x(data[this._findIndexForLatLng(data[this._maskGaps[i + 1]].latlng)][xAttr]);
			this._mask
				.append("rect")
				.attr("x", x1)
				.attr("y", 0)
				.attr("width", x2 - x1 )
				.attr("height", this._height())
				.attr('class', 'gap')
				.attr('fill-opacity', '0.8')
				.attr("fill", 'black');  // hide = black (mask)
		});
	},

	_updateLegend: function () {
		let xAxesB  = this._axis.selectAll('.x.axis.bottom').nodes().length;
		
		// Get legend items
		let items = Object.keys(this._paths);

		// Calculate legend item positions
		let n      = items.length;
		let v      = Array(Math.floor(n / 2)).fill(null).map((d, i) => (i + 1) * 2 - (1 - Math.sign(n % 2)));
		let rev    = v.slice().reverse().map((d) => -(d));

		// push a fake element to handle center alignment (odd numbers)
		if (n % 2 !== 0) {
			rev.push(0);
		}
		v = rev.concat(v);

		// Reset legend items
		this._legend.selectAll('g').remove();

		// Render legend items
		_.each(this._props.legendItems, (legend) => {
			this._legend.append("g").call(
				D3.LegendItem(L.extend({
					width  : this._width(),
					height : this._height(),
					margins: this.options.margins,
				}, legend))
			);
		});

		// Render legend item switcher
		if (n > 1) {
			this._legend.append("g").call(
				D3.LegendSmall({
					width  : this._width(),
					height : this._height(),
					items  : items,
					onClick: (selected) => {
						_.each(items, name => this._togglePath(name, selected == name, true));
						this._updateArea();
					}
				})
			);
		}
		
		_.each(items, (name, i) => {
			// Adjust legend item positions
			this._legend.select('[data-name=' + name + ']').attr("transform", "translate(" + (v[i] * 55) + ", " + (xAxesB * 2) + ")");
			// Set initial state (disabled controls)
			this._togglePath(name, !(name in this.options && this.options[name] == 'disabled'), true);
		});

	},

	_updateMargins() {
		// Get chart margins
		let xAxesB  = this._axis.selectAll('.x.axis.bottom').nodes().length;
		let xAxesL  = this._axis.selectAll('.y.axis.left').nodes().length;
		let xAxesR  = this._axis.selectAll('.y.axis.right').nodes().length;
		let marginB = 60 + (xAxesB * 2);
		let marginL = 10 + (xAxesL * 30);
		let marginR = 40 + (xAxesR * 30);

		let marginsUpdated = false

		// Adjust right margin
		if (xAxesR && this.options.margins.right < marginR) {
			this.options.margins.right  = marginR;
			marginsUpdated = true;
		}

		// Adjust left margin
		if (xAxesL && this.options.margins.left < marginL) {
			this.options.margins.left = marginL;
			marginsUpdated = true;
		}

		// Adjust bottom margin
		if (xAxesB && this.options.margins.bottom < marginB) {
			this.options.margins.bottom = marginB;
			marginsUpdated = true;
		}

		if (marginsUpdated) {
			this.fire('margins_updated');
		}
	},

	_updateScale() {
		if (this.zooming) return { x: this._x, y: this._y };

		for (let i in this._scales) {
			this._scales[i]
				.domain(this._domains[i](this._data))
				.range(this._ranges[i](this._width(), this._height()))
		}

		return { x: this._x, y: this._y };
	},

	/**
	 * Calculates chart width.
	 */
	_width() {
		if (this._chart) this._chart._width;
		const { width, margins } = this.options;
		return width - margins.left - margins.right;
	},

	/**
	 * Calculates chart height.
	 */
	_height() {
		if (this._chart) return this._chart._height;
		const { height, margins } = this.options;
		return height - margins.top - margins.bottom;
	},

	/*
	 * Finds data entries above a given y-elevation value and returns geo-coordinates
	 */
	_findCoordsForY(y) {
		let data = this._data;
		let z    = this._y.invert(y);

		// save indexes of elevation values above the horizontal line
		const list = data.reduce((array, item, index) => {
			if (item[this.options.yAttr] >= z) array.push(index);
			return array;
		}, []);

		let start = 0;
		let next;

		// split index list into blocks of coordinates
		const coords = list.reduce((array, _, curr) => {
			next = curr + 1;
			if (list[next] !== list[curr] + 1 || next === list.length) {
				array.push(
					list
					.slice(start, next)
					.map(i => data[i].latlng)
				);
				start = next;
			}
			return array;
		}, []);

		return coords;
	},

	/*
	 * Finds a data entry for a given x-coordinate of the diagram
	 */
	_findIndexForXCoord(x) {
		return d3
			.bisector(d => d[this.options.xAttr])
			.left(this._data || [0, 1], this._x.invert(x));
	},

	/*
	 * Finds a data entry for a given latlng of the map
	 */
	_findIndexForLatLng(latlng) {
		let result = null;
		let d      = Infinity;
		this._data.forEach((item, index) => {
			let dist = latlng.distanceTo(item.latlng);
			if (dist < d) {
				d = dist;
				result = index;
			}
		});
		return result;
	},

	/*
	 * Removes the drag rectangle and zoms back to the total extent of the data.
	 */
	_resetDrag() {
		if (this._chart.panes.brush.select(".selection").attr('width')) {
			this._chart.panes.brush.call(this._brush.clear);
			this._hideDiagramIndicator();
			this.fire('reset_drag');
		}
	},

	_resetZoom() {
		if (this._zoom) {
			this._zoom.transform(this._chart.svg, d3.zoomIdentity);
		}
	},

	/**
	 * Display distance and altitude level ("focus-rect").
	 */
	_showDiagramIndicator(item, xCoordinate) {
		this._tooltip
			.attr("display", null)
			.call(D3.Tooltip({
				xCoord: xCoordinate,
				yCoord: this._y(item[this.options.yAttr]),
				height: this._height(),
				width : this._width(),
				labels: this._props.tooltipItems,
				item: item
			}));
	},

	_togglePath(name, enabled = true, lazy = false) {
		let path    = this._paths[name];
		let legend  = this._container.select('.legend [data-name=' + name + ']');
		path.classed('leaflet-hidden', !enabled);
		legend.select('text').style('text-decoration-line', enabled ? '' : 'line-through');
		legend.select('rect').style('fill-opacity', enabled ? '': '0');
		// Apply d3-zoom (bind <clipPath> mask)
		if (this._mask) {
			path.attr('mask', 'url(#' + this._mask.attr('id') + ')');
		}
		if (!lazy) {
			this._updateArea();
		}
	},

	_hideDiagramIndicator() {
		this._tooltip.attr("display", 'none');
	},

	/**
	 * Calculate chart xTicks
	 */
	_xTicks() {
	if (this.__xTicks) this.__xTicks = this.options.xTicks;
	return this.__xTicks || Math.round(this._width() / 75);
},

	/**
	 * Calculate chart yTicks
	 */
	_yTicks() {
		if (this.__yTicks) this.__yTicks = this.options.yTicks;
		return this.__yTicks || Math.round(this._height() / 30);
	}
});
