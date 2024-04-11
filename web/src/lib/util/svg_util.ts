export function createRect(
    x: number,
    y: number,
    width: number,
    height: number,
    fill: string,
    border?: string,
) {
    const svgns = "http://www.w3.org/2000/svg";

    const rect = document.createElementNS(svgns, "rect");
    rect.setAttribute("x", x.toString());
    rect.setAttribute("y", y.toString());
    rect.setAttribute("width", width.toString());
    rect.setAttribute("height", height.toString());
    rect.setAttribute("style", `fill:${fill};stroke-width:1;stroke:${border ?? ''}`);
    return rect;
}

export function createText(content: string, x: number,
    y: number, fontSize?: string) {
    const svgns = "http://www.w3.org/2000/svg";

    const text = document.createElementNS(svgns, "text");
    text.setAttribute("x", x.toString());
    text.setAttribute("y", y.toString());
    text.style.fontSize = fontSize ?? ""
    text.textContent = content;
    return text;
}