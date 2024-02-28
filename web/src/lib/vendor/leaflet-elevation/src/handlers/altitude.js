export function Altitude() {
	
	const _  = L.Control.Elevation.Utils;

	let opts       = this.options;
	let altitude   = {};

	let theme           = opts.theme.split(' ')[0].replace('-theme', '');
	let color           = _.Colors[theme] || {};

	opts.altitudeFactor = opts.imperial ? this.__footFactor : (opts.altitudeFactor || 1); // 1 m = (1 m)
	altitude.label      = opts.imperial ? "ft" : opts.yLabel;

	return {
		name: 'altitude',
		required: this.options.slope,
		meta: 'z',
		unit: altitude.label,
		statsName: 'elevation',
		deltaMax: this.options.altitudeDeltaMax,
		clampRange: this.options.altitudeRange,
		// init: ({point}) => {
		// 	// "alt" property is generated inside "leaflet"
		// 	if ("alt" in point) point.meta.ele = point.alt;
		// },
		pointToAttr: (point, i) => {
			if ("alt" in point) point.meta.ele = point.alt; // "alt" property is generated inside "leaflet"
			return this._data[i].z *= opts.altitudeFactor;
		},
		stats: { max: _.iMax, min: _.iMin, avg: _.iAvg },
		grid: {
			axis      : "y",
			position  : "left",
			scale     : "y" // this._chart._y,
		},
		scale: {
			axis    : "y",
			position: "left",
			scale   : "y", // this._chart._y,
			labelX  : -3,
			labelY  : -8,
		},
		path: {
			label        : 'Altitude',
			scaleX       : 'distance',
			scaleY       : 'altitude',
			className    : 'area',
			color        : color.area || theme,
			strokeColor  : opts.detached ? color.stroke : '#000',
			strokeOpacity: "1",
			fillOpacity  : opts.detached ? (color.alpha || '0.8') : 1,
			preferCanvas : opts.preferCanvas,
		},
		tooltip: {
			name: 'y',
			chart: (item) => L._("y: ") + _.round(item[opts.yAttr], opts.decimalsY) + " " + altitude.label,
			marker: (item) => _.round(item[opts.yAttr], opts.decimalsY) + " " + altitude.label,
			order: 10,
		},
		summary: {
			"minele"  : {
				label: "Min Elevation: ",
				value: (track, unit) => (track.elevation_min || 0).toFixed(2) + '&nbsp;' + unit,
				order: 30,
			},
			"maxele"  : {
				label: "Max Elevation: ",
				value: (track, unit) => (track.elevation_max || 0).toFixed(2) + '&nbsp;' + unit,
				order: 31,
			},
			"avgele"  : {
				label: "Avg Elevation: ",
				value: (track, unit) => (track.elevation_avg || 0).toFixed(2) + '&nbsp;' + unit,
				order: 32,
			},
		}
	};
}
