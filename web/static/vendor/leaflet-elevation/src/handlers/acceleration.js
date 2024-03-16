export function Acceleration() {

	const _ = L.Control.Elevation.Utils;

	let opts         = this.options;
	let acceleration = {};

	acceleration.label      = opts.accelerationLabel  || L._(opts.imperial ? 'ft/s²' : 'm/s²');
	opts.accelerationFactor = opts.accelerationFactor || 1;

	return {
		name: 'acceleration',
		unit: acceleration.label,
		deltaMax: this.options.accelerationDeltaMax,
		clampRange: this.options.accelerationRange,
		decimals: 2,
		pointToAttr: (_, i) => {
			let dv     = (this._data[i].speed - this._data[i > 0 ? i - 1 : i].speed) * (1000 / opts.timeFactor);
			let dt     = (this._data[i].time - this._data[i > 0 ? i - 1 : i].time) / 1000;
			return dt > 0 ? Math.abs((dv / dt)) * opts.accelerationFactor : NaN;
		},
		stats: { max: _.iMax, min: _.iMin, avg: _.iAvg },
		scale: {
			axis       : "y",
			position   : "right",
			scale      : { min: 0, max: +1 },
			tickPadding: 16,
			labelX     : 25,
			labelY     : -8,
		},
		path: {
			label        : 'Acceleration',
			yAttr        : 'acceleration',
			scaleX       : 'distance',
			scaleY       : 'acceleration',
			color        : '#050402',
			strokeColor  : '#000',
			strokeOpacity: "0.5",
			fillOpacity  : "0.25",
		},
		tooltip: {
			chart: (item) => L._("a: ") + item.acceleration + " " + acceleration.label,
			marker: (item) => Math.round(item.acceleration) + " " + acceleration.label,
			order: 60,
		},
		summary: {
			"minacceleration"  : {
				label: "Min Acceleration: ",
				value: (track, unit) => Math.round(track.acceleration_min || 0) + '&nbsp;' + unit,
				order: 60
			},
			"maxacceleration"  : {
				label: "Max Acceleration: ",
				value: (track, unit) => Math.round(track.acceleration_max || 0) + '&nbsp;' + unit,
				order: 61
			},
			"avgacceleration": {
				label: "Avg Acceleration: ",
				value: (track, unit) => Math.round(track.acceleration_avg || 0) + '&nbsp;' + unit,
				order: 62
			},
		}
	};
}
