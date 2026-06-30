<script lang="ts">
  import { onDestroy } from "svelte";
  import { scaleLinear } from "d3-scale";
  import { line } from "d3-shape";
  import type { AnalysisDTO, XY } from "$lib/api";
  import { ZONE_COLORS } from "$lib/format";

  export let data: AnalysisDTO;
  export let onAnchorDrag: (lt1: number, lt2: number) => void = () => {};
  // static mode (e.g. the report): no layer toggles, no dragging.
  export let isStatic = false;

  const height = 360;
  const margin = { top: 16, right: 48, bottom: 36, left: 44 };

  let wrap: HTMLDivElement;
  let svgEl: SVGSVGElement;
  let width = 0;

  // --- layer visibility toggles (FR-C1/OI-18) --------------------------
  let showCurve = true;
  let showHR = true;
  let showPoints = true;
  let showZones = true;

  // --- live drag state -------------------------------------------------
  let activeDrag: "lt1" | "lt2" | null = null;
  let liveLt1: number | null = null;
  let liveLt2: number | null = null;

  // displayed anchor positions (live value while dragging, else from data)
  $: lt1Intensity = liveLt1 ?? data.lt1.intensity;
  $: lt2Intensity = liveLt2 ?? data.lt2.intensity;

  // --- scales (recompute reactively when data or width changes) --------
  function maxY(arr: XY[]): number {
    let m = 0;
    for (const p of arr) if (p.y > m) m = p.y;
    return m;
  }

  $: x = scaleLinear()
    .domain([data.domainLow, data.domainHigh])
    .range([margin.left, Math.max(margin.left, width - margin.right)]);

  $: maxLac = Math.max(maxY(data.rawPoints), maxY(data.curve), 1) * 1.1;
  $: yLactate = scaleLinear()
    .domain([0, maxLac])
    .range([height - margin.bottom, margin.top]);

  $: hrVals = data.hrPoints.map((p) => p.y);
  $: hrMin = hrVals.length ? Math.min(...hrVals) : 0;
  $: hrMax = hrVals.length ? Math.max(...hrVals) : 1;
  $: yHR = scaleLinear()
    .domain(hrMin === hrMax ? [hrMin - 1, hrMax + 1] : [hrMin, hrMax])
    .range([height - margin.bottom, margin.top]);

  // --- generated paths -------------------------------------------------
  $: curvePath =
    line<XY>()
      .x((d) => x(d.x))
      .y((d) => yLactate(d.y))(data.curve) ?? "";
  $: hrPath =
    line<XY>()
      .x((d) => x(d.x))
      .y((d) => yHR(d.y))(data.hrPoints) ?? "";

  // --- ticks -----------------------------------------------------------
  $: xTicks = width > 0 ? x.ticks(6) : [];
  $: lacTicks = yLactate.ticks(6);
  $: hrTicks = yHR.ticks(6);

  function fmt(v: number, places = 1): string {
    return v % 1 === 0 ? String(v) : v.toFixed(places);
  }

  // --- drag ------------------------------------------------------------
  function startDrag(which: "lt1" | "lt2") {
    return (e: MouseEvent) => {
      e.preventDefault();
      activeDrag = which;
      window.addEventListener("mousemove", onMove);
      window.addEventListener("mouseup", onUp);
    };
  }

  function onMove(e: MouseEvent) {
    if (!activeDrag || !svgEl) return;
    const rect = svgEl.getBoundingClientRect();
    const px = e.clientX - rect.left;
    let intensity = x.invert(px);
    intensity = Math.max(data.domainLow, Math.min(data.domainHigh, intensity));
    if (activeDrag === "lt1") {
      liveLt1 = intensity;
      onAnchorDrag(intensity, lt2Intensity);
    } else {
      liveLt2 = intensity;
      onAnchorDrag(lt1Intensity, intensity);
    }
  }

  function onUp() {
    activeDrag = null;
    liveLt1 = null;
    liveLt2 = null;
    window.removeEventListener("mousemove", onMove);
    window.removeEventListener("mouseup", onUp);
  }

  onDestroy(() => {
    window.removeEventListener("mousemove", onMove);
    window.removeEventListener("mouseup", onUp);
  });

  $: plotTop = margin.top;
  $: plotBottom = height - margin.bottom;
</script>

<div class="chart" class:static={isStatic} bind:this={wrap} bind:clientWidth={width}>
  {#if !isStatic}
  <!-- layer toggles (top-right, clear of plot; container is click-through) -->
  <div class="layers" role="group" aria-label="Chart layers">
    <button
      type="button"
      class="chip"
      class:on={showCurve}
      aria-pressed={showCurve}
      on:click={() => (showCurve = !showCurve)}
    >
      Curve
    </button>
    <button
      type="button"
      class="chip"
      class:on={showHR}
      aria-pressed={showHR}
      on:click={() => (showHR = !showHR)}
    >
      HR
    </button>
    <button
      type="button"
      class="chip"
      class:on={showPoints}
      aria-pressed={showPoints}
      on:click={() => (showPoints = !showPoints)}
    >
      Points
    </button>
    <button
      type="button"
      class="chip"
      class:on={showZones}
      aria-pressed={showZones}
      on:click={() => (showZones = !showZones)}
    >
      Zones
    </button>
  </div>
  {/if}

  {#if width > 0}
    <svg
      bind:this={svgEl}
      {width}
      {height}
      viewBox={`0 0 ${width} ${height}`}
      role="img"
      aria-label="Lactate fitting chart"
    >
      <!-- (0) gridlines + axes -->
      <g class="axis">
        {#each lacTicks as t}
          <line
            class="grid"
            x1={margin.left}
            x2={width - margin.right}
            y1={yLactate(t)}
            y2={yLactate(t)}
          />
          <text
            class="mono tick"
            x={margin.left - 6}
            y={yLactate(t)}
            text-anchor="end"
            dominant-baseline="middle">{fmt(t)}</text
          >
        {/each}

        {#if showHR}
          {#each hrTicks as t}
            <text
              class="mono tick hr"
              x={width - margin.right + 6}
              y={yHR(t)}
              text-anchor="start"
              dominant-baseline="middle">{Math.round(t)}</text
            >
          {/each}
        {/if}

        {#each xTicks as t}
          <text
            class="mono tick"
            x={x(t)}
            y={plotBottom + 16}
            text-anchor="middle"
            dominant-baseline="hanging">{fmt(t)}</text
          >
        {/each}

        <line class="axis-line" x1={margin.left} x2={width - margin.right} y1={plotBottom} y2={plotBottom} />

        <text class="mono axis-title" x={margin.left} y={plotTop - 4} text-anchor="start">
          mmol/L
        </text>
        {#if showHR}
          <text class="mono axis-title hr" x={width - margin.right} y={plotTop - 4} text-anchor="end">
            bpm
          </text>
        {/if}
        <text class="mono axis-title" x={(margin.left + width - margin.right) / 2} y={height - 2} text-anchor="middle">
          {data.unit || "intensity"}
        </text>
      </g>

      <!-- (1) zone bands -->
      {#if showZones}
        <g class="zones">
          {#each data.zones as z (z.index)}
            <rect
              x={x(z.intensityLow)}
              width={Math.max(0, x(z.intensityHigh) - x(z.intensityLow))}
              y={plotTop}
              height={plotBottom - plotTop}
              fill={ZONE_COLORS[z.label] ?? "var(--surface-2)"}
              fill-opacity="0.12"
            />
          {/each}
        </g>
      {/if}

      <!-- (3) HR overlay (behind lactate curve, faint dashed) -->
      {#if showHR && hrPath}
        <path class="hr-line" d={hrPath} />
      {/if}

      <!-- (2) fitted lactate curve -->
      {#if showCurve && curvePath}
        <path class="curve" d={curvePath} />
      {/if}

      <!-- (4) raw measured points -->
      {#if showPoints}
        <g class="raw">
          {#each data.rawPoints as p}
            <circle cx={x(p.x)} cy={yLactate(p.y)} r="3.5" />
          {/each}
        </g>
      {/if}

      <!-- (5) threshold markers (always visible) -->
      {#each [{ key: "lt1", label: "IAS", xi: lt1Intensity, manual: data.lt1.manual }, { key: "lt2", label: "IANS", xi: lt2Intensity, manual: data.lt2.manual }] as m}
        <g
          class="marker"
          class:dragging={activeDrag === m.key}
          on:mousedown={isStatic ? undefined : startDrag(m.key === "lt1" ? "lt1" : "lt2")}
          role="slider"
          tabindex="0"
          aria-label={m.label}
          aria-valuenow={Math.round(m.xi)}
        >
          <!-- wide invisible hit area -->
          <rect class="hit" x={x(m.xi) - 6} width="12" y={plotTop} height={plotBottom - plotTop} />
          <line
            class="marker-line"
            class:manual={m.manual}
            x1={x(m.xi)}
            x2={x(m.xi)}
            y1={plotTop}
            y2={plotBottom}
          />
          <rect class="handle" class:manual={m.manual} x={x(m.xi) - 5} y={plotTop - 2} width="10" height="10" rx="2" />
          <text class="mono label" class:manual={m.manual} x={x(m.xi)} y={plotTop - 6} text-anchor="middle">
            {m.label}
          </text>
        </g>
      {/each}
    </svg>
  {/if}
</div>

<style>
  .chart {
    position: relative;
    width: 100%;
    height: 360px;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    overflow: hidden;
  }
  svg {
    display: block;
  }

  /* layer toggles — subtle, top-right, click-through container so they
     never swallow marker drag events outside the chip hit-areas */
  .layers {
    position: absolute;
    top: var(--space-2);
    right: var(--space-2);
    z-index: 2;
    display: flex;
    gap: 2px;
    padding: 2px;
    border-radius: var(--radius-pill);
    background: var(--surface);
    border: 1px solid var(--border);
    pointer-events: none;
  }
  .chip {
    pointer-events: auto;
    appearance: none;
    border: 0;
    background: transparent;
    cursor: pointer;
    font-family: inherit;
    font-size: var(--fs-eyebrow);
    line-height: 1;
    letter-spacing: 0.02em;
    color: var(--text-faint);
    padding: 3px 7px;
    border-radius: var(--radius-pill);
    transition: color 0.12s ease, background-color 0.12s ease;
  }
  .chip:hover {
    color: var(--text-muted);
  }
  .chip.on {
    color: var(--accent);
    background: var(--surface-2);
  }
  .chip:focus-visible {
    outline: 2px solid var(--focus);
    outline-offset: 1px;
  }

  .tick {
    font-size: var(--fs-eyebrow);
    fill: var(--text-faint);
  }
  .tick.hr {
    fill: var(--series-hr);
    opacity: 0.75;
  }
  .axis-title {
    font-size: var(--fs-eyebrow);
    fill: var(--text-faint);
    letter-spacing: 0.04em;
  }
  .axis-title.hr {
    fill: var(--series-hr);
    opacity: 0.8;
  }
  .grid {
    stroke: var(--border);
    stroke-width: 1;
    opacity: 0.5;
  }
  .axis-line {
    stroke: var(--border-strong);
    stroke-width: 1;
  }

  .curve {
    fill: none;
    stroke: var(--series-lactate);
    stroke-width: 2;
    stroke-linejoin: round;
    stroke-linecap: round;
  }
  .hr-line {
    fill: none;
    stroke: var(--series-hr);
    stroke-width: 1.5;
    stroke-dasharray: 4 3;
    opacity: 0.55;
  }
  .raw circle {
    fill: var(--surface);
    stroke: var(--series-lactate);
    stroke-width: 1.5;
  }

  /* threshold markers */
  .marker {
    cursor: ew-resize;
    outline: none;
  }
  .static .marker {
    cursor: default;
  }
  .hit {
    fill: transparent;
  }
  /* the turn point — the app's signature — is marked in signal red. Solid =
     algorithmic, dashed = manual override (FR-C3). */
  .marker-line {
    stroke: var(--series-turnpoint);
    stroke-width: 1.25;
  }
  .marker-line.manual {
    stroke: var(--series-turnpoint);
    stroke-width: 1.5;
    stroke-dasharray: 5 3;
  }
  .handle {
    fill: var(--surface);
    stroke: var(--series-turnpoint);
    stroke-width: 1.25;
  }
  .handle.manual {
    stroke: var(--series-turnpoint);
    stroke-width: 1.5;
  }
  .marker.dragging .handle,
  .marker:hover .handle {
    fill: var(--series-turnpoint);
    stroke: var(--series-turnpoint);
  }
  .label {
    font-size: var(--fs-micro);
    letter-spacing: var(--track-micro);
    text-transform: uppercase;
    fill: var(--text-muted);
    pointer-events: none;
  }
  .label.manual {
    fill: var(--series-turnpoint);
  }
</style>
