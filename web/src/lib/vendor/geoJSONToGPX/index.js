export default function GeoJsonToGpx(geoJson, options, implementation) {
    if (implementation === void 0) { implementation = document.implementation; }
    var doc = implementation.createDocument('http://www.topografix.com/GPX/1/1', '');
    var instruct = doc.createProcessingInstruction('xml', 'version="1.0" encoding="UTF-8"');
    doc.appendChild(instruct);
    var defaultPackageName = '@dwayneparton/geojson-to-gpx';
    var creator = (options === null || options === void 0 ? void 0 : options.creator) || defaultPackageName;
    var createElementWithNS = function (tagName) {
        return doc.createElementNS('http://www.topografix.com/GPX/1/1', tagName);
    };
    var gpx = createElementWithNS('gpx');
    gpx.setAttribute('version', '1.1');
    gpx.setAttribute('creator', creator);
    var wpts = [];
    var trks = [];
    function createTagInParentElement(parent, tagName, content) {
        if (content === undefined) {
            return;
        }
        var element = createElementWithNS(tagName);
        var contentEl = doc.createTextNode(String(content));
        element.appendChild(contentEl);
        parent.appendChild(element);
    }
    function addSupportedPropertiesFromObject(el, supports, properties) {
        if (properties && typeof properties === 'object') {
            supports.forEach(function (key) {
                var value = properties[key];
                if (value && typeof value === 'string' && supports.includes(key)) {
                    createTagInParentElement(el, key, value);
                }
            });
        }
    }
    function createLinkInParentElement(parent, props) {
        var href = props.href;
        if (!href) {
            return;
        }
        var el = createElementWithNS('link');
        el.setAttribute('href', href);
        addSupportedPropertiesFromObject(el, ['text', 'type'], props);
        parent.appendChild(el);
    }
    function createTrk(properties) {
        var el = createElementWithNS('trk');
        var supports = ['name', 'desc', 'src', 'type'];
        addSupportedPropertiesFromObject(el, supports, properties);
        return el;
    }
    function createPt(type, position, properties) {
        var lon = position[0], lat = position[1], ele = position[2], time = position[3];
        var el = createElementWithNS(type);
        el.setAttribute('lat', String(lat));
        el.setAttribute('lon', String(lon));
        createTagInParentElement(el, 'ele', ele);
        createTagInParentElement(el, 'time', time);
        var supports = ['name', 'desc', 'src', 'type'];
        addSupportedPropertiesFromObject(el, supports, properties);
        return el;
    }
    function createTrkSeg(coordinates) {
        var el = createElementWithNS('trkseg');
        coordinates.forEach(function (point) {
            el.appendChild(createPt('trkpt', point));
        });
        return el;
    }
    function interpretFeature(feature) {
        var geometry = feature.geometry, properties = feature.properties;
        var type = geometry.type;
        switch (type) {
            case 'Polygon':
                break;
            case 'Point': {
                wpts.push(createPt('wpt', geometry.coordinates, properties));
                break;
            }
            case 'MultiPoint': {
                geometry.coordinates.forEach(function (coord) {
                    wpts.push(createPt('wpt', coord, properties));
                });
                break;
            }
            case 'LineString': {
                var lineTrk = createTrk(properties);
                var trkseg = createTrkSeg(geometry.coordinates);
                lineTrk.appendChild(trkseg);
                trks.push(lineTrk);
                break;
            }
            case 'MultiLineString': {
                var trk_1 = createTrk(properties);
                geometry.coordinates.forEach(function (pos) {
                    var trkseg = createTrkSeg(pos);
                    trk_1.appendChild(trkseg);
                });
                trks.push(trk_1);
                break;
            }
            default:
                break;
        }
    }
    if (options && typeof options.metadata === 'object') {
        var meta = options.metadata;
        var metadata = createElementWithNS('metadata');
        createTagInParentElement(metadata, 'name', meta.name);
        createTagInParentElement(metadata, 'desc', meta.desc);
        if (meta.author && typeof meta.author === 'object') {
            var author = createElementWithNS('author');
            createTagInParentElement(author, 'name', meta.author.name);
            createTagInParentElement(author, 'email', meta.author.email);
            if (meta.author.link && typeof meta.author.link === 'object') {
                createLinkInParentElement(author, meta.author.link);
            }
            metadata.appendChild(author);
        }
        if (typeof meta.copyright === 'object') {
            var copyright = createElementWithNS('copyright');
            if (meta.copyright.author) {
                copyright.setAttribute('author', meta.copyright.author);
            }
            createTagInParentElement(copyright, 'year', meta.copyright.year);
            createTagInParentElement(copyright, 'license', meta.copyright.license);
            metadata.appendChild(copyright);
        }
        if (typeof meta.link === 'object') {
            createLinkInParentElement(metadata, meta.link);
        }
        createTagInParentElement(metadata, 'time', meta.time);
        createTagInParentElement(metadata, 'keywords', meta.keywords);
        gpx.appendChild(metadata);
    }
    var type = geoJson.type;
    switch (type) {
        case 'Feature': {
            interpretFeature(geoJson);
            break;
        }
        case 'FeatureCollection': {
            var features = geoJson.features;
            features.forEach(function (feature) {
                interpretFeature(feature);
            });
            break;
        }
        default:
            break;
    }
    wpts.forEach(function (wpt) { return gpx.appendChild(wpt); });
    trks.forEach(function (trk) { return gpx.appendChild(trk); });
    doc.appendChild(gpx);
    return doc;
}
