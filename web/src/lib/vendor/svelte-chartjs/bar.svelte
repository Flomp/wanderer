<script lang="ts">
  import {
    Chart,
    Tooltip,
    BarController,
    type ChartData,
    type ChartOptions,
  } from 'chart.js';
  import type { HTMLCanvasAttributes } from 'svelte/elements';


  interface Props extends HTMLCanvasAttributes {
    data: ChartData<'bar', number[], string>;
    options: ChartOptions<'bar'>;
  }

  const { data, options, ...rest }: Props = $props();

  Chart.register(Tooltip, BarController);

  let canvasElem: HTMLCanvasElement;
  let chart: Chart;

  $effect(() => {
    chart = new Chart(canvasElem, {
      type: 'bar',
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
