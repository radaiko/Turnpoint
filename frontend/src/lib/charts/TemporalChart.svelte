<script lang="ts">
  import { scaleLinear } from "d3-scale";
  import { line } from "d3-shape";
  import type { AnalysisDTO } from "$lib/api";
  import { mmss } from "$lib/format";

  // Primary results temporal chart (FR-C4 "Zeitliche Darstellung").
  // Time on X; lactate on the left axis, heart rate on the right axis,
  // intensity steps as faint backing bars behind both series.
  export let data: AnalysisDTO;

  type Pt = { x: number; y: number };

  const margin = { top: 16, right: 48, bottom: 36, left: 44 };
  const height = 360;

  let clientWidth = 0;

  // --- source series (defensive against partial DTOs) -------------------
  $: timeLactate = (data?.timeLactate ?? []) as Pt[];
  $: timeHR = (data?.timeHR ?? []) as Pt[];
  $: stepBars = data?.stepBars ?? [];
  $: maxIntensity = data?.maxIntensity ?? 0;

  $: hasData = timeLactate.length > 0 || timeHR.length > 0;
  $: ready = clientWidth > 0 && hasData;

  // --- layout -----------------------------------------------------------
  $: width = clientWidth || 0;
  $: innerW = Math.max(0, width - margin.left - margin.right);
  $: innerH = Math.max(0, height - margin.top - margin.bottom);

  // --- domains ----------------------------------------------------------
  const maxOf = (xs: number[], fb = 0) => (xs.length ? Math.max(...xs) : fb);
  const minOf = (xs: number[], fb = 0) => (xs.length ? Math.min(...xs) : fb);

  $: maxTime = maxOf([
    ...timeLactate.map((p) => p.x),
    ...timeHR.map((p) => p.x),
    ...stepBars.map((b) => b.endS),
  ]);

  // include the 4 mmol/L reference so the line always stays in view
  $: maxLactate = maxOf([4, ...timeLactate.map((p) => p.y)], 4);

  $: hrVals = timeHR.map((p) => p.y);
  $: minHR = minOf(hrVals, 0);
  $: maxHR = maxOf(hrVals, 0);

  // --- scales -----------------------------------------------------------
  $: x = scaleLinear()
    .domain([0, maxTime || 1])
    .range([0, innerW]);

  $: yLactate = scaleLinear()
    .domain([0, (maxLactate || 1) * 1.1])
    .range([innerH, 0]);

  $: yHR = scaleLinear()
    .domain([minHR - 5, maxHR + 5])
    .range([innerH, 0]);

  // hidden third scale: intensity → chart height (for the backing bars)
  $: yIntensity = scaleLinear()
    .domain([0, maxIntensity || 1])
    .range([innerH, 0]);

  // --- line generators --------------------------------------------------
  $: lactatePath =
    line<Pt>()
      .x((p) => x(p.x))
      .y((p) => yLactate(p.y))(timeLactate) ?? "";

  $: hrPath =
    line<Pt>()
      .x((p) => x(p.x))
      .y((p) => yHR(p.y))(timeHR) ?? "";

  // --- ticks ------------------------------------------------------------
  $: xTicks = ready ? x.ticks(Math.max(2, Math.floor(innerW / 90))) : [];
  $: lactateTicks = ready ? yLactate.ticks(5) : [];
  $: hrTicks = ready ? yHR.ticks(5) : [];

  const fmtTime = (s: number) => (s <= 0 ? "00:00" : mmss(s));
</script>

<div class="chart" bind:clientWidth>
  {#if ready}
    <div class="legend">
      <span class="chip">
        <span class="swatch line" style="border-top-color: var(--series-lactate)"></span>
        Lactate
      </span>
      <span class="chip">
        <span class="swatch line" style="border-top-color: var(--series-hr)"></span>
        HR
      </span>
      <span class="chip">
        <span class="swatch bar"></span>
        Intensity
      </span>
    </div>

    <svg {width} {height} role="img" aria-label="Temporal lactate and heart rate chart">
      <g transform="translate({margin.left},{margin.top})">
        <!-- (1) intensity step bars (behind everything) -->
        {#each stepBars as b}
          {@const bx = x(b.startS)}
          {@const bw = Math.max(0, x(b.endS) - x(b.startS))}
          {@const by = yIntensity(b.intensity)}
          <rect
            class="step"
            x={bx}
            y={by}
            width={bw}
            height={Math.max(0, innerH - by)}
          />
        {/each}

        <!-- horizontal gridlines (left / lactate axis) -->
        {#each lactateTicks as t}
          <line class="grid" x1="0" x2={innerW} y1={yLactate(t)} y2={yLactate(t)} />
        {/each}

        <!-- (4) 4 mmol/L reference line -->
        <line class="ref" x1="0" x2={innerW} y1={yLactate(4)} y2={yLactate(4)} />
        <text class="ref-label num" x={innerW - 4} y={yLactate(4) - 5} text-anchor="end">
          4 mmol/L
        </text>

        <!-- (2) lactate line -->
        <path class="series-lactate" d={lactatePath} />

        <!-- (3) HR line -->
        <path class="series-hr" d={hrPath} />

        <!-- (5) data point markers -->
        {#each timeLactate as p}
          <circle class="dot-lactate" cx={x(p.x)} cy={yLactate(p.y)} r="2.5" />
        {/each}
        {#each timeHR as p}
          <circle class="dot-hr" cx={x(p.x)} cy={yHR(p.y)} r="2" />
        {/each}

        <!-- X axis -->
        <line class="axis" x1="0" x2={innerW} y1={innerH} y2={innerH} />
        {#each xTicks as t}
          <g transform="translate({x(t)},{innerH})">
            <line class="tick" y1="0" y2="4" />
            <text class="tick-label num" y="16" text-anchor="middle">{fmtTime(t)}</text>
          </g>
        {/each}

        <!-- Left Y axis (lactate) -->
        <line class="axis" x1="0" x2="0" y1="0" y2={innerH} />
        {#each lactateTicks as t}
          <g transform="translate(0,{yLactate(t)})">
            <line class="tick" x1="-4" x2="0" />
            <text class="tick-label num" x="-8" dy="0.32em" text-anchor="end">{t}</text>
          </g>
        {/each}
        <text
          class="axis-title"
          transform="translate({-margin.left + 12},{innerH / 2}) rotate(-90)"
          text-anchor="middle">Lactate [mmol/L]</text
        >

        <!-- Right Y axis (heart rate) -->
        <line class="axis" x1={innerW} x2={innerW} y1="0" y2={innerH} />
        {#each hrTicks as t}
          <g transform="translate({innerW},{yHR(t)})">
            <line class="tick" x1="0" x2="4" />
            <text class="tick-label num" x="8" dy="0.32em" text-anchor="start">{t}</text>
          </g>
        {/each}
        <text
          class="axis-title"
          transform="translate({innerW + margin.right - 12},{innerH / 2}) rotate(90)"
          text-anchor="middle">HR [bpm]</text
        >
      </g>
    </svg>
  {:else}
    <div class="empty">No temporal data</div>
  {/if}
</div>

<style>
  .chart {
    position: relative;
    width: 100%;
  }

  svg {
    display: block;
    width: 100%;
    height: auto;
  }

  /* legend -------------------------------------------------------------- */
  .legend {
    position: absolute;
    top: 0;
    right: var(--space-4);
    display: flex;
    gap: var(--space-3);
  }
  .chip {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: var(--fs-eyebrow);
    color: var(--text-muted);
  }
  .swatch {
    display: inline-block;
  }
  .swatch.line {
    width: 14px;
    height: 0;
    border-top: 2px solid;
  }
  .swatch.bar {
    width: 11px;
    height: 10px;
    border-radius: 2px;
    background: var(--border-strong);
    opacity: 0.4;
  }

  /* data layers --------------------------------------------------------- */
  .step {
    fill: var(--border-strong);
    opacity: 0.4;
  }
  .series-lactate {
    fill: none;
    stroke: var(--series-lactate);
    stroke-width: 2;
    stroke-linejoin: round;
    stroke-linecap: round;
  }
  .series-hr {
    fill: none;
    stroke: var(--series-hr);
    stroke-width: 2;
    stroke-linejoin: round;
    stroke-linecap: round;
  }
  .dot-lactate {
    fill: var(--series-lactate);
  }
  .dot-hr {
    fill: var(--series-hr);
  }

  /* reference + axes ---------------------------------------------------- */
  .ref {
    stroke: var(--text-faint);
    stroke-width: 1;
    stroke-dasharray: 4 4;
  }
  .ref-label {
    fill: var(--text-faint);
    font-family: var(--font-mono);
    font-size: var(--fs-eyebrow);
  }
  .grid {
    stroke: var(--border);
    stroke-width: 1;
    opacity: 0.45;
  }
  .axis {
    stroke: var(--border-strong);
    stroke-width: 1;
  }
  .tick {
    stroke: var(--border-strong);
    stroke-width: 1;
  }
  .tick-label {
    fill: var(--text-faint);
    font-family: var(--font-mono);
    font-size: var(--fs-eyebrow);
  }
  .axis-title {
    fill: var(--text-muted);
    font-size: var(--fs-caption);
  }

  .num {
    font-variant-numeric: tabular-nums;
  }

  .empty {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 360px;
    color: var(--text-faint);
    font-size: var(--fs-caption);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    background: var(--surface);
  }
</style>
