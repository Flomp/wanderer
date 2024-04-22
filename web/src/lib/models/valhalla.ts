interface ValhallaRouteResponse {
    trip: {
        legs: {
            shape: string;
        }[];
    };
}

interface ValhallaHeightResponse {
    height: number[];
}

export { type ValhallaRouteResponse, type ValhallaHeightResponse }