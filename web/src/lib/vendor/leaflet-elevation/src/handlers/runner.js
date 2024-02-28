/**
 * @see https://github.com/Igor-Vladyka/leaflet.motion
 * 
 * @example
 * ```js
 * L.control.Elevation({ handlers: [ 'Runner' ], runnerOptions: { polyline: {..}, motion: {..}, marker {..} } })
 * ```
 */
export async function Runner() {

  await this.import(this.__LMOTION || 'https://unpkg.com/leaflet.motion@0.3.2/dist/leaflet.motion.min.js', typeof L.Motion !== 'object')

  let { runnerOptions } = this.options;

  runnerOptions = L.extend(
    { polyline: {}, motion: {}, marker: undefined },
    'object' === typeof runnerOptions ? runnerOptions : {}
  );

  // Custom tooltips
  this._registerTooltip({
    name: 'here',
    marker: (item) => L._("You are here: "),
    order: 1,
  });

  this._registerTooltip({
    name: 'distance',
    marker: (item) => Math.round(item.dist) + " " + this.options.xLabel,
    order: 2,
  });

  this.addCheckpoint = function (checkpoint) {
    return this._registerCheckPoint({ // <-- NB these are private functions use them at your own risk!
      latlng: this._findItemForX(this._x(checkpoint.dist)).latlng,
      label: checkpoint.label || ''
    });
  }

  this.addRunner = function (runner) {
    let x = this._x(runner.dist);
    let y = this._y(this._findItemForX(x).z)
    let g = d3.select(this._container)
      .select('svg > g')
      .append("svg:circle")
      .attr("class", "runner " + this.options.theme + " height-focus circle-lower")
      .attr("r", 6)
      .attr("cx", x)
      .attr("cy", y);
    return g;
  }

  this.setPositionFromLatLng = function (latlng) {
    this._onMouseMoveLayer({ latlng: latlng }); // Update map and chart "markers" from latlng
  };

  this.tick = function (runner, dist = 0, inc = 0.1) {
    dist = (dist <= this.track_info.distance - inc) ? dist + inc : 0;
    this.updateRunnerPos(runner, dist);
    setTimeout(() => this.tick(runner, dist), 150);
  };

  this.updateRunnerPos = function (runner, pos) {
    let curr, x;

    if (pos instanceof L.LatLng) {
      curr = this._findItemForLatLng(pos);
      x = this._x(curr.dist);
    } else {
      x    = this._x(pos);
      curr = this._findItemForX(x);
    }

    runner
      .attr("cx", x)
      .attr("cy", this._y(curr.z));
    this.setPositionFromLatLng(curr.latlng);
  };

  this.animate = function (layer, speed = 1500) {

    if (this._runner) {
      this._runner.remove();
    }

    layer.setStyle({ opacity: 0.5 });

    const geo = L.geoJson(layer.toGeoJSON(), { coordsToLatLng: (coords) => L.latLng(coords[0], coords[1], coords[2] * (this.options.altitudeFactor || 1)) });
    this._runner = L.motion.polyline(
      geo.toGeoJSON().features[0].geometry.coordinates, 
      L.extend({}, { color: 'red', pane: 'elevationPane', attribution: '' }, runnerOptions.polyline), 
      L.extend({}, { auto: true, speed: speed, }, runnerOptions.motion),
      runnerOptions.marker || undefined
    );

    // Override default function behavior: `L.Motion.Polyline::_drawMarker()` 
    this._runner._drawMarker = new Proxy(this._runner._drawMarker, {
      apply: (target, thisArg, argArray) => {
        thisArg._runner = thisArg._runner || this.addRunner({ dist: 0 });
        this.updateRunnerPos(thisArg._runner, argArray[0]);
        return target.apply(thisArg, argArray);
      }
    });

    this._runner.addTo(this._map);
  };

  return {};
}
