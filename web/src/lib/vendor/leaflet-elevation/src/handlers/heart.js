export function Heart() {

    const _ = L.Control.Elevation.Utils;

    return {
        name: 'heart',      // <-- Your custom option name (eg. "heart: true")
        unit: 'bpm',
        meta: 'hr',         // <-- point.meta.hr
        coordinateProperties: ["heart", "heartRates", "heartRate"], // List of GPX Extensions ("coordinateProperties") to be handled by "@tmcw/toGeoJSON"
        pointToAttr: (point, i) => (point.hr ?? point.meta.hr ?? point.prev('heart')) || 0,
		stats: { max: _.iMax, min: _.iMin, avg: _.iAvg },
        scale: {
            axis       : "y",
            position   : "left",
            scale      : { min: -1, max: +1 },
            tickPadding: 25,
            labelX     : -30,
            labelY     : -8,
        },
        path: {
            label        : 'ECG',
            yAttr        : 'heart',
            scaleX       : 'distance',
            scaleY       : 'heart',
            color        : 'white',
            strokeColor  : 'red',
            strokeOpacity: "0.85",
            fillOpacity  : "0.1",
        },
        tooltip: {
            chart: (item) => L._("hr: ") + item.heart + " " + 'bpm',
            marker: (item) => Math.round(item.heart) + " " + 'bpm',
            order: 1
        },
        summary: {
            "minbpm": {
                label: "Min BPM: ",
                value: (track, unit) => Math.round(track.heart_min || 0) + '&nbsp;' + unit,
                // order: 30
            },
            "maxbpm": {
                label: "Max BPM: ",
                value: (track, unit) => Math.round(track.heart_max || 0) + '&nbsp;' + unit,
                // order: 30
            },
            "avgbpm": {
                label: "Avg BPM: ",
                value: (track, unit) => Math.round(track.heart_avg || 0) + '&nbsp;' + unit,
                // order: 20
            },
        }
    };
};