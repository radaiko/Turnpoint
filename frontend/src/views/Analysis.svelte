<script lang="ts">
  import { onMount } from "svelte";
  import { get } from "svelte/store";
  import { analysis, analyzing, runAnalysis, recomputeZones } from "$lib/stores/analysis";
  import { config, markerOptions, applyConfig, resetConfig } from "$lib/stores/config";
  import { ui } from "$lib/stores/ui";
  import { toast } from "$lib/stores/toast";
  import { num, intStr } from "$lib/format";
  import FitChart from "$lib/charts/FitChart.svelte";
  import Button from "$lib/components/Button.svelte";
  import Select from "$lib/components/Select.svelte";
  import Field from "$lib/components/Field.svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";

  onMount(() => {
    const u = get(ui);
    if (!get(analysis) && u.activeTestId) runAnalysis(u.activeTestId);
  });

  function run() {
    if ($ui.activeTestId) runAnalysis($ui.activeTestId);
  }

  function onAnchorDrag(lt1: number, lt2: number) {
    if ($ui.activeTestId) recomputeZones($ui.activeTestId, lt1, lt2);
  }

  function warnColor(severity: string): string {
    if (severity === "warn" || severity === "warning") return "var(--warn)";
    if (severity === "error" || severity === "danger") return "var(--danger)";
    return "var(--text-muted)";
  }

  // --- Configuration panel (FR-D2/F2) ---
  let configOpen = false;

  const fitOptions = [
    { value: "poly3", label: "Polynomial (3)" },
    { value: "exp", label: "Exponential" },
    { value: "spline", label: "Spline" },
  ];

  const isObla = (name: string) => name.startsWith("OBLA");
  const isBsln = (name: string) => name.startsWith("Bsln");

  // Default OBLA concentration / baseline delta derive from the marker name
  // (e.g. "OBLA 4.0" → 4.0, "Bsln+1.5" → 1.5) when no override is stored.
  function numFromName(name: string): number {
    const m = name.match(/\d+(?:\.\d+)?/);
    return m ? parseFloat(m[0]) : 0;
  }

  function toggleMarker(name: string) {
    config.update((c) => {
      if (!c) return c;
      const enabled = c.enabledMarkers.includes(name)
        ? c.enabledMarkers.filter((n) => n !== name)
        : [...c.enabledMarkers, name];
      return { ...c, enabledMarkers: enabled } as typeof c;
    });
  }

  function setParam(name: string, key: "oblaConc" | "baselineDelta", e: Event) {
    const v = parseFloat((e.target as HTMLInputElement).value);
    if (!isFinite(v)) return;
    config.update((c) => {
      if (!c) return c;
      const mp = { ...c.methodParams };
      const cur = mp[name] ?? { oblaConc: numFromName(name), baselineDelta: numFromName(name) };
      mp[name] = { ...cur, [key]: v } as any;
      return { ...c, methodParams: mp } as typeof c;
    });
  }

  async function apply() {
    if (!$ui.activeTestId) return;
    await applyConfig($ui.activeTestId);
    toast("Configuration applied", "ok");
  }

  async function reset() {
    if (!$ui.activeTestId) return;
    await resetConfig($ui.activeTestId);
    toast("Reset to defaults", "ok");
  }

  // Markers exposing an editable concentration / baseline-delta parameter.
  $: advancedMarkers = $markerOptions.filter((o) => isObla(o.name) || isBsln(o.name));

  // Convenience alias for the current DTO.
  $: a = $analysis;
</script>

<div class="analysis">
  {#if a}
    <header class="head">
      <div>
        <span class="eyebrow mono">{a.sport} · {a.unit}</span>
        <h2>Analysis</h2>
      </div>
      {#if $analyzing}
        <span class="analyzing mono">Analyzing…</span>
      {/if}
    </header>

    <div class="chart-panel">
      <FitChart data={a} {onAnchorDrag} />
    </div>

    <!-- Configuration (FR-D2/F2): collapsible, default collapsed -->
    <section class="config-card">
      <button
        class="config-head"
        type="button"
        aria-expanded={configOpen}
        on:click={() => (configOpen = !configOpen)}
      >
        <div class="config-title">
          <span class="chev" class:open={configOpen}>▸</span>
          <h3>Configuration</h3>
        </div>
        <span class="config-sub mono">{$config?.profileName ?? "—"}</span>
      </button>

      {#if configOpen}
        {#if $config}
          <div class="config-body">
            <div class="config-row">
              <Select label="Display fit" options={fitOptions} bind:value={$config.displayFit} />
              <label class="toggle">
                <input type="checkbox" bind:checked={$config.includeBaselineInFit} />
                <span>Include baseline in fit</span>
              </label>
            </div>

            <div class="config-section">
              <span class="config-label">Methods</span>
              <div class="markers-list">
                {#each $markerOptions as o (o.name)}
                  <label class="marker-toggle">
                    <input
                      type="checkbox"
                      checked={$config.enabledMarkers.includes(o.name)}
                      on:change={() => toggleMarker(o.name)}
                    />
                    <span class="mk-name">{o.name}</span>
                    <span class="mk-fit">{o.fitType}</span>
                  </label>
                {/each}
              </div>
            </div>

            {#if advancedMarkers.length}
              <details class="advanced">
                <summary>Advanced · concentrations &amp; baseline deltas</summary>
                <div class="params-list">
                  {#each advancedMarkers as o (o.name)}
                    <div class="param-row">
                      <span class="param-name mono">{o.name}</span>
                      {#if isObla(o.name)}
                        <Field
                          type="number"
                          step="0.1"
                          suffix="mmol/L"
                          value={$config.methodParams?.[o.name]?.oblaConc ?? numFromName(o.name)}
                          on:change={(e) => setParam(o.name, "oblaConc", e)}
                        />
                      {:else}
                        <Field
                          type="number"
                          step="0.1"
                          suffix="Δ mmol/L"
                          value={$config.methodParams?.[o.name]?.baselineDelta ?? numFromName(o.name)}
                          on:change={(e) => setParam(o.name, "baselineDelta", e)}
                        />
                      {/if}
                    </div>
                  {/each}
                </div>
              </details>
            {/if}
          </div>

          <div class="config-footer">
            <Button variant="primary" on:click={apply}>Apply</Button>
            <Button variant="subtle" on:click={reset}>Reset to defaults</Button>
          </div>
        {:else}
          <div class="config-body">
            <p class="config-empty">Configuration not loaded.</p>
          </div>
        {/if}
      {/if}
    </section>

    <!-- Anchors: IAS (LT1) / IANS (LT2) (FR-D5) -->
    <section class="block">
      <h3>Anchors</h3>
      <div class="anchors">
        {#each [{ label: "IAS · LT1", val: a.lt1 }, { label: "IANS · LT2", val: a.lt2 }] as anc}
          <div class="anchor">
            <div class="anchor-head">
              <span class="eyebrow">{anc.label}</span>
              {#if anc.val?.manual}
                <span class="manual"><span class="dot" />manual</span>
              {/if}
            </div>
            <div class="anchor-marker">{anc.val?.marker ?? "—"}</div>
            <dl class="anchor-grid">
              <div>
                <dt>Intensity</dt>
                <dd class="mono num">{num(anc.val?.intensity, 1)} <span class="unit">{a.unit}</span></dd>
              </div>
              <div>
                <dt>Lactate</dt>
                <dd class="mono num">{num(anc.val?.lactate, 1)} <span class="unit">mmol/L</span></dd>
              </div>
              <div>
                <dt>HR</dt>
                <dd class="mono num">{intStr(anc.val?.heartRate)}</dd>
              </div>
              {#if a.hasPace}
                <div>
                  <dt>Pace</dt>
                  <dd class="mono num">{anc.val?.pace || "—"}</dd>
                </div>
              {/if}
            </dl>
          </div>
        {/each}
      </div>
    </section>

    <!-- Markers results table (FR-D5) -->
    <section class="block">
      <h3>Markers</h3>
      <table class="results">
        <thead>
          <tr>
            <th>Marker</th>
            <th class="r">Intensity ({a.unit})</th>
            <th class="r">Lactate</th>
            <th class="r">HR</th>
            <th class="r">%max</th>
            <th class="r">Pace</th>
            <th>Fit</th>
          </tr>
        </thead>
        <tbody>
          {#each a.markers as m}
            <tr class:noncomp={!m.computable}>
              <td class="marker">
                <span class="mk">{m.marker}</span>
                {#if !m.computable && m.reason}
                  <span class="reason">{m.reason}</span>
                {/if}
              </td>
              <td class="r mono num">{m.computable ? num(m.intensity, 1) : "—"}</td>
              <td class="r mono num">{m.computable ? num(m.lactate, 2) : "—"}</td>
              <td class="r mono num">{m.computable ? intStr(m.heartRate) : "—"}</td>
              <td class="r mono num">{m.computable ? num(m.pctMax, 0) + "%" : "—"}</td>
              <td class="r mono num">{m.computable ? m.pace || "—" : "—"}</td>
              <td>{#if m.fitType}<span class="fit">{m.fitType}</span>{/if}</td>
            </tr>
          {/each}
          {#if !a.markers.length}
            <tr><td colspan="7" class="empty-row">No markers computed.</td></tr>
          {/if}
        </tbody>
      </table>
    </section>

    <!-- Warnings (FR-D5) -->
    {#if a.warnings?.length}
      <section class="block">
        <h3>Warnings</h3>
        <ul class="warnings">
          {#each a.warnings as w}
            <li style="color: {warnColor(w.severity)}">
              <span class="sev">{w.severity}</span>
              {#if w.subject}<span class="subj">{w.subject}</span>{/if}
              <span class="msg">{w.message}</span>
            </li>
          {/each}
        </ul>
      </section>
    {/if}
  {:else}
    <EmptyState
      title="No analysis yet"
      hint="Run the model to fit the lactate curve and compute markers and zones."
    >
      <Button variant="primary" on:click={run} disabled={$analyzing}>
        {$analyzing ? "Analyzing…" : "Run analysis"}
      </Button>
    </EmptyState>
  {/if}
</div>

<style>
  .analysis {
    padding: var(--space-5);
    max-width: 1100px;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
  }
  .head {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
  }
  .eyebrow {
    display: block;
    font-size: var(--fs-eyebrow);
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--text-faint);
  }
  .head h2 {
    margin-top: 2px;
  }
  .analyzing {
    font-size: var(--fs-caption);
    color: var(--text-muted);
    padding: 2px var(--space-3);
    border: 1px solid var(--border);
    border-radius: var(--radius-pill);
    background: var(--surface);
  }

  .chart-panel {
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    background: var(--surface);
    padding: var(--space-3);
    overflow: hidden;
  }

  /* Configuration panel */
  .config-card {
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    background: var(--surface);
    overflow: hidden;
  }
  .config-head {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    padding: var(--space-3) var(--space-4);
    border: none;
    background: transparent;
    text-align: left;
    cursor: pointer;
  }
  .config-head:hover {
    background: var(--surface-2);
  }
  .config-title {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .config-title h3 {
    margin: 0;
    font-size: var(--fs-h3);
    color: var(--text);
  }
  .chev {
    display: inline-block;
    color: var(--text-faint);
    font-size: var(--fs-caption);
    transition: transform 150ms ease;
  }
  .chev.open {
    transform: rotate(90deg);
  }
  .config-sub {
    font-size: var(--fs-caption);
    color: var(--text-faint);
  }
  .config-body {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
    padding: var(--space-4);
    border-top: 1px solid var(--border);
  }
  .config-empty {
    margin: 0;
    font-size: var(--fs-caption);
    color: var(--text-faint);
  }
  .config-row {
    display: flex;
    align-items: flex-end;
    gap: var(--space-5);
    flex-wrap: wrap;
  }
  .config-row :global(.field) {
    min-width: 200px;
  }
  .toggle {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    height: 32px;
    cursor: pointer;
    font-size: var(--fs-body);
    color: var(--text);
  }
  .toggle input,
  .marker-toggle input {
    width: 14px;
    height: 14px;
    accent-color: var(--accent);
  }
  .config-section {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .config-label {
    font-size: var(--fs-label);
    font-weight: 500;
    color: var(--text-muted);
  }
  .markers-list {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
    gap: var(--space-1) var(--space-3);
    max-height: 220px;
    overflow: auto;
    padding: var(--space-2);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    background: var(--surface-2);
  }
  .marker-toggle {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    width: 100%;
    padding: var(--space-1) var(--space-2);
    border-radius: var(--radius-sm);
    cursor: pointer;
    font-size: var(--fs-body);
    color: var(--text);
  }
  .marker-toggle:hover {
    background: var(--surface);
  }
  .mk-name {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .mk-fit {
    font-size: var(--fs-eyebrow);
    text-transform: uppercase;
    letter-spacing: 0.04em;
    padding: 1px 6px;
    border-radius: var(--radius-pill);
    background: var(--inset);
    color: var(--text-faint);
  }
  .advanced {
    border-top: 1px solid var(--border);
    padding-top: var(--space-3);
  }
  .advanced summary {
    cursor: pointer;
    font-size: var(--fs-label);
    font-weight: 500;
    color: var(--text-muted);
  }
  .params-list {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
    gap: var(--space-3);
    margin-top: var(--space-3);
  }
  .param-row {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }
  .param-name {
    font-size: var(--fs-caption);
    color: var(--text-muted);
  }
  .config-footer {
    display: flex;
    gap: var(--space-2);
    padding: var(--space-3) var(--space-4);
    border-top: 1px solid var(--border);
    background: var(--surface-2);
  }

  .block {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }
  .block h3 {
    font-size: var(--fs-h3);
    color: var(--text);
    margin: 0;
  }

  /* Anchors */
  .anchors {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: var(--space-3);
  }
  .anchor {
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    background: var(--surface);
    padding: var(--space-4);
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .anchor-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .anchor-marker {
    font-size: var(--fs-h3);
    font-weight: 600;
  }
  .manual {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    font-size: var(--fs-eyebrow);
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--text-muted);
  }
  .manual .dot {
    width: 6px;
    height: 6px;
    border-radius: var(--radius-pill);
    background: var(--accent);
  }
  .anchor-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: var(--space-3);
    margin: 0;
  }
  .anchor-grid dt {
    font-size: var(--fs-eyebrow);
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--text-faint);
  }
  .anchor-grid dd {
    margin: 2px 0 0;
    font-size: var(--fs-body);
    color: var(--text);
  }
  .unit {
    color: var(--text-faint);
    font-size: var(--fs-caption);
  }

  /* Results table */
  .results {
    width: 100%;
    border-collapse: collapse;
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    overflow: hidden;
    font-size: var(--fs-body);
  }
  .results th,
  .results td {
    padding: var(--space-2) var(--space-3);
    text-align: left;
    border-bottom: 1px solid var(--border);
  }
  .results thead th {
    font-size: var(--fs-eyebrow);
    text-transform: uppercase;
    letter-spacing: 0.04em;
    font-weight: 600;
    color: var(--text-muted);
    background: var(--surface-2);
  }
  .results tbody tr:last-child td {
    border-bottom: none;
  }
  .results tbody tr:hover {
    background: var(--surface);
  }
  .results .r {
    text-align: right;
  }
  .num {
    font-variant-numeric: tabular-nums;
  }
  .marker {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .mk {
    font-weight: 500;
  }
  .reason {
    font-size: var(--fs-caption);
    color: var(--text-faint);
  }
  tr.noncomp {
    color: var(--text-faint);
  }
  tr.noncomp .mk {
    font-weight: 400;
    color: var(--text-muted);
  }
  .fit {
    display: inline-block;
    font-size: var(--fs-eyebrow);
    text-transform: uppercase;
    letter-spacing: 0.04em;
    padding: 2px 8px;
    border-radius: var(--radius-pill);
    background: var(--surface-2);
    color: var(--text-muted);
  }
  .empty-row {
    text-align: center;
    color: var(--text-faint);
    padding: var(--space-5);
  }

  /* Warnings */
  .warnings {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }
  .warnings li {
    display: flex;
    align-items: baseline;
    gap: var(--space-2);
    font-size: var(--fs-caption);
    padding: var(--space-2) var(--space-3);
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    background: var(--surface);
  }
  .warnings .sev {
    font-size: var(--fs-eyebrow);
    text-transform: uppercase;
    letter-spacing: 0.04em;
    font-weight: 600;
  }
  .warnings .subj {
    font-family: var(--font-mono);
    color: var(--text-muted);
  }
  .warnings .msg {
    color: var(--text);
  }
</style>
