/**
 * TODO: exget computed styles of theese values from actual "CSS vars"
 **/
export const Colors = {
	'lightblue': { area: '#3366CC', alpha: 0.45, stroke: '#3366CC' },
	'magenta'  : { area: '#FF005E' },
	'yellow'   : { area: '#FF0' },
	'purple'   : { area: '#732C7B' },
	'steelblue': { area: '#4682B4' },
	'red'      : { area: '#F00' },
	'lime'     : { area: '#9CC222', line: '#566B13' }
};

const SEC  = 1000;
const MIN  = SEC * 60;
const HOUR = MIN * 60;
const DAY  = HOUR * 24;

export function resolveURL(src, baseUrl) {
	return (new URL(src, (src.startsWith('../') || src.startsWith('./')) ? baseUrl : undefined)).toString()
};

/**
 * Convert a time (millis) to a human readable duration string (%Dd %H:%M'%S")
 */
export function formatTime(t) {
	let d = Math.floor(t / DAY);
	let h = Math.floor( (t - d * DAY) / HOUR);
	let m = Math.floor( (t - d * DAY - h * HOUR) / MIN);
	let s = Math.round( (t - d * DAY - h * HOUR - m * MIN) / SEC);
	if ( s === 60 ) { m++; s = 0; }
	if ( m === 60 ) { h++; m = 0; }
	if ( h === 24 ) { d++; h = 0; }
	return (d ? d + "d " : '') + h.toString().padStart(2, 0) + ':' + m.toString().padStart(2, 0) + "'" + s.toString().padStart(2, 0) + '"';
}

/**
 * Convert a time (millis) to human readable date string (dd-mm-yyyy hh:mm:ss)
 */
 export function formatDate(format) {
	if (!format) {
		return (time) => (new Date(time)).toLocaleString().replaceAll('/', '-').replaceAll(',', ' ');
	} else if (format == 'time') {
		return (time) => (new Date(time)).toLocaleTimeString();
	} else if (format == 'date') {
		return (time) => (new Date(time)).toLocaleDateString();
	}
	return (time) => format(time);
}

/**
 * Generate download data event.
 */
 export function saveFile(dataURI, fileName) {
	let a = create('a', '', { href: dataURI, target: '_new', download: fileName || "", style: "display:none;" });
	let b = document.body;
	b.appendChild(a);
	a.click();
	b.removeChild(a);
}


/**
 * Convert SVG Path into Path2D and then update canvas
 */
 export function drawCanvas(ctx, path) {
	path.classed('canvas-path', true);

	ctx.beginPath();
	ctx.moveTo(0, 0);
	let p = new Path2D(path.attr('d'));

	ctx.strokeStyle = path.__strokeStyle || path.attr('stroke');
	ctx.fillStyle   = path.__fillStyle   || path.attr('fill');
	ctx.lineWidth   = 1.25;
	ctx.globalCompositeOperation = 'source-over';

	// stroke opacity
	ctx.globalAlpha = path.attr('stroke-opacity') || 0.3;
	ctx.stroke(p);

	// fill opacity
	ctx.globalAlpha = path.attr('fill-opacity')   || 0.45;
	ctx.fill(p);

	ctx.globalAlpha = 1;

	ctx.closePath();
}

/**
 * Loop and extract GPX Extensions handled by "@tmcw/toGeoJSON" (eg. "coordinateProperties" > "times")
 */
export function coordPropsToMeta(coordProps, name, parser) {
	return coordProps && (({props, point, id, isMulti }) => {
		if (props) {
			for (const key of coordProps) {
				if (key in props) {
					point.meta[name] = (parser || parseNumeric).call(this, (isMulti ? props[key][isMulti] : props[key]), id);
					break;
				}
			}
		}
	});
}

/**
 * Extract numeric property (id) from GeoJSON object
 */
export const parseNumeric  = (property, id) => parseInt((typeof property === 'object' ? property[id] : property));

/**
 * Extract datetime property (id) from GeoJSON object
 */
export const parseDate = (property, id) => new Date(Date.parse((typeof property === 'object' ? property[id] : property)));

/**
 * A little bit shorter than L.DomUtil
 */
export const addClass      = (n, str)             => n && str.split(" ").every(s => s && L.DomUtil.addClass(n, s));
export const removeClass   = (n, str)             => n && str.split(" ").every(s => s && L.DomUtil.removeClass(n, s));
export const toggleClass   = (n, str, cond)       => (cond ? addClass : removeClass)(n, str);
export const replaceClass  = (n, rem, add)        => (rem && removeClass(n, rem)) || (add && addClass(n, add));
export const style         = (n, k, v)            => (typeof v === "undefined" && L.DomUtil.getStyle(n, k)) || n.style.setProperty(k, v);
export const toggleStyle   = (n, k, v, cond)      => style(n, k, cond ? v : '');
export const setAttributes = (n, attrs)           => { for (let k in attrs) { n.setAttribute(k, attrs[k]); } };
export const toggleEvent   = (el, e, fn, cond)    => el[cond ? 'on' : 'off'](e, fn);
export const create        = (tag, str, attrs, n) => { let elem = L.DomUtil.create(tag, str || ""); if (attrs) setAttributes(elem, attrs); if (n) append(n, elem); return elem; };
export const append        = (n, c)               => n.appendChild(c);
export const insert        = (n, c, pos)          => n.insertAdjacentElement(pos, c);
export const select        = (str, n)             => (n || document).querySelector(str);
export const each          = (obj, fn)            => { for (let i in obj) fn(obj[i], i); };
export const randomId      = ()                   => Math.random().toString(36).substr(2, 9);

/**
 * TODO: use generators instead? (ie. "yield")
 */
export const iMax = (iVal, max = -Infinity) => (iVal > max ? iVal : max); 
export const iMin = (iVal, min = +Infinity) => (iVal < min ? iVal : min);
export const iAvg = (iVal, avg = 0, idx = 1) => (iVal + avg * (idx - 1)) / idx;
export const iSum = (iVal, sum = 0) => iVal + sum;

/**
 * Alias for some leaflet core functions
 */
export const { on, off }           = L.DomEvent;
export const { throttle, wrapNum } = L.Util;
export const { hasClass }          = L.DomUtil;

/**
 * Limit floating point precision
 */
export const round        = L.Util.formatNum;

/**
 * Limit a number between min / max values
 */
export const clamp     = (val, range)           => range ? (val < range[0] ? range[0] : val > range[1] ? range[1] : val) : val;

/**
 * Limit a delta difference between two values
 */
export const wrapDelta = (curr, prev, deltaMax) => Math.abs(curr - prev) > deltaMax ? prev + deltaMax * Math.sign(curr - prev) : curr;

/**
 * A deep copy implementation that takes care of correct prototype chain and cycles, references
 * 
 * @see https://web.dev/structured-clone/#features-and-limitations
 */
export function cloneDeep(o, skipProps = [], cache = []) {
	switch(!o || typeof o) {
		case 'object':
			const hit = cache.filter(c => o === c.original)[0];
			if (hit) return hit.copy;                             // handle circular structures
			const copy = Array.isArray(o) ? [] : Object.create(Object.getPrototypeOf(o));
			cache.push({ original: o, copy });
			Object
				.getOwnPropertyNames(o)
				.forEach(function (prop) {
					const propdesc = Object.getOwnPropertyDescriptor(o, prop);
					Object.defineProperty(
						copy,
						prop,
						propdesc.get || propdesc.set
							? propdesc                                    // just copy accessor properties
							: {                                           // deep copy data properties
								writable:     propdesc.writable,
								configurable: propdesc.configurable,
								enumerable:   propdesc.enumerable,
								value:        skipProps.includes(prop) ? propdesc.value : cloneDeep(propdesc.value, skipProps, cache),
							}
					);
				});
			return copy;
		case 'function':
		case 'symbol':
			console.warn('cloneDeep: ' + typeof o + 's not fully supported:', o);
		case true:
			// null, undefined or falsy primitive
		default:
			return o;
	}
}