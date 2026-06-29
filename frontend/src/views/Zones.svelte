<script lang="ts">
  import { analysis } from "$lib/stores/analysis";
  import { ZONE_COLORS, num } from "$lib/format";
  import EmptyState from "$lib/components/EmptyState.svelte";

  // "lo–hi" with N decimal places; em-dash placeholder handled by num()
  function range(lo: number, hi: number, places = 1): string {
    return `${num(lo, places)}–${num(hi, places)}`;
  }

  // heart-rate values render rounded to whole bpm
  function hr(v: number | null | undefined): string {
    return v != null && isFinite(v) && v > 0 ? String(Math.round(v)) : "—";
  }
  function hrRange(lo: number, hi: number): string {
    return `${hr(lo)}–${hr(hi)}`;
  }

  // pace strings arrive pre-formatted as mm:ss; blank → em-dash
  function pace(s: string | null | undefined): string {
    return s && s.trim() ? s : "—";
  }
  function paceRange(lo: string, hi: string): string {
    return `${pace(lo)}–${pace(hi)}`;
  }
</script>

{#if !$analysis}
  <EmptyState title="No zones yet" hint="Run the analysis first." />
{:else}
  <div class="wrap">
    <header class="head">
      <h2>Training zones</h2>
      <p class="sub">Anchored on the lactate threshold markers below.</p>
    </header>

    <div class="anchors">
      <div class="card">
        <span class="eyebrow">IAS · aerobic threshold</span>
        <div class="big mono">
          {num($analysis.lt1.intensity, 1)}<span class="u">{$analysis.unit}</span>
        </div>
        <div class="card-meta mono">
          {hr($analysis.lt1.heartRate)} bpm · {pace($analysis.lt1.pace)}
        </div>
      </div>
      <div class="card">
        <span class="eyebrow">IANS · anaerobic threshold</span>
        <div class="big mono">
          {num($analysis.lt2.intensity, 1)}<span class="u">{$analysis.unit}</span>
        </div>
        <div class="card-meta mono">
          {hr($analysis.lt2.heartRate)} bpm · {pace($analysis.lt2.pace)}
        </div>
      </div>
    </div>

    <div class="table">
      <div class="thead">
        <span>Zone</span>
        <span>Intensity ({$analysis.unit})</span>
        <span>Lactate (mmol/L)</span>
        <span>HR (bpm)</span>
        <span>Pace</span>
      </div>
      {#each $analysis.zones as z (z.index)}
        <div class="zrow" style="--chip: {ZONE_COLORS[z.label] ?? 'var(--border-strong)'}">
          <span class="zlabel">{z.label}</span>
          <span class="mono">{range(z.intensityLow, z.intensityHigh, 1)}</span>
          <span class="mono">{range(z.lactateLow, z.lactateHigh, 1)}</span>
          <span class="mono">{hrRange(z.hrLow, z.hrHigh)}</span>
          <span class="mono">{paceRange(z.paceLow, z.paceHigh)}</span>
        </div>
      {/each}
    </div>

    <p class="note">Zones derive from the IANS anchor via the selected training profile.</p>
  </div>
{/if}

<style>
  .wrap {
    max-width: 1000px;
    margin: 0 auto;
    padding: var(--space-6) var(--space-5) var(--space-7);
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
  }

  .head h2 {
    font-size: var(--fs-h2);
    margin: 0;
  }
  .head .sub {
    margin: var(--space-1) 0 0;
    font-size: var(--fs-caption);
    color: var(--text-muted);
  }

  /* anchor stat cards */
  .anchors {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-4);
  }
  .card {
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    background: var(--surface);
    padding: var(--space-4);
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .eyebrow {
    font-size: var(--fs-eyebrow);
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--text-faint);
  }
  .big {
    font-size: var(--fs-display);
    line-height: 1;
    color: var(--text);
    font-weight: 500;
  }
  .big .u {
    font-size: var(--fs-body);
    color: var(--text-muted);
    margin-left: var(--space-2);
    font-weight: 400;
  }
  .card-meta {
    font-size: var(--fs-caption);
    color: var(--text-muted);
  }

  /* zone table */
  .table {
    --grid: 0.8fr 1.3fr 1.2fr 1.1fr 1.1fr;
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    overflow: hidden;
    background: var(--surface);
  }
  .thead {
    display: grid;
    grid-template-columns: var(--grid);
    gap: var(--space-3);
    padding: var(--space-3) var(--space-4);
    border-bottom: 1px solid var(--border);
    background: var(--surface-2);
    font-size: var(--fs-eyebrow);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-faint);
  }
  .zrow {
    display: grid;
    grid-template-columns: var(--grid);
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-3) var(--space-4);
    border-left: 3px solid var(--chip);
    border-bottom: 1px solid var(--border);
    background: color-mix(in srgb, var(--chip) 8%, transparent);
    font-size: var(--fs-body);
  }
  .zrow:last-child {
    border-bottom: none;
  }
  .zlabel {
    font-weight: 600;
    letter-spacing: 0.02em;
    color: var(--text);
  }
  .zrow .mono {
    color: var(--text-muted);
  }

  .note {
    margin: 0;
    font-size: var(--fs-caption);
    color: var(--text-faint);
  }
</style>
