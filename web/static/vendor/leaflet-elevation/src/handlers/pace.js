export function Pace() {

	const _ = L.Control.Elevation.Utils;

	let opts = this.options;
	let pace = {};

	pace.label      = opts.paceLabel  || L._(opts.imperial ? 'min/mi' : 'min/km');
	opts.paceFactor = opts.paceFactor || 60; // 1 min = 60 sec

	return {
		name: 'pace',
		unit: pace.label,
		deltaMax: this.options.paceDeltaMax,
		clampRange: this.options.paceRange,
		decimals: 2,
		pointToAttr: (_, i) => {
			let dx   = (this._data[i].dist - this._data[i > 0 ? i - 1 : i].dist) * 1000;
			let dt   = this._data[i].time - this._data[ i > 0 ? i - 1 : i].time;
			return dx > 0 ? Math.abs((dt / dx) / opts.paceFactor) : NaN;
		},
		stats: { max: _.iMax, min: _.iMin, avg: _.iAvg },
		scale : (this.options.pace && this.options.pace != "summary") && {
			axis       : "y",
			position   : "right",
			scale      : { min : 0, max : +1 },
			tickPadding: 16,
			labelX     : 25,
			labelY     : -8,
		},
		path: (this.options.pace && this.options.pace != "summary") && {
			// name         : 'pace',
			label        : 'Pace',
			yAttr        : "pace",
			scaleX       : 'distance',
			scaleY       : 'pace',
			color        : '#03ffff',
			strokeColor  : '#000',
			strokeOpacity: "0.5",
			fillOpacity  : "0.25",
		},
		tooltip: (this.options.pace) && {
			chart: (item) => L._('pace: ') + item.pace + " " + pace.label,
			marker: (item) => Math.round(item.pace) + " " + pace.label,
			order: 50,
		},
		summary: (this.options.pace) && {
			"minpace"  : {
				label: "Min Pace: ",
				value: (track, unit) => Math.round(track.pace_min || 0) + '&nbsp;' + unit,
				order: 51
			},
			"maxpace"  : {
				label: "Max Pace: ",
				value: (track, unit) => Math.round(track.pace_max || 0) + '&nbsp;' + unit,
				order: 51
			},
			"avgpace": {
				label: "Avg Pace: ",
				value: (track, unit) => Math.round(track.pace_avg || 0) + '&nbsp;' + unit,
				order: 52
			},
		}
	};
}
