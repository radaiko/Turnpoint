<script lang="ts">
  import { onMount } from "svelte";
  import { App, type Test, type AnchorDTO } from "$lib/api";
  import { analysis } from "$lib/stores/analysis";
  import { ui } from "$lib/stores/ui";
  import { toast } from "$lib/stores/toast";
  import { num, intStr, ZONE_COLORS } from "$lib/format";
  import Button from "$lib/components/Button.svelte";
  import Field from "$lib/components/Field.svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import TemporalChart from "$lib/charts/TemporalChart.svelte";

  let test: Test | null = null;

  // Reference to the printable sheet — used to locate the chart <svg> for export.
  let sheetEl: HTMLElement;

  // --- Report settings (on-screen only; not printed) --------------------
  let header = "";
  let footer = "";
  let commentary = "";
  let logoUrl = "";

  // Block include / omit / reorder. Defaults match reference pages 2 & 3:
  // cover / remarks / evaluation off, the rest on.
  let blocks: { id: string; label: string; visible: boolean }[] = [
    { id: "cover", label: "Cover & athlete", visible: false },
    { id: "remarks", label: "Remarks", visible: false },
    { id: "rawTable", label: "Raw data table", visible: true },
    { id: "chart", label: "Temporal chart", visible: true },
    { id: "thresholds", label: "Threshold table", visible: true },
    { id: "zones", label: "Training zones", visible: true },
    { id: "evaluation", label: "Evaluation", visible: false },
  ];

  function toggleBlock(id: string) {
    blocks = blocks.map((b) => (b.id === id ? { ...b, visible: !b.visible } : b));
  }
  function moveBlock(i: number, dir: -1 | 1) {
    const j = i + dir;
    if (j < 0 || j >= blocks.length) return;
    const next = blocks.slice();
    [next[i], next[j]] = [next[j], next[i]];
    blocks = next;
  }

  function onLogoChange(e: Event) {
    const input = e.target as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = () => {
      logoUrl = typeof reader.result === "string" ? reader.result : "";
    };
    reader.onerror = () => toast("Could not read image", "danger");
    reader.readAsDataURL(file);
  }
  function clearLogo() {
    logoUrl = "";
  }

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

  // Raw measurement rows — zip lactate (rawPoints) with HR (hrPoints) by index.
  $: rawHasHr = ($analysis?.hrPoints ?? []).length > 0;
  $: rawRows = ($analysis?.rawPoints ?? []).map((p, i) => ({
    intensity: p.x,
    lactate: p.y,
    hr: $analysis?.hrPoints?.[i]?.y,
  }));

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

  // --- generic download helper ------------------------------------------
  function downloadBlob(blob: Blob, filename: string) {
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
  }

  // Locate the rendered TemporalChart <svg> inside the printable sheet.
  function findChartSvg(): SVGSVGElement | null {
    const root: ParentNode = sheetEl ?? document;
    return root.querySelector("svg") as SVGSVGElement | null;
  }

  function exportChartSvg() {
    const svg = findChartSvg();
    if (!svg) {
      toast("No chart to export — enable the Temporal chart block", "warn");
      return;
    }
    const str = new XMLSerializer().serializeToString(svg);
    const blob = new Blob([str], { type: "image/svg+xml;charset=utf-8" });
    downloadBlob(blob, "chart.svg");
    toast("Chart SVG exported", "ok");
  }

  function exportChartPng() {
    const svg = findChartSvg();
    if (!svg) {
      toast("No chart to export — enable the Temporal chart block", "warn");
      return;
    }
    const str = new XMLSerializer().serializeToString(svg);
    const rect = svg.getBoundingClientRect();
    const w = Math.round(rect.width || svg.clientWidth || 820);
    const h = Math.round(rect.height || svg.clientHeight || 360);
    const scale = 2; // ~2x for crisp raster output

    const svgUrl =
      "data:image/svg+xml;charset=utf-8," + encodeURIComponent(str);
    const img = new Image();
    img.onload = () => {
      const canvas = document.createElement("canvas");
      canvas.width = w * scale;
      canvas.height = h * scale;
      const ctx = canvas.getContext("2d");
      if (!ctx) {
        toast("Export failed", "danger");
        return;
      }
      ctx.scale(scale, scale);
      ctx.fillStyle = "#ffffff";
      ctx.fillRect(0, 0, w, h);
      ctx.drawImage(img, 0, 0, w, h);
      canvas.toBlob((blob) => {
        if (blob) {
          downloadBlob(blob, "chart.png");
          toast("Chart PNG exported", "ok");
        } else {
          toast("Export failed", "danger");
        }
      }, "image/png");
    };
    img.onerror = () => toast("Export failed", "danger");
    img.src = svgUrl;
  }

  async function exportCsv() {
    const id = $ui.activeTestId;
    if (!id) return;
    try {
      const csv = await App.ExportCSV(id);
      const blob = new Blob([csv], { type: "text/csv;charset=utf-8" });
      downloadBlob(blob, "test.csv");
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
    <div class="toolbar no-print">
      <span class="eyebrow">Report preview</span>
      <div class="tools">
        <Button variant="subtle" on:click={exportChartPng}>Export PNG</Button>
        <Button variant="subtle" on:click={exportChartSvg}>Export SVG</Button>
        <Button variant="subtle" on:click={exportCsv}>Export CSV</Button>
        <Button variant="primary" on:click={printReport}>Print / Save PDF</Button>
      </div>
    </div>

    <div class="layout">
      <!-- (S) Report settings — on-screen only, never printed. -->
      <aside class="settings no-print">
        <h3 class="settings-title">Report settings</h3>

        <div class="settings-group">
          <Field label="Header text" type="text" bind:value={header} placeholder="e.g. Performance diagnostics" />
          <Field label="Footer text" type="text" bind:value={footer} placeholder="e.g. © 2026 · Lab name" />
        </div>

        <div class="settings-group">
          <span class="lbl">Logo</span>
          <div class="logo-row">
            <input
              class="file"
              type="file"
              accept="image/*"
              on:change={onLogoChange}
            />
            {#if logoUrl}
              <Button variant="ghost" on:click={clearLogo}>Clear</Button>
            {/if}
          </div>
          {#if logoUrl}
            <img class="logo-preview" src={logoUrl} alt="Logo preview" />
          {/if}
        </div>

        <div class="settings-group commentary">
          <span class="lbl">Commentary</span>
          <textarea
            rows="5"
            bind:value={commentary}
            placeholder="Free-text evaluation of the test…"
          ></textarea>
        </div>

        <div class="settings-group">
          <span class="lbl">Report blocks</span>
          <ul class="block-list">
            {#each blocks as b, i (b.id)}
              <li class="block-row" class:off={!b.visible}>
                <label class="block-toggle">
                  <input
                    type="checkbox"
                    checked={b.visible}
                    on:change={() => toggleBlock(b.id)}
                  />
                  <span>{b.label}</span>
                </label>
                <span class="reorder">
                  <button
                    type="button"
                    title="Move up"
                    disabled={i === 0}
                    on:click={() => moveBlock(i, -1)}>↑</button
                  >
                  <button
                    type="button"
                    title="Move down"
                    disabled={i === blocks.length - 1}
                    on:click={() => moveBlock(i, 1)}>↓</button
                  >
                </span>
              </li>
            {/each}
          </ul>
        </div>
      </aside>

      <article class="sheet" bind:this={sheetEl}>
        <!-- (A) Header band — always present; carries logo + editable header. -->
        <header class="band">
          <div class="band-lead">
            {#if logoUrl}
              <img class="logo" src={logoUrl} alt="Logo" />
            {/if}
            <div>
              {#if header}<p class="band-header">{header}</p>{/if}
              <h1>Lactate Threshold Report</h1>
              <p class="subtitle mono">
                {testDate} · {sport}{unit ? ` · ${unit}` : ""}
              </p>
            </div>
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

        <!-- Blocks render in `blocks` order; only those toggled visible. -->
        {#each blocks as block (block.id)}
          {#if block.visible}
            {#if block.id === "cover"}
              <!-- Cover & athlete -->
              <section class="block cover">
                <h2>Cover</h2>
                {#if header}<p class="cover-title">{header}</p>{/if}
                <dl class="meta">
                  <div><dt>Date</dt><dd class="mono">{testDate}</dd></div>
                  <div><dt>Sport</dt><dd>{sport}</dd></div>
                  {#if unit}<div><dt>Unit</dt><dd class="mono">{unit}</dd></div>{/if}
                </dl>
              </section>
            {:else if block.id === "remarks"}
              <!-- Remarks -->
              <section class="block">
                <h2>Remarks</h2>
                {#if test?.pretestNote}
                  <p class="note"><span class="lbl">Pre-test</span> {test.pretestNote}</p>
                {/if}
                {#if test?.remarks}
                  <p class="note"><span class="lbl">Remarks</span> {test.remarks}</p>
                {/if}
                {#if !test?.pretestNote && !test?.remarks}
                  <p class="note muted">No remarks recorded.</p>
                {/if}
              </section>
            {:else if block.id === "rawTable"}
              <!-- Raw data table -->
              <section class="block">
                <h2>Raw data</h2>
                {#if rawRows.length}
                  <table class="grid">
                    <thead>
                      <tr>
                        <th class="left">#</th>
                        <th class="r">{unit || "Intensity"}</th>
                        <th class="r">mmol/L</th>
                        {#if rawHasHr}<th class="r">bpm</th>{/if}
                      </tr>
                    </thead>
                    <tbody>
                      {#each rawRows as r, i}
                        <tr>
                          <td class="left mono num">{i + 1}</td>
                          <td class="r mono num">{num(r.intensity, 1)}</td>
                          <td class="r mono num">{num(r.lactate, 2)}</td>
                          {#if rawHasHr}<td class="r mono num">{intStr(r.hr)}</td>{/if}
                        </tr>
                      {/each}
                    </tbody>
                  </table>
                {:else}
                  <p class="note muted">No raw measurement points.</p>
                {/if}
              </section>
            {:else if block.id === "chart"}
              <!-- Temporal chart -->
              <section class="block chart-block">
                <h2>Lactate curve</h2>
                <p class="note">
                  Measured lactate and heart rate plotted against intensity, with the fitted
                  lactate curve and the determined threshold anchors (IAS / IANS).
                </p>
                <div class="chart">
                  <TemporalChart data={$analysis} />
                </div>
              </section>
            {:else if block.id === "thresholds"}
              <!-- Threshold results table -->
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
            {:else if block.id === "zones"}
              <!-- Training-zones table -->
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
            {:else if block.id === "evaluation"}
              <!-- Evaluation / commentary -->
              <section class="block">
                <h2>Evaluation</h2>
                {#if commentary.trim()}
                  <p class="commentary-text">{commentary}</p>
                {:else}
                  <p class="note muted">No commentary entered.</p>
                {/if}
              </section>
            {/if}
          {/if}
        {/each}

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

        {#if footer}
          <footer class="sheet-footer">{footer}</footer>
        {/if}
      </article>
    </div>
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

  /* settings + sheet side by side on screen */
  .layout {
    display: flex;
    align-items: flex-start;
    justify-content: center;
    gap: var(--space-5);
    padding: var(--space-6) var(--space-5);
  }

  /* (S) Report settings panel */
  .settings {
    position: sticky;
    top: calc(56px + var(--space-5));
    flex: none;
    width: 260px;
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
    padding: var(--space-4);
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
  }
  .settings-title {
    font-size: var(--fs-h3);
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--text-muted);
  }
  .settings-group {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .lbl {
    font-size: var(--fs-label);
    font-weight: 500;
    color: var(--text-muted);
  }
  .logo-row {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .file {
    flex: 1;
    min-width: 0;
    font-size: var(--fs-caption);
    color: var(--text-muted);
  }
  .logo-preview {
    max-width: 100%;
    max-height: 64px;
    object-fit: contain;
    align-self: flex-start;
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    background: var(--surface-2);
    padding: var(--space-1);
  }
  .commentary textarea {
    width: 100%;
    resize: vertical;
    min-height: 88px;
    padding: var(--space-2) var(--space-3);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    background: var(--surface);
    color: var(--text);
    font-family: inherit;
    font-size: var(--fs-body);
    line-height: 1.5;
  }
  .commentary textarea:focus-within,
  .commentary textarea:focus {
    outline: none;
    border-color: var(--accent);
  }

  .block-list {
    list-style: none;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .block-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-2);
    padding: var(--space-1) var(--space-2);
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    background: var(--surface);
  }
  .block-row.off {
    background: var(--surface-2);
    color: var(--text-faint);
  }
  .block-toggle {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    font-size: var(--fs-caption);
    min-width: 0;
    cursor: pointer;
  }
  .block-toggle span {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .block-toggle input {
    accent-color: var(--accent);
    flex: none;
  }
  .reorder {
    display: inline-flex;
    gap: 2px;
    flex: none;
  }
  .reorder button {
    width: 22px;
    height: 22px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    background: var(--surface);
    color: var(--text-muted);
    font-size: var(--fs-caption);
    line-height: 1;
    cursor: pointer;
    transition: background 150ms ease, border-color 150ms ease;
  }
  .reorder button:hover:not(:disabled) {
    background: var(--surface-2);
    border-color: var(--border-strong);
  }
  .reorder button:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  /* The page sheet. */
  .sheet {
    flex: 0 1 820px;
    max-width: 820px;
    margin: 0;
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
  .band-lead {
    display: flex;
    align-items: flex-start;
    gap: var(--space-4);
  }
  .logo {
    max-width: 96px;
    max-height: 64px;
    object-fit: contain;
    flex: none;
  }
  .band-header {
    font-size: var(--fs-caption);
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--text-muted);
    margin-bottom: var(--space-1);
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
  .note.muted {
    color: var(--text-faint);
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

  /* cover block */
  .cover-title {
    font-size: var(--fs-h1);
    font-weight: 600;
    letter-spacing: -0.01em;
    color: var(--text);
  }
  .meta {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-5);
  }
  .meta div {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .meta dt {
    font-size: var(--fs-eyebrow);
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--text-faint);
  }
  .meta dd {
    font-size: var(--fs-body);
    color: var(--text);
  }

  /* evaluation / commentary */
  .commentary-text {
    font-size: var(--fs-body);
    color: var(--text);
    line-height: 1.6;
    white-space: pre-wrap;
    max-width: 70ch;
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

  /* footer */
  .sheet-footer {
    margin-top: var(--space-3);
    padding-top: var(--space-3);
    border-top: 1px solid var(--border-strong);
    font-size: var(--fs-caption);
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
    /* Hide app chrome, toolbar and settings; reveal only the report sheet. */
    :global(body) * {
      visibility: hidden;
    }
    .sheet,
    .sheet * {
      visibility: visible;
    }
    .no-print {
      display: none !important;
    }
    .toolbar {
      display: none !important;
    }
    .layout {
      display: block;
      padding: 0;
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
    .sheet-footer {
      border-top: 1px solid #000;
      color: #333;
    }
  }
</style>
