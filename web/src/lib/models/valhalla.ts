interface ValhallaResponse {
    trip: {
        legs: {
            shape: string;
        }[];
    };
}

export { type ValhallaResponse }