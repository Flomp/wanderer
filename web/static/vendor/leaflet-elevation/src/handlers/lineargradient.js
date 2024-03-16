/**
 * @see https://github.com/Raruto/leaflet-elevation/issues/251
 * 
 * @example
 * ```js
 * L.control.Elevation({
 *   altitude: true,
 *   distance: true,
 *   handlers: [ 'Altitude', 'Distance', 'LinearGradient', ],
 *   linearGradient: {
 *     attr: 'z',
 *     path: 'altitude',
 *     range: { 0.0: '#008800', 0.5: '#ffff00', 1.0: '#ff0000' },
 *     min: 'elevation_min',
 *     max: 'elevation_max',
 *   },
 * })
 * ```
 */
export function LinearGradient() {

  if (!this.options.linearGradient) {
    return {};
  }

  const _ = L.Control.Elevation.Utils;

  /**
   * Initialize gradient color palette.
   */
  const get_palette = function ({range, min, max, depth = 256}) {
    const canvas = document.createElement('canvas'),
      ctx = canvas.getContext('2d'),
      gradient = ctx.createLinearGradient(0, 0, 0, depth);

    canvas.width = 1;
    canvas.height = depth;

    for (let i in range) {
      gradient.addColorStop(i, range[i]);
    }

    ctx.fillStyle = gradient;
    ctx.fillRect(0, 0, 1, depth);

    const { data } = ctx.getImageData(0, 0, 1, depth);

    return {
      /**
       * Gets the RGB values of a given z value of the current palette.
       * 
       * @param {number} value - Value to get the color for, should be between min and max.
       * @returns {string} The RGB values as `rgb(r, g, b)` string
       */
      getRGBColor(value) {
        const idx = Math.floor(Math.min(Math.max((value - min) / (max - min), 0), 0.999) * depth) * 4;
        return 'rgb(' + [data[idx], data[idx + 1], data[idx + 2]].join(',') + ')';
      }
    };
  };

  const { preferCanvas } = this.options;
  const { attr, path: path_name, range, min, max } = L.extend({
    attr: 'z',
    path: 'altitude',
    range: { 0.0: '#008800', 0.5: '#ffff00', 1.0: '#ff0000' },
    min: 'elevation_min',
    max: 'elevation_max',
  }, (true === this.options.linearGradient) ? {} : this.options.linearGradient);

  const gradient_id = path_name + '-gradient-' + _.randomId();
  const legend_id   = 'legend-' + gradient_id;

  // Charte profile gradient 
  this.on('elechart_axis', () => {
    if (!this._data.length) return;

    const chart = this._chart;
    const path = chart._paths[path_name];
    const { defs } = chart._chart.utils;

    const palette = get_palette({
      min: isFinite(this.track_info[min]) ? this.track_info[min] : 0,
      max: isFinite(this.track_info[max]) ? this.track_info[max] : 1,
      range,
    });

    let gradient;

    if (preferCanvas) {
      /** ref: `path.__fillStyle` within L.Control.Elevation.Utils::drawCanvas(ctx, path) */
      path.__fillStyle = gradient = chart._context.createLinearGradient(0, 0, chart._width(), 0);
    } else {
      defs.select('#' + gradient_id).remove();
      gradient = defs.append('svg:linearGradient').attr('id', gradient_id);
      gradient.addColorStop = function(offset, color) { gradient.append('svg:stop').attr('offset', offset).attr('stop-color', color) };
      path.attr('fill', 'url(#' + gradient_id + ')').classed('area', false);
    }

    // Generate gradient for each segment picking colors from palette
    for (let i = 0, data = this._data; i < data.length; i++) {
      gradient.addColorStop((i) / data.length, palette.getRGBColor(data[i][attr]));
    }

  });

  // Legend item gradient
  this.on('elechart_updated', () => {
    const chart = this._chart;
    const { defs } = chart._chart.utils;
    defs.select('#' + legend_id).remove();
    const legendGradient = defs.append('svg:linearGradient').attr('id', legend_id);
    Object.keys(range).sort().forEach(i => legendGradient.append('svg:stop').attr('offset', i).attr('stop-color', range[i]));

    chart._container
      .select('.legend-' + path_name + ' > rect')
      .attr('fill', 'url(#' + legend_id + ')')
      .classed('area', false);
  });

  return { };
}