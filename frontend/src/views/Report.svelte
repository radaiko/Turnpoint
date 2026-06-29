<script lang="ts">
  import { onMount } from "svelte";
  import { App, type Test, type AnchorDTO } from "$lib/api";
  import { analysis } from "$lib/stores/analysis";
  import { ui } from "$lib/stores/ui";
  import { toast } from "$lib/stores/toast";
  import { num, intStr, ZONE_COLORS } from "$lib/format";
  import Button from "$lib/components/Button.svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import TemporalChart from "$lib/charts/TemporalChart.svelte";

  let test: Test | null = null;

  onMount(async () => {
    const id = $ui.activeTestId;
    if (id) {
      try {
        test = await App.GetTest(id);
      } catch {
        test = null;
      }
    }
  });

  // Derived view flags — columns appear only when the data carries them.
  $: unit = $analysis?.unit ?? "";
  $: sport = test?.sport ?? $analysis?.sport ?? "—";
  $: testDate = test?.testDate ?? "—";
  $: showPace = !!$analysis?.hasPace;
  $: showKcal = !!$analysis?.markers?.some((m) => m.hasKcal);
  $: anchors = [
    { tag: "IAS", a: $analysis?.lt1 },
    { tag: "IANS", a: $analysis?.lt2 },
  ] as { tag: string; a: AnchorDTO | undefined }[];

  const ZONE_KEYS = ["REKOM", "GA1", "GA2", "EB", "SB"];
  function zoneColor(label: string): string {
    const up = (label ?? "").toUpperCase();
    const key = ZONE_KEYS.find((k) => up.includes(k));
    return key ? ZONE_COLORS[key] : "var(--border-strong)";
  }

  // %HRmax may arrive as a fraction (0–1) or a percentage (0–100); normalise.
  function pct(v: number | null | undefined): string {
    if (v === null || v === undefined || !isFinite(v) || v <= 0) return "—";
    return num(v <= 1.5 ? v * 100 : v, 0) + "%";
  }

  function rng(lo: number, hi: number, p = 1): string {
    return `${num(lo, p)}–${num(hi, p)}`;
  }
  function rngInt(lo: number, hi: number): string {
    return `${intStr(lo)}–${intStr(hi)}`;
  }
  function rngStr(lo: string, hi: string): string {
    if (!lo && !hi) return "—";
    return `${lo || "—"}–${hi || "—"}`;
  }

  function printReport() {
    window.print();
  }

  async function exportCsv() {
    const id = $ui.activeTestId;
    if (!id) return;
    try {
      const csv = await App.ExportCSV(id);
      const blob = new Blob([csv], { type: "text/csv;charset=utf-8" });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = "test.csv";
      document.body.appendChild(a);
      a.click();
      a.remove();
      URL.revokeObjectURL(url);
      toast("CSV exported", "ok");
    } catch (e) {
      toast("Export failed", "danger");
    }
  }
</script>

<div class="report">
  {#if !$analysis}
    <EmptyState
      title="No analysis yet"
      hint="Run the analysis from the Analysis stage to produce a printable report."
    />
  {:else}
    <div class="toolbar">
      <span class="eyebrow">Report preview</span>
      <div class="tools">
        <Button variant="subtle" on:click={exportCsv}>Export CSV</Button>
        <Button variant="primary" on:click={printReport}>Print / Save PDF</Button>
      </div>
    </div>

    <article class="sheet">
      <!-- (A) Header band -->
      <header class="band">
        <div>
          <h1>Lactate Threshold Report</h1>
          <p class="subtitle mono">{testDate} · {sport}{unit ? ` · ${unit}` : ""}</p>
        </div>
        <div class="anchors">
          {#each anchors as { tag, a } (tag)}
            <div class="anchor">
              <span class="eyebrow">{tag}</span>
              <span class="big mono num">{num(a?.intensity, 1)}<small>{unit}</small></span>
              <span class="anchor-meta mono">
                {num(a?.lactate, 2)} mmol · {intStr(a?.heartRate)} bpm{#if showPace && a?.pace} · {a.pace}{/if}
              </span>
            </div>
          {/each}
        </div>
      </header>

      <!-- (B) Raw-data note + chart -->
      <section class="block chart-block">
        <h2>Lactate curve</h2>
        <p class="note">
          Measured lactate and heart rate plotted against intensity, with the fitted
          lactate curve and the determined threshold anchors (IAS / IANS).
        </p>
        {#if test?.pretestNote}
          <p class="note"><span class="lbl">Pre-test</span> {test.pretestNote}</p>
        {/if}
        {#if test?.remarks}
          <p class="note"><span class="lbl">Remarks</span> {test.remarks}</p>
        {/if}
        <div class="chart">
          <TemporalChart data={$analysis} />
        </div>
      </section>

      <!-- (C) Threshold results table -->
      <section class="block">
        <h2>Threshold results</h2>
        <table class="grid">
          <thead>
            <tr>
              <th class="left">Marker</th>
              <th class="r">{unit || "Intensity"}</th>
              <th class="r">mmol/L</th>
              <th class="r">bpm</th>
              <th class="r">%HRmax</th>
              {#if showPace}<th class="r">Pace</th>{/if}
              {#if showKcal}<th class="r">kcal/h</th>{/if}
            </tr>
          </thead>
          <tbody>
            {#each $analysis.markers as m (m.marker)}
              <tr class:dim={!m.computable} title={m.computable ? "" : m.reason ?? ""}>
                <td class="left">{m.marker}</td>
                <td class="r mono num">{m.computable ? num(m.intensity, 1) : "—"}</td>
                <td class="r mono num">{m.computable ? num(m.lactate, 2) : "—"}</td>
                <td class="r mono num">{m.computable ? intStr(m.heartRate) : "—"}</td>
                <td class="r mono num">{m.computable ? pct(m.pctMax) : "—"}</td>
                {#if showPace}<td class="r mono num">{m.computable ? m.pace || "—" : "—"}</td>{/if}
                {#if showKcal}<td class="r mono num">{m.computable && m.hasKcal ? intStr(m.kcalPerHr) : "—"}</td>{/if}
              </tr>
            {/each}
          </tbody>
        </table>
      </section>

      <!-- (D) Training-zones table -->
      <section class="block">
        <h2>Training zones</h2>
        <table class="grid">
          <thead>
            <tr>
              <th class="left">Zone</th>
              <th class="r">{unit || "Intensity"}</th>
              <th class="r">bpm</th>
              <th class="r">mmol/L</th>
              {#if showPace}<th class="r">Pace</th>{/if}
            </tr>
          </thead>
          <tbody>
            {#each $analysis.zones as z (z.index)}
              <tr>
                <td class="left">
                  <span class="chip">
                    <span class="swatch" style="background:{zoneColor(z.label)}"></span>
                    {z.label}
                  </span>
                </td>
                <td class="r mono num">{rng(z.intensityLow, z.intensityHigh, 1)}</td>
                <td class="r mono num">{rngInt(z.hrLow, z.hrHigh)}</td>
                <td class="r mono num">{rng(z.lactateLow, z.lactateHigh, 2)}</td>
                {#if showPace}<td class="r mono num">{rngStr(z.paceLow, z.paceHigh)}</td>{/if}
              </tr>
            {/each}
          </tbody>
        </table>
      </section>

      {#if $analysis.warnings?.length}
        <section class="block notes">
          <h2>Notes</h2>
          <ul>
            {#each $analysis.warnings as w}
              <li>
                <span
                  class="dot"
                  style="background:{w.severity === 'error'
                    ? 'var(--danger)'
                    : w.severity === 'warn'
                      ? 'var(--warn)'
                      : 'var(--text-faint)'}"
                ></span>
                <span class="w-subject">{w.subject}</span>
                <span class="w-msg">{w.message}</span>
              </li>
            {/each}
          </ul>
        </section>
      {/if}
    </article>
  {/if}
</div>

<style>
  .report {
    height: 100%;
    overflow: auto;
    background: var(--bg);
  }

  /* Toolbar — on-screen only. */
  .toolbar {
    position: sticky;
    top: 0;
    z-index: 2;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-4);
    padding: var(--space-3) var(--space-5);
    background: var(--surface);
    border-bottom: 1px solid var(--border);
  }
  .tools {
    display: flex;
    gap: var(--space-2);
  }

  /* The page sheet. */
  .sheet {
    max-width: 820px;
    margin: var(--space-6) auto;
    background: var(--surface);
    color: var(--text);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    padding: var(--space-7);
    display: flex;
    flex-direction: column;
    gap: var(--space-6);
  }

  /* (A) header band */
  .band {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--space-5);
    padding-bottom: var(--space-4);
    border-bottom: 1px solid var(--border-strong);
  }
  .subtitle {
    margin-top: var(--space-1);
    font-size: var(--fs-caption);
    color: var(--text-muted);
    text-transform: capitalize;
  }
  .anchors {
    display: flex;
    gap: var(--space-3);
  }
  .anchor {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 120px;
    padding: var(--space-3);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    background: var(--surface-2);
  }
  .big {
    font-size: var(--fs-h1);
    font-weight: 600;
    letter-spacing: -0.01em;
  }
  .big small {
    font-size: var(--fs-caption);
    font-weight: 500;
    color: var(--text-muted);
    margin-left: 2px;
  }
  .anchor-meta {
    font-size: var(--fs-eyebrow);
    color: var(--text-muted);
  }

  /* blocks */
  .block {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    break-inside: avoid;
  }
  .block h2 {
    font-size: var(--fs-h3);
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--text-muted);
  }
  .note {
    font-size: var(--fs-caption);
    color: var(--text-muted);
    max-width: 70ch;
  }
  .note .lbl {
    font-weight: 600;
    color: var(--text);
    margin-right: var(--space-1);
  }
  .chart {
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    padding: var(--space-3);
    background: var(--surface);
  }

  /* tables */
  .grid {
    width: 100%;
    border-collapse: collapse;
    font-size: var(--fs-caption);
  }
  .grid th,
  .grid td {
    padding: var(--space-2) var(--space-3);
    border-bottom: 1px solid var(--border);
    text-align: right;
    white-space: nowrap;
  }
  .grid th {
    font-size: var(--fs-eyebrow);
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--text-faint);
    font-weight: 600;
    border-bottom: 1px solid var(--border-strong);
  }
  .grid th.left,
  .grid td.left {
    text-align: left;
  }
  .grid tbody tr:last-child td {
    border-bottom: none;
  }
  .grid tr.dim td {
    color: var(--text-faint);
  }

  .chip {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    font-weight: 500;
    color: var(--text);
  }
  .swatch {
    width: 10px;
    height: 10px;
    border-radius: var(--radius-sm);
    flex: none;
  }

  /* notes / warnings */
  .notes ul {
    list-style: none;
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .notes li {
    display: flex;
    align-items: baseline;
    gap: var(--space-2);
    font-size: var(--fs-caption);
  }
  .dot {
    width: 8px;
    height: 8px;
    border-radius: var(--radius-pill);
    flex: none;
    transform: translateY(1px);
  }
  .w-subject {
    font-weight: 600;
  }
  .w-msg {
    color: var(--text-muted);
  }

  /* ---------- print ---------- */
  @page {
    size: A4 portrait;
    margin: 16mm;
  }
  @media print {
    :global(body) {
      background: #fff;
    }
    /* Hide app chrome and toolbar; reveal only the report sheet. */
    :global(body) * {
      visibility: hidden;
    }
    .sheet,
    .sheet * {
      visibility: visible;
    }
    .toolbar {
      display: none !important;
    }
    .report {
      overflow: visible;
      background: #fff;
    }
    .sheet {
      position: absolute;
      left: 0;
      top: 0;
      width: 100%;
      max-width: none;
      margin: 0;
      padding: 0;
      border: none;
      border-radius: 0;
      background: #fff;
      color: #000;
      gap: 18px;
    }
    .band {
      border-bottom: 1px solid #000;
    }
    .chart,
    .anchor {
      border-color: #000;
      background: #fff;
    }
    .block {
      break-inside: avoid;
      page-break-inside: avoid;
    }
    .grid,
    .grid tr {
      break-inside: avoid;
      page-break-inside: avoid;
    }
    .grid th,
    .grid td {
      border-bottom: 1px solid #000;
      color: #000;
    }
    .subtitle,
    .note,
    .w-msg,
    .grid tr.dim td {
      color: #333;
    }
  }
</style>
