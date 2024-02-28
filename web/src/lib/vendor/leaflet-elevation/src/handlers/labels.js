/**
 * @see https://github.com/Raruto/leaflet-elevation/issues/211
 *
 * @example
 * ```js
 * L.control.Elevation({ handlers: [ 'Labels' ], labelsRotation: 25, labelsAlign: 'start' })
 * ```
 */
export function Labels() {

  this.on('elechart_updated', function(e) {

    const pointG = this._chart._chart.pane('point');

    const textRotation = this.options.labelsRotation ?? -90;
    const textAnchor = this.options.labelsAlign;

    if (90 == Math.abs(textRotation)) {

      pointG.selectAll('text')
        .attr('dy', '4px')
        .attr("dx", (d, i, el) => Math.sign(textRotation) * (this._height() - d3.select(el[i].parentElement).datum().y - 8) + 'px')
        .attr('text-anchor', textRotation > 0 ? 'end' : 'start')
        .attr('transform', 'rotate(' + textRotation + ')')

      pointG.selectAll('circle')
        .attr('r', '2.5')
        .attr('fill', '#fff')
        .attr("stroke", '#000')
        .attr("stroke-width", '1.1');

    } else if (!isNaN(textRotation)) {

      pointG.selectAll('text')
        .attr("dx", "4px")
        .attr("dy", "-9px")
        .attr('text-anchor', textAnchor ?? (0 == textRotation ? 'start' : 'middle'))
        .attr('transform', 'rotate('+ -textRotation +')')

    }

  });

	return { };
}
