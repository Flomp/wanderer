export function Temperature() {

	const _ = L.Control.Elevation.Utils;

	let temperature = {};
	let opts = this.options;
	temperature.label = opts.label || L._(opts.imperial ? '°F' : '°C');

	// Fahrenheit = (Celsius * 9/5) + 32
	opts.temperatureFactor1 = opts.temperatureFactor1 ?? (opts.imperial ? 1.8 : 1);
	opts.temperatureFactor2 = opts.temperatureFactor2 ?? (opts.imperial ? 32 : 0);

	return {
		name: 'temperature',              // <-- Your custom option name (eg. "temperature: true")
		unit: temperature.label,
		meta: 'atemps',                   // <-- point.meta.atemps
		coordinateProperties: ["atemps"], // List of GPX Extensions ("coordinateProperties") to be handled by "@tmcw/toGeoJSON"
		deltaMax: this.options.temperatureDeltaMax,
		clampRange: this.options.temperatureRange,
		decimals: 2,
		pointToAttr: (point, i) => (point.meta.atemps ?? point.meta.atemps ?? point.prev('temperature')) * opts.temperatureFactor1 + opts.temperatureFactor2,
		stats: { max: _.iMax, min: _.iMin, avg: _.iAvg },
		scale: {
			axis       : "y",
			position   : "right",
			scale      : { min: -1, max: +1 },
			tickPadding: 16,
			labelX     : +18,
			labelY     : -8,
		},
		path: {
			label        : temperature.label,
			yAttr        : 'temperature',
			scaleX       : 'distance',
			scaleY       : 'temperature',
			color        : 'transparent',
			strokeColor  : '#000',
			strokeOpacity: "0.85",
			// fillOpacity  : "0.1",
		},
		tooltip: {
			name: 'temperature',
			chart: (item) => L._("Temp: ") + Math.round(item.temperature).toLocaleString() + " " +  temperature.label,
			marker: (item) => Math.round(item.temperature).toLocaleString() + " "  + temperature.label,
			order: 1
		},
		summary: {
			"mintemp": {
				label: "Min Temp: ",
				value: (track, unit) =>  Math.round(track.temperature_min || 0) + '&nbsp;' + unit,
			},
			"maxtemp": {
				label: "Max Temp: ",
				value: (track, unit) => Math.round(track.temperature_max || 0) + '&nbsp;' + unit,
			},
			"avgtemp": {
				label: "Avg Temp: ",
				value: (track, unit) => Math.round(track.temperature_avg || 0) + '&nbsp;' + unit,
			},
		}
	};
}
