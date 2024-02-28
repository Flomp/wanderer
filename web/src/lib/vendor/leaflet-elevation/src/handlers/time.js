export function Time() {

	const _ = L.Control.Elevation.Utils;

	let opts = this.options;
	let time = {};

	time.label      = opts.timeLabel  || 't';
	opts.timeFactor = opts.timeFactor || 3600;

	/**
	 * Common AVG speeds:
	 * ----------------------
	 *  slow walk = 1.8  km/h
	 *  walking   = 3.6  km/h <-- default: 3.6
	 *  running   = 10.8 km/h
	 *  cycling   = 18   km/h
	 *  driving   = 72   km/h
	 * ----------------------
	 */
	this._timeAVGSpeed = (opts.timeAVGSpeed || 3.6) * (opts.speedFactor || 1);

	if (!opts.timeFormat) {
		opts.timeFormat = (time) => (new Date(time)).toLocaleString().replaceAll('/', '-').replaceAll(',', ' ');
	} else if (opts.timeFormat == 'time') {
		opts.timeFormat = (time) => (new Date(time)).toLocaleTimeString();
	} else if (opts.timeFormat == 'date') {
		opts.timeFormat = (time) => (new Date(time)).toLocaleDateString();
	}

	opts.xTimeFormat = opts.xTimeFormat || ((t) => _.formatTime(t).split("'")[0]);

	return {
		name: 'time',
		required: (this.options.speed || this.options.acceleration || this.options.timestamps),
		coordinateProperties: ["coordTimes", "times", "time"],
		coordPropsToMeta: _.parseDate,
		pointToAttr: function(point, i) {
			// Add missing timestamps (see: options.timeAVGSpeed)
			if (!point.meta || !point.meta.time) {
				point.meta = point.meta || {};
				if (i > 0) {
					let dx = (this._data[i].dist - this._data[i - 1].dist);
					let t0 = this._data[i - 1].time.getTime();
					point.meta.time = new Date(t0 + ( dx / this._timeAVGSpeed) * this.options.timeFactor * 1000);
				} else {
					point.meta.time = new Date(Date.now())
				}
			}
			// Handle timezone offset
			let time = (point.meta.time.getTime() - point.meta.time.getTimezoneOffset() * 60 * 1000 !== 0) ? point.meta.time : 0;
			// Update duration
			this._data[i].duration = i > 0 ? (this._data[i - 1].duration || 0) + Math.abs(time - this._data[i - 1].time) : 0;
			return time;
		},
		onPointAdded: (_, i) => this.track_info.time = this._data[i].duration,
		scale: (opts.time && opts.time != "summary" && !L.Browser.mobile) && {
			axis       : "x",
			position   : "top",
			scale      : {
				attr       : "duration",
				min        : 0,
			},
			label      : time.label,
			labelY     : -10,
			labelX     : () => this._width(),
			name       : "time",
			ticks      : () => _.clamp(this._chart._xTicks() / 2, [4, +Infinity]),
			tickFormat : (d)  => (d == 0 ? '' : opts.xTimeFormat(d)),
			onAxisMount: axis => {
				axis.select(".domain")
					.remove();
				axis.selectAll("text")
					.attr('opacity', 0.65)
					.style('font-family', 'Monospace')
					.style('font-size', '110%');
				axis.selectAll(".tick line")
					.attr('y2', this._height())
					.attr('stroke-dasharray', 2)
					.attr('opacity', 0.75);
				}
		},
		tooltips: [
			(this.options.time) && {
				name: 'time',
				chart: (item) => L._("T: ") + _.formatTime(item.duration || 0),
				order: 20
			},
			(this.options.timestamps) && {
				name: 'date',
				chart: (item) => L._("t: ") + this.options.timeFormat(item.time),
				order: 21,
			}
		],
		summary: (this.options.time) && {
			"tottime"  : {
				label: "Total Time: ",
				value: (track) => _.formatTime(track.time || 0),
				order: 20
			}
		}
	};
}