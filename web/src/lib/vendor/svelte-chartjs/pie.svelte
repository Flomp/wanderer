<script lang="ts">
  import {
    Chart,
    PieController,
    Tooltip,
    type ChartData,
    type ChartOptions,
  } from 'chart.js';
  import type { HTMLCanvasAttributes } from 'svelte/elements';

  interface Props extends HTMLCanvasAttributes {
    data: ChartData<'pie', number[], string>;
    options: ChartOptions<'pie'>;
  }

  const { data, options, ...rest }: Props = $props();

  Chart.register(Tooltip, PieController);

  let canvasElem: HTMLCanvasElement;
  let chart: Chart;

  $effect(() => {
    chart = new Chart(canvasElem, {
      type: 'pie',
      data,
      options,
    });

    return () => {
      chart.destroy();
    };
  });

  $effect(() => {
    if (chart) {
      chart.data = data;
      chart.update();
    }
  });
</script>

<canvas bind:this={canvasElem} {...rest}></canvas>