export function Cadence() {

    const _ = L.Control.Elevation.Utils;

    return {
        name: 'cadence',    // <-- Your custom option name (eg. "cadence: true")
        unit: 'rpm',
        meta: 'cad',        // <-- point.meta.cad
        coordinateProperties: ["cads", "cadences", "cad", "cadence"], // List of GPX Extensions ("coordinateProperties") to be handled by "@tmcw/toGeoJSON"
        pointToAttr: (point, i) => (point.cad ?? point.meta.cad ?? point.prev('cadence')) || 0,
		stats: { max: _.iMax, min: _.iMin, avg: _.iAvg },
        scale: {
            axis       : "y",
            position   : "right",
            scale      : { min: -1, max: +1 },
            tickPadding: 16,
            labelX     : 25,
            labelY     : -8,
        },
        path: {
            label        : 'RPM',
            yAttr        : 'cadence',
            scaleX       : 'distance',
            scaleY       : 'cadence',
            color        : '#FFF',
            strokeColor  : 'blue',
            strokeOpacity: "0.85",
            fillOpacity  : "0.1",
        },
        tooltip: {
            name: 'cadence',
            chart: (item) => L._("cad: ") + item.cadence + " " +  'rpm',
            marker: (item) => Math.round(item.cadence) + " "  + 'rpm',
            order: 1
        },
        summary: {
            "minrpm": {
                label: "Min RPM: ",
                value: (track, unit) =>  Math.round(track.cadence_min || 0) + '&nbsp;' + unit,
                // order: 30
            },
            "maxrpm": {
                label: "Max RPM: ",
                value: (track, unit) => Math.round(track.cadence_max || 0) + '&nbsp;' + unit,
                // order: 30
            },
            "avgrpm": {
                label: "Avg RPM: ",
                value: (track, unit) => Math.round(track.cadence_avg || 0) + '&nbsp;' + unit,
                // order: 20
            },
        }
    };
}