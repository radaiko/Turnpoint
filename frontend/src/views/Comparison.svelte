<script lang="ts">
  import { onMount } from "svelte";
  import { scaleLinear } from "d3-scale";
  import { line } from "d3-shape";
  import { App, type AthleteSummary, type Test, type AnalysisDTO, type MarkerRow } from "$lib/api";
  import { num, intStr } from "$lib/format";
  import { toast } from "$lib/stores/toast";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import Select from "$lib/components/Select.svelte";

  // distinct overlay strokes, cycled per selected test (sorted order)
  const SERIES = [
    "var(--series-lactate)",
    "var(--series-hr)",
    "var(--zone-ga1)",
    "var(--zone-eb)",
    "var(--zone-sb)",
  ];

  let athletes: AthleteSummary[] = [];
  let athleteId = 0; // 0 = none chosen
  let tests: Test[] = [];
  let loadingTests = false;

  // selection + memoised analyses keyed by test id
  let selected: Record<number, boolean> = {};
  let cache: Record<number, AnalysisDTO> = {};
  let pending: Record<number, boolean> = {};

  // latest-wins guard — bumped on every athlete switch so stale awaits drop
  let token = 0;

  onMount(async () => {
    athletes = await App.ListAthletes("");
  });

  async function onAthleteChange() {
    const mine = ++token;
    selected = {};
    cache = {};
    pending = {};
    tests = [];
    if (!athleteId) return;
    loadingTests = true;
    try {
      const t = await App.ListTests(athleteId);
      if (mine !== token) return; // superseded
      tests = t;
    } catch (e) {
      if (mine === token) toast("Failed to load tests", "danger");
    } finally {
      if (mine === token) loadingTests = false;
    }
  }

  async function toggleTest(t: Test) {
    const on = !selected[t.id];
    selected = { ...selected, [t.id]: on };
    if (!on || cache[t.id]) return;
    const mine = token;
    pending = { ...pending, [t.id]: true };
    try {
      const dto = await App.Analyze(t.id);
      if (mine !== token) return; // athlete changed mid-flight
      cache = { ...cache, [t.id]: dto };
    } catch (e) {
      if (mine === token) {
        selected = { ...selected, [t.id]: false };
        toast("Analysis failed for that test", "danger");
      }
    } finally {
      if (mine === token) {
        const { [t.id]: _drop, ...rest } = pending;
        pending = rest;
      }
    }
  }

  function maxMarker(dto: AnalysisDTO): MarkerRow | null {
    return dto.markers?.find((m) => m.marker === "MAX") ?? null;
  }

  function fmtTick(v: number): string {
    return Number.isInteger(v) ? String(v) : v.toFixed(1);
  }

  // chosen = checked tests that have an analysis, sorted by test date (ascending)
  $: chosen = tests
    .filter((t) => selected[t.id] && cache[t.id])
    .map((t) => ({ test: t, dto: cache[t.id] }))
    .sort((a, b) => a.test.testDate.localeCompare(b.test.testDate));

  $: anyPending = Object.values(pending).some(Boolean);

  // ---- overlay chart geometry ----
  let chartW = 0;
  const chartH = 320;
  const M = { top: 16, right: 16, bottom: 28, left: 40 };

  $: innerW = Math.max(10, chartW - M.left - M.right);
  $: innerH = chartH - M.top - M.bottom;

  $: allPts = chosen.flatMap((c) => c.dto.curve ?? []);
  $: xLo = allPts.length ? Math.min(...allPts.map((p) => p.x)) : 0;
  $: xHiRaw = allPts.length ? Math.max(...allPts.map((p) => p.x)) : 1;
  $: xHi = xHiRaw > xLo ? xHiRaw : xLo + 1;
  $: yHi = allPts.length ? Math.max(...allPts.map((p) => p.y)) : 1;

  $: x = scaleLinear().domain([xLo, xHi]).range([0, innerW]);
  $: y = scaleLinear().domain([0, yHi]).range([innerH, 0]).nice();
  $: lineGen = line<{ x: number; y: number }>()
    .x((d) => x(d.x))
    .y((d) => y(d.y));

  $: paths = chosen.map((c, i) => ({
    d: lineGen(c.dto.curve ?? "") ?? "",
    color: SERIES[i % SERIES.length],
    date: c.test.testDate,
  }));

  $: xTicks = allPts.length ? x.ticks(6) : [];
  $: yTicks = allPts.length ? y.ticks(5) : [];
</script>

<div class="page">
  <header class="head">
    <div>
      <h1>Comparison</h1>
      <p class="sub">Track thresholds across an athlete's tests, side by side.</p>
    </div>
    <div class="picker">
      <Select
        label="Athlete"
        bind:value={athleteId}
        on:change={onAthleteChange}
        options={[
          { value: 0, label: "Select athlete…" },
          ...athletes.map((a) => ({ value: a.id, label: a.name })),
        ]}
      />
    </div>
  </header>

  {#if !athleteId}
    <EmptyState title="No athlete selected" hint="Pick an athlete above to compare their step tests." />
  {:else if loadingTests}
    <EmptyState title="Loading tests…" hint="Fetching this athlete's tests." />
  {:else if !tests.length}
    <EmptyState title="No tests" hint="This athlete has no tests to compare yet." />
  {:else}
    <section class="panel checklist">
      <div class="panel-head"><h3>Tests</h3><span class="hint">Tick tests to overlay them.</span></div>
      <div class="chks">
        {#each tests as t (t.id)}
          <label class="chk" class:on={selected[t.id]}>
            <input type="checkbox" checked={!!selected[t.id]} on:change={() => toggleTest(t)} />
            <span class="mono">{t.testDate}</span>
            <span class="tag">{t.sport}</span>
            {#if pending[t.id]}<span class="spin">analyzing…</span>{/if}
          </label>
        {/each}
      </div>
    </section>

    {#if !chosen.length}
      <EmptyState
        title={anyPending ? "Analyzing…" : "Nothing selected"}
        hint={anyPending ? "Crunching the selected tests." : "Tick one or more tests above to build the comparison."}
      />
    {:else}
      <section class="panel">
        <div class="panel-head"><h3>Longitudinal</h3></div>
        <div class="tbl-wrap">
          <table class="tbl">
            <thead>
              <tr>
                <th class="l">Date</th>
                <th class="r">IAS km/h</th>
                <th class="r">IANS km/h</th>
                <th class="r">IAS HR</th>
                <th class="r">IANS HR</th>
                <th class="r">MAX</th>
              </tr>
            </thead>
            <tbody>
              {#each chosen as c, i (c.test.id)}
                <tr>
                  <td class="l">
                    <span class="swatch" style="--c:{SERIES[i % SERIES.length]}"></span>
                    <span class="mono">{c.test.testDate}</span>
                  </td>
                  <td class="r num">{num(c.dto.lt1?.intensity)}</td>
                  <td class="r num">{num(c.dto.lt2?.intensity)}</td>
                  <td class="r num">{intStr(c.dto.lt1?.heartRate)}</td>
                  <td class="r num">{intStr(c.dto.lt2?.heartRate)}</td>
                  <td class="r num">{num(maxMarker(c.dto)?.intensity)}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </section>

      <section class="panel">
        <div class="panel-head"><h3>Lactate curves</h3></div>
        <div class="chart" bind:clientWidth={chartW}>
          <svg width={chartW} height={chartH} role="img" aria-label="Overlaid lactate curves">
            <g transform="translate({M.left},{M.top})">
              {#each yTicks as ty}
                <g class="grid" transform="translate(0,{y(ty)})">
                  <line x1="0" x2={innerW} y1="0" y2="0" />
                  <text x="-8" y="0" dy="0.32em" class="axis r">{fmtTick(ty)}</text>
                </g>
              {/each}
              {#each xTicks as tx}
                <text x={x(tx)} y={innerH + 18} class="axis c">{fmtTick(tx)}</text>
              {/each}
              <line class="axis-line" x1="0" x2={innerW} y1={innerH} y2={innerH} />
              {#each paths as p}
                <path d={p.d} fill="none" stroke={p.color} stroke-width="2" stroke-linejoin="round" />
              {/each}
            </g>
          </svg>
          <div class="legend">
            {#each chosen as c, i (c.test.id)}
              <span class="leg">
                <span class="swatch" style="--c:{SERIES[i % SERIES.length]}"></span>
                <span class="mono">{c.test.testDate}</span>
              </span>
            {/each}
          </div>
        </div>
      </section>
    {/if}
  {/if}
</div>

<style>
  .page {
    max-width: 1100px;
    margin: 0 auto;
    padding: var(--space-5);
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }
  .head {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: var(--space-4);
  }
  h1 {
    font-size: var(--fs-h1);
    margin: 0;
  }
  .sub {
    font-size: var(--fs-caption);
    color: var(--text-muted);
    margin: 2px 0 0;
  }
  .picker {
    width: 280px;
  }
  h3 {
    font-size: var(--fs-h3);
    margin: 0;
  }
  .hint {
    font-size: var(--fs-caption);
    color: var(--text-faint);
  }

  .panel {
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    background: var(--surface);
  }
  .panel-head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: var(--space-3);
    padding: var(--space-3) var(--space-4);
    border-bottom: 1px solid var(--border);
  }

  .chks {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    padding: var(--space-3) var(--space-4);
  }
  .chk {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-3);
    border: 1px solid var(--border);
    border-radius: var(--radius-pill);
    background: var(--surface);
    cursor: pointer;
  }
  .chk:hover {
    border-color: var(--border-strong);
  }
  .chk.on {
    border-color: var(--accent);
    background: color-mix(in srgb, var(--accent) 8%, transparent);
  }
  .chk input {
    accent-color: var(--accent);
  }
  .tag {
    font-size: var(--fs-eyebrow);
    text-transform: uppercase;
    letter-spacing: 0.04em;
    padding: 1px 7px;
    border-radius: var(--radius-pill);
    background: var(--surface-2);
    color: var(--text-muted);
  }
  .spin {
    font-size: var(--fs-caption);
    color: var(--text-faint);
  }

  .tbl-wrap {
    overflow-x: auto;
  }
  .tbl {
    width: 100%;
    border-collapse: collapse;
    font-size: var(--fs-body);
  }
  .tbl th,
  .tbl td {
    padding: var(--space-2) var(--space-4);
    border-bottom: 1px solid var(--border);
    white-space: nowrap;
  }
  .tbl thead th {
    font-size: var(--fs-eyebrow);
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--text-faint);
    font-weight: 600;
  }
  .tbl tbody tr:last-child td {
    border-bottom: none;
  }
  .tbl .l {
    text-align: left;
  }
  .tbl .r {
    text-align: right;
  }
  td.l {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .num {
    font-variant-numeric: tabular-nums;
    font-family: var(--font-mono);
  }

  .swatch {
    width: 10px;
    height: 10px;
    flex: 0 0 auto;
    border-radius: 2px;
    background: var(--c);
  }

  .chart {
    padding: var(--space-3) var(--space-4) var(--space-4);
  }
  svg {
    display: block;
  }
  .grid line {
    stroke: var(--border);
    stroke-width: 1;
  }
  .axis-line {
    stroke: var(--border-strong);
    stroke-width: 1;
  }
  .axis {
    fill: var(--text-faint);
    font-size: var(--fs-eyebrow);
    font-family: var(--font-mono);
  }
  .axis.r {
    text-anchor: end;
  }
  .axis.c {
    text-anchor: middle;
  }
  .legend {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-3);
    padding-top: var(--space-3);
  }
  .leg {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    font-size: var(--fs-caption);
    color: var(--text-muted);
  }
</style>
