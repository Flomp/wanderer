export function range(to: number, from: number = 0): ReadonlyArray<number> {
    return [...Array(to - from).keys()].map(i => i + from);
}