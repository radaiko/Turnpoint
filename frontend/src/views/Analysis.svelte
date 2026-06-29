<script lang="ts">
  import { onMount } from "svelte";
  import { get } from "svelte/store";
  import { analysis, analyzing, runAnalysis, recomputeZones } from "$lib/stores/analysis";
  import { ui } from "$lib/stores/ui";
  import { num, intStr } from "$lib/format";
  import FitChart from "$lib/charts/FitChart.svelte";
  import Button from "$lib/components/Button.svelte";
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
                <dt>HR</dt>
                <dd class="mono num">{intStr(anc.val?.heartRate)}</dd>
              </div>
              <div>
                <dt>Pace</dt>
                <dd class="mono num">{anc.val?.pace || "—"}</dd>
              </div>
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
