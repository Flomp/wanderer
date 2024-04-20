interface ValhallaResponse {
    trip: {
        legs: {
            shape: string[];
        }[];
    };
}

class ValhallaRoute {
    type: "Feature";
    properties: {};
    geometry: {
        type: "LineString";
        coordinates: number[][];
    };

    constructor() {
        this.type = "Feature"
        this.properties = {}
        this.geometry = {
            type: "LineString",
            coordinates: []
        }
    }

    toString() {
        return JSON.stringify(this);
    }
}

export {type ValhallaResponse, ValhallaRoute}