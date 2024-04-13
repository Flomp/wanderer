import QRCode from "qrcode";

'use strict';

/*! *****************************************************************************
Copyright (c) Microsoft Corporation.

Permission to use, copy, modify, and/or distribute this software for any
purpose with or without fee is hereby granted.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
PERFORMANCE OF THIS SOFTWARE.
***************************************************************************** */

function __awaiter(thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
}

function __generator(thisArg, body) {
    var _ = { label: 0, sent: function () { if (t[0] & 1) throw t[1]; return t[1]; }, trys: [], ops: [] }, f, y, t, g;
    return g = { next: verb(0), "throw": verb(1), "return": verb(2) }, typeof Symbol === "function" && (g[Symbol.iterator] = function () { return this; }), g;
    function verb(n) { return function (v) { return step([n, v]); }; }
    function step(op) {
        if (f) throw new TypeError("Generator is already executing.");
        while (_) try {
            if (f = 1, y && (t = op[0] & 2 ? y["return"] : op[0] ? y["throw"] || ((t = y["return"]) && t.call(y), 0) : y.next) && !(t = t.call(y, op[1])).done) return t;
            if (y = 0, t) op = [op[0] & 2, t.value];
            switch (op[0]) {
                case 0: case 1: t = op; break;
                case 4: _.label++; return { value: op[1], done: false };
                case 5: _.label++; y = op[1]; op = [0]; continue;
                case 7: op = _.ops.pop(); _.trys.pop(); continue;
                default:
                    if (!(t = _.trys, t = t.length > 0 && t[t.length - 1]) && (op[0] === 6 || op[0] === 2)) { _ = 0; continue; }
                    if (op[0] === 3 && (!t || (op[1] > t[0] && op[1] < t[3]))) { _.label = op[1]; break; }
                    if (op[0] === 6 && _.label < t[1]) { _.label = t[1]; t = op; break; }
                    if (t && _.label < t[2]) { _.label = t[2]; _.ops.push(op); break; }
                    if (t[2]) _.ops.pop();
                    _.trys.pop(); continue;
            }
            op = body.call(thisArg, _);
        } catch (e) { op = [6, e]; y = 0; } finally { f = t = 0; }
        if (op[0] & 5) throw op[1]; return { value: op[0] ? op[1] : void 0, done: true };
    }
}

/*
 * @Author: super
 * @Date: 2019-06-27 16:29:43
 * @Last Modified by: super
 * @Last Modified time: 2019-06-27 17:46:21
 */
/**
 * promisify promise化，使得promisify(func).then()更加方便，不用每次都構造 promise
 * Making Promise more convenient, without having to construct a promise every time
 * @param f {function} 異步函數
 */
var promisify = function (f) {
    return function () {
        var args = Array.prototype.slice.call(arguments);
        return new Promise(function (resolve, reject) {
            args.push(function (err, result) {
                if (err)
                    reject(err);
                else
                    resolve(result);
            });
            f.apply(null, args);
        });
    };
};
/**
 * 判斷是否是函數
 * Determine if it is a function
 * @param o {function} 函數
 */
function isFunction(o) {
    return typeof o === "function";
}
/**
 * 判斷是不是字符串
 * Determine if it is a string
 * @param o {string} 字符串
 */
function isString(o) {
    return typeof o === "string";
}

// @ts-ignore
// import QRCode from "qrcode"
var toCanvas$1 = promisify(QRCode.toCanvas);
var renderQrCode = function (_a) {
    var canvas = _a.canvas, content = _a.content, _b = _a.width, width = _b === void 0 ? 0 : _b, _c = _a.nodeQrCodeOptions, nodeQrCodeOptions = _c === void 0 ? {} : _c;
    // 容错率，默认对内容少的二维码采用高容错率，内容多的二维码采用低容错率
    // according to the content length to choose different errorCorrectionLevel
    nodeQrCodeOptions.errorCorrectionLevel =
        nodeQrCodeOptions.errorCorrectionLevel || getErrorCorrectionLevel(content);
    return getOriginWidth(content, nodeQrCodeOptions).then(function (_width) {
        // 得到原始比例后还原至设定值，再放大4倍以获取高清图
        // Restore to the set value according to the original ratio, and then zoom in 4 times to get the HD image.
        nodeQrCodeOptions.scale = width === 0 ? undefined : (width / _width) * 4;
        // @ts-ignore
        return toCanvas$1(canvas, content, nodeQrCodeOptions);
    });
};
// 得到原QrCode的大小，以便缩放得到正确的QrCode大小
// Get the size of the original QrCode
var getOriginWidth = function (content, nodeQrCodeOption) {
    var _canvas = document.createElement("canvas");
    // @ts-ignore
    return toCanvas$1(_canvas, content, nodeQrCodeOption).then(function () { return _canvas.width; });
};
// 对于内容少的QrCode，增大容错率
// Increase the fault tolerance for QrCode with less content
var getErrorCorrectionLevel = function (content) {
    if (content.length > 36) {
        return "M";
    }
    else if (content.length > 16) {
        return "Q";
    }
    else {
        return "H";
    }
};

var drawLogo = function (_a) {
    var canvas = _a.canvas, logo = _a.logo;
    if (!logo)
        return Promise.resolve();
    if (logo === '')
        return Promise.resolve();
    var canvasWidth = canvas.width;
    if (isString(logo)) {
        logo = { src: logo };
    }
    var _b = logo, _c = _b.logoSize, logoSize = _c === void 0 ? 0.15 : _c, _d = _b.borderColor, borderColor = _d === void 0 ? "#ffffff" : _d, _e = _b.bgColor, bgColor = _e === void 0 ? borderColor || "#ffffff" : _e, _f = _b.borderSize, borderSize = _f === void 0 ? 0.05 : _f, crossOrigin = _b.crossOrigin, _g = _b.borderRadius, borderRadius = _g === void 0 ? 8 : _g, _h = _b.logoRadius, logoRadius = _h === void 0 ? 0 : _h;
    var logoSrc = typeof logo === "string" ? logo : logo.src;
    var logoWidth = canvasWidth * logoSize;
    var logoXY = (canvasWidth * (1 - logoSize)) / 2;
    var logoBgWidth = canvasWidth * (logoSize + borderSize);
    var logoBgXY = (canvasWidth * (1 - logoSize - borderSize)) / 2;
    var ctx = canvas.getContext("2d");
    // logo 底色, draw logo background color
    canvasRoundRect(ctx)(logoBgXY, logoBgXY, logoBgWidth, logoBgWidth, borderRadius);
    ctx.fillStyle = bgColor;
    ctx.fill();
    // logo
    var image = new Image();
    image.setAttribute("crossOrigin", crossOrigin || "anonymous");
    image.src = logoSrc;
    // 使用image绘制可以避免某些跨域情况
    // Use image drawing to avoid some cross-domain situations
    var drawLogoWithImage = function (image) {
        ctx.drawImage(image, logoXY, logoXY, logoWidth, logoWidth);
    };
    // 使用canvas绘制以获得更多的功能
    // Use canvas to draw more features, such as borderRadius
    var drawLogoWithCanvas = function (image) {
        var canvasImage = document.createElement("canvas");
        canvasImage.width = logoXY + logoWidth;
        canvasImage.height = logoXY + logoWidth;
        canvasImage
            .getContext("2d")
            .drawImage(image, logoXY, logoXY, logoWidth, logoWidth);
        canvasRoundRect(ctx)(logoXY, logoXY, logoWidth, logoWidth, logoRadius);
        // @ts-ignore
        ctx.fillStyle = ctx.createPattern(canvasImage, "no-repeat");
        ctx.fill();
    };
    // 将 logo绘制到 canvas上
    // Draw the logo on the canvas
    return new Promise(function (resolve, reject) {
        image.onload = function () {
            logoRadius ? drawLogoWithCanvas(image) : drawLogoWithImage(image);
            resolve();
        };
        image.onerror = function () {
            reject('logo load fail!');
        };
    });
};
// draw radius
var canvasRoundRect = function (ctx) {
    return function (x, y, w, h, r) {
        var minSize = Math.min(w, h);
        if (r > minSize / 2) {
            r = minSize / 2;
        }
        ctx.beginPath();
        ctx.moveTo(x + r, y);
        ctx.arcTo(x + w, y, x + w, y + h, r);
        ctx.arcTo(x + w, y + h, x, y + h, r);
        ctx.arcTo(x, y + h, x, y, r);
        ctx.arcTo(x, y, x + w, y, r);
        ctx.closePath();
        return ctx;
    };
};

/*
 * @Author: super
 * @Date: 2019-06-27 16:29:34
 * @Last Modified by: super
 * @Last Modified time: 2019-06-27 16:47:22
 */
var toCanvas = function (options) {
    return renderQrCode(options).then(function () { return drawLogo(options); });
};

var toImage = function (options, instance) {
    return __awaiter(this, void 0, void 0, function () {
        var canvas, image, _a, downloadName, download, startDownload;
        return __generator(this, function (_b) {
            switch (_b.label) {
                case 0:
                    canvas = options.canvas;
                    if (options.logo) {
                        if (isString(options.logo)) {
                            options.logo = { src: options.logo };
                        }
                        options.logo.crossOrigin = 'Anonymous';
                    }
                    if (!!instance.ifCanvasDrawed) return [3 /*break*/, 2];
                    return [4 /*yield*/, toCanvas(options)];
                case 1:
                    _b.sent();
                    _b.label = 2;
                case 2:
                    image = options.image, _a = options.downloadName, downloadName = _a === void 0 ? 'qr-code' : _a;
                    download = options.download;
                    if (canvas.toDataURL()) {
                        image.src = canvas.toDataURL();
                    }
                    else {
                        throw new Error('Can not get the canvas DataURL');
                    }
                    instance.ifImageCreated = true;
                    if (download !== true && !isFunction(download)) {
                        return [2 /*return*/];
                    }
                    download = download === true ? function (start) { return start(); } : download;
                    startDownload = function () {
                        return saveImage(image, downloadName);
                    };
                    if (download) {
                        return [2 /*return*/, download(startDownload)];
                    }
                    return [2 /*return*/, Promise.resolve()];
            }
        });
    });
};
/**save image */
var saveImage = function (image, name) {
    return new Promise(function (resolve, reject) {
        try {
            var dataURL = image.src;
            var link = document.createElement('a');
            link.download = name;
            link.href = dataURL;
            link.dispatchEvent(new MouseEvent('click'));
            resolve(true);
        }
        catch (err) {
            reject(err);
        }
    });
};

var version = "1.0.5";

/*
 * @Author: super
 * @Date: 2019-06-27 16:29:31
 * @Last Modified by: suporka
 * @Last Modified time: 2020-03-04 12:24:50
 */
var QrCodeWithLogo = /** @class */ (function () {
    function QrCodeWithLogo(option) {
        this.ifCanvasDrawed = false;
        this.ifImageCreated = false;
        this.defaultOption = {
            canvas: undefined,
            image: undefined,
            content: ''
        };
        this.option = Object.assign(this.defaultOption, option);
        if (!this.option.canvas)
            this.option.canvas = document.createElement("canvas");
        if (!this.option.image)
            this.option.image = document.createElement("img");
        this.toCanvas().then(this.toImage.bind(this));
    }
    QrCodeWithLogo.prototype.toCanvas = function () {
        var _this = this;
        return toCanvas.call(this, this.option).then(function () {
            _this.ifCanvasDrawed = true;
            return Promise.resolve();
        });
    };
    QrCodeWithLogo.prototype.toImage = function () {
        return toImage(this.option, this);
    };
    QrCodeWithLogo.prototype.downloadImage = function (name) {
        if (name === void 0) { name = 'qrcode.png'; }
        return __awaiter(this, void 0, void 0, function () {
            return __generator(this, function (_a) {
                switch (_a.label) {
                    case 0:
                        if (!!this.ifImageCreated) return [3 /*break*/, 2];
                        return [4 /*yield*/, this.toImage()];
                    case 1:
                        _a.sent();
                        _a.label = 2;
                    case 2: return [2 /*return*/, saveImage(this.option.image, name)];
                }
            });
        });
    };
    QrCodeWithLogo.prototype.getImage = function () {
        return __awaiter(this, void 0, Promise, function () {
            return __generator(this, function (_a) {
                switch (_a.label) {
                    case 0:
                        if (!!this.ifImageCreated) return [3 /*break*/, 2];
                        return [4 /*yield*/, this.toImage()];
                    case 1:
                        _a.sent();
                        _a.label = 2;
                    case 2: return [2 /*return*/, this.option.image];
                }
            });
        });
    };
    QrCodeWithLogo.prototype.getCanvas = function () {
        return __awaiter(this, void 0, Promise, function () {
            return __generator(this, function (_a) {
                switch (_a.label) {
                    case 0:
                        if (!!this.ifCanvasDrawed) return [3 /*break*/, 2];
                        return [4 /*yield*/, this.toCanvas()];
                    case 1:
                        _a.sent();
                        _a.label = 2;
                    case 2: return [2 /*return*/, this.option.canvas];
                }
            });
        });
    };
    QrCodeWithLogo.version = version;
    return QrCodeWithLogo;
}());

export default QrCodeWithLogo;