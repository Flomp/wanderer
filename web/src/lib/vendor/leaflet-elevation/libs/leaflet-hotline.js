/*
(c) 2017, iosphere GmbH
Leaflet.hotline, a Leaflet plugin for drawing gradients along polylines.
https://github.com/iosphere/Leaflet.hotline/
*/

(function (root, plugin) {
    /**
     * UMD wrapper.
     * When used directly in the Browser it expects Leaflet to be globally
     * available as `L`. The plugin then adds itself to Leaflet.
     * When used as a CommonJS module (e.g. with browserify) only the plugin
     * factory gets exported, so one hast to call the factory manually and pass
     * Leaflet as the only parameter.
     * @see {@link https://github.com/umdjs/umd}
     */
    if (typeof define === 'function' && define.amd) {
        define(['leaflet'], plugin);
    } else if (typeof exports === 'object') {
        module.exports = plugin;
    } else {
        plugin(root.L);
    }
}(window, function (L) {
    // Plugin is already added to Leaflet
    if (L.Hotline) {
        return L;
    }

    /**
     * Core renderer.
     * @constructor
     * @param {HTMLElement | string} canvas - &lt;canvas> element or its id
     * to initialize the instance on.
     */
    var Hotline = function (canvas) {
        if (!(this instanceof Hotline)) { return new Hotline(canvas); }

        var defaultPalette = {
            0.0: 'green',
            0.5: 'yellow',
            1.0: 'red'
        };

        this._canvas = canvas = ('string' === typeof canvas) 
            ? document.getElementById(canvas)
            : canvas;

        this._ctx = canvas.getContext('2d');
        this._width = canvas.width;
        this._height = canvas.height;

        this._weight = 5;
        this._outlineWidth = 1;
        this._outlineColor = 'black';

        this._min = 0;
        this._max = 1;

        this._data = [];

        this.palette(defaultPalette);
    };

    Hotline.prototype = {
        /**
         * Sets the width of the canvas. Used when clearing the canvas.
         * @param {number} width - Width of the canvas.
         */
        width: function (width) {
            this._width = width;
            return this;
        },

        /**
         * Sets the height of the canvas. Used when clearing the canvas.
         * @param {number} height - Height of the canvas.
         */
        height: function (height) {
            this._height = height;
            return this;
        },

        /**
         * Sets the weight of the path.
         * @param {number} weight - Weight of the path in px.
         */
        weight: function (weight) {
            this._weight = weight;
            return this;
        },

        /**
         * Sets the width of the outline around the path.
         * @param {number} outlineWidth - Width of the outline in px.
         */
        outlineWidth: function (outlineWidth) {
            this._outlineWidth = outlineWidth;
            return this;
        },

        /**
         * Sets the color of the outline around the path.
         * @param {string} outlineColor - A CSS color value.
         */
        outlineColor: function (outlineColor) {
            this._outlineColor = outlineColor;
            return this;
        },

        /**
         * Sets the palette gradient.
         * @param {Object.<number, string>} palette  - Gradient definition.
         * e.g. { 0.0: 'white', 1.0: 'black' }
         */
        palette: function (palette) {
            var canvas = document.createElement('canvas'),
                ctx = canvas.getContext('2d'),
                gradient = ctx.createLinearGradient(0, 0, 0, 256);

            canvas.width = 1;
            canvas.height = 256;

            for (var i in palette) {
                gradient.addColorStop(i, palette[i]);
            }

            ctx.fillStyle = gradient;
            ctx.fillRect(0, 0, 1, 256);

            this._palette = ctx.getImageData(0, 0, 1, 256).data;

            return this;
        },

        /**
         * Sets the value used at the start of the palette gradient.
         * @param {number} min
         */
        min: function (min) {
            this._min = min;
            return this;
        },

        /**
         * Sets the value used at the end of the palette gradient.
         * @param {number} max
         */
        max: function (max) {
            this._max = max;
            return this;
        },

        /**
         * A path to rander as a hotline.
         * @typedef Array.<{x:number, y:number, z:number}> Path - Array of x, y and z coordinates.
         */

        /**
         * Sets the data that gets drawn on the canvas.
         * @param {(Path|Path[])} data - A single path or an array of paths.
         */
        data: function (data) {
            this._data = data;
            return this;
        },

        /**
         * Adds a path to the list of paths.
         * @param {Path} path
         */
        add: function (path) {
            this._data.push(path);
            return this;
        },

        /**
         * Draws the currently set paths.
         */
        draw: function () {
            var ctx = this._ctx;

            ctx.globalCompositeOperation = 'source-over';
            ctx.lineCap = 'round';

            this._drawOutline(ctx);
            this._drawHotline(ctx);

            return this;
        },

        /**
         * Gets the RGB values of a given z value of the current palette.
         * @param {number} value - Value to get the color for, should be between min and max.
         * @returns {Array.<number>} The RGB values as an array [r, g, b]
         */
        getRGBForValue: function (value) {
            var valueRelative = Math.min(Math.max((value - this._min) / (this._max - this._min), 0), 0.999);
            var paletteIndex = Math.floor(valueRelative * 256) * 4;

            return [
                this._palette[paletteIndex],
                this._palette[paletteIndex + 1],
                this._palette[paletteIndex + 2]
            ];
        },

        /**
         * Draws the outline of the graphs.
         * @private
         */
        _drawOutline: function (ctx) {
            var i, j, dataLength, path, pathLength, pointStart, pointEnd;

            if (this._outlineWidth) {
                for (i = 0, dataLength = this._data.length; i < dataLength; i++) {
                    path = this._data[i];
                    ctx.lineWidth = this._weight + 2 * this._outlineWidth;

                    for (j = 1, pathLength = path.length; j < pathLength; j++) {
                        pointStart = path[j - 1];
                        pointEnd = path[j];

                        ctx.strokeStyle = this._outlineColor;
                        ctx.beginPath();
                        ctx.moveTo(pointStart.x, pointStart.y);
                        ctx.lineTo(pointEnd.x, pointEnd.y);
                        ctx.stroke();
                    }
                }
            }
        },

        /**
         * Draws the color encoded hotline of the graphs.
         * @private
         */
        _drawHotline: function (ctx) {
            var i, j, dataLength, path, pathLength, pointStart, pointEnd,
                gradient, gradientStartRGB, gradientEndRGB;

            ctx.lineWidth = this._weight;

            for (i = 0, dataLength = this._data.length; i < dataLength; i++) {
                path = this._data[i];

                for (j = 1, pathLength = path.length; j < pathLength; j++) {
                    pointStart = path[j - 1];
                    pointEnd = path[j];

                    // Create a gradient for each segment, pick start end end colors from palette gradient
                    gradient = ctx.createLinearGradient(pointStart.x, pointStart.y, pointEnd.x, pointEnd.y);
                    gradientStartRGB = this.getRGBForValue(pointStart.z);
                    gradientEndRGB = this.getRGBForValue(pointEnd.z);
                    gradient.addColorStop(0, 'rgb(' + gradientStartRGB.join(',') + ')');
                    gradient.addColorStop(1, 'rgb(' + gradientEndRGB.join(',') + ')');

                    ctx.strokeStyle = gradient;
                    ctx.beginPath();
                    ctx.moveTo(pointStart.x, pointStart.y);
                    ctx.lineTo(pointEnd.x, pointEnd.y);
                    ctx.stroke();
                }
            }
        }
    };


    var Renderer = L.Canvas.extend({
        _initContainer: function () {
            L.Canvas.prototype._initContainer.call(this);
            this._hotline = new Hotline(this._container);
        },

        _update: function () {
            L.Canvas.prototype._update.call(this);
            this._hotline.width(this._container.width);
            this._hotline.height(this._container.height);
        },

        _updatePoly: function (layer) {
            if (!this._drawing) { return; }

            var parts = layer._parts;

            if (!parts.length) { return; }

            this._updateOptions(layer);

            this._hotline
                .data(parts)
                .draw();
        },

        _updateOptions: function (layer) {
            if (layer.options.min != null) {
                this._hotline.min(layer.options.min);
            }
            if (layer.options.max != null) {
                this._hotline.max(layer.options.max);
            }
            if (layer.options.weight != null) {
                this._hotline.weight(layer.options.weight);
            }
            if (layer.options.outlineWidth != null) {
                this._hotline.outlineWidth(layer.options.outlineWidth);
            }
            if (layer.options.outlineColor != null) {
                this._hotline.outlineColor(layer.options.outlineColor);
            }
            if (layer.options.palette) {
                this._hotline.palette(layer.options.palette);
            }
        }
    });

    var renderer = function (options) {
        return L.Browser.canvas ? new Renderer(options) : null;
    };


    var Util = {
        /**
         * This is just a copy of the original Leaflet version that support a third z coordinate.
         * @see {@link http://leafletjs.com/reference.html#lineutil-clipsegment|Leaflet}
         */
        clipSegment: function (a, b, bounds, useLastCode, round) {
            var codeA = useLastCode ? this._lastCode : L.LineUtil._getBitCode(a, bounds),
                codeB = L.LineUtil._getBitCode(b, bounds),
                codeOut, p, newCode;

            // save 2nd code to avoid calculating it on the next segment
            this._lastCode = codeB;

            while (true) {
                // if a,b is inside the clip window (trivial accept)
                if (!(codeA | codeB)) {
                    return [a, b];
                    // if a,b is outside the clip window (trivial reject)
                } else if (codeA & codeB) {
                    return false;
                    // other cases
                } else {
                    codeOut = codeA || codeB;
                    p = L.LineUtil._getEdgeIntersection(a, b, codeOut, bounds, round);
                    newCode = L.LineUtil._getBitCode(p, bounds);

                    if (codeOut === codeA) {
                        p.z = a.z;
                        a = p;
                        codeA = newCode;
                    } else {
                        p.z = b.z;
                        b = p;
                        codeB = newCode;
                    }
                }
            }
        }
    };


    L.Hotline = L.Polyline.extend({
        statics: {
            Renderer: Renderer,
            renderer: renderer
        },

        options: {
            renderer: renderer(),
            min: 0,
            max: 1,
            palette: {
                0.0: 'green',
                0.5: 'yellow',
                1.0: 'red'
            },
            weight: 5,
            outlineColor: 'black',
            outlineWidth: 1
        },

        getRGBForValue: function (value) {
            return this._renderer._hotline.getRGBForValue(value);
        },

        /**
         * Just like the Leaflet version, but with support for a z coordinate.
         */
        _projectLatlngs: function (latlngs, result, projectedBounds) {
            var flat = latlngs[0] instanceof L.LatLng,
                len = latlngs.length,
                i, ring;

            if (flat) {
                ring = [];
                for (i = 0; i < len; i++) {
                    ring[i] = this._map.latLngToLayerPoint(latlngs[i]);
                    // Add the altitude of the latLng as the z coordinate to the point
                    ring[i].z = latlngs[i].alt;
                    projectedBounds.extend(ring[i]);
                }
                result.push(ring);
            } else {
                for (i = 0; i < len; i++) {
                    this._projectLatlngs(latlngs[i], result, projectedBounds);
                }
            }
        },

        /**
         * Just like the Leaflet version, but uses `Util.clipSegment()`.
         */
        _clipPoints: function () {
            if (this.options.noClip) {
                this._parts = this._rings;
                return;
            }

            this._parts = [];

            var parts = this._parts,
                bounds = this._renderer._bounds,
                i, j, k, len, len2, segment, points;

            for (i = 0, k = 0, len = this._rings.length; i < len; i++) {
                points = this._rings[i];

                for (j = 0, len2 = points.length; j < len2 - 1; j++) {
                    segment = Util.clipSegment(points[j], points[j + 1], bounds, j, true);

                    if (!segment) { continue; }

                    parts[k] = parts[k] || [];
                    parts[k].push(segment[0]);

                    // if segment goes out of screen, or it's the last one, it's the end of the line part
                    if ((segment[1] !== points[j + 1]) || (j === len2 - 2)) {
                        parts[k].push(segment[1]);
                        k++;
                    }
                }
            }
        },

        _clickTolerance: function () {
            return this.options.weight / 2 + this.options.outlineWidth + (L.Browser.touch ? 10 : 0);
        }
    });

    L.hotline = function (latlngs, options) {
        return new L.Hotline(latlngs, options);
    };


    return L;
}));