<script lang="ts">
  import { analysis } from "$lib/stores/analysis";
  import {
    config,
    markerOptions,
    profileOptions,
    applyConfig,
    setAnchorMethod,
    setAnchorIntensity,
  } from "$lib/stores/config";
  import { ui } from "$lib/stores/ui";
  import { ZONE_COLORS, num } from "$lib/format";
  import Button from "$lib/components/Button.svelte";
  import Field from "$lib/components/Field.svelte";
  import Select from "$lib/components/Select.svelte";
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

  // The two threshold anchors, rendered with identical controls (FR-Z3).
  const anchors = [
    { which: "lt1", name: "IAS", full: "aerobic threshold" },
    { which: "lt2", name: "IANS", full: "anaerobic threshold" },
  ] as const;

  // Training profile selection drives zone derivation (FR-Z5). Mutate $config
  // immutably, then re-run the analysis so the table reflects the new profile.
  function onProfile(e: Event) {
    const id = $ui.activeTestId;
    const name = (e.target as HTMLSelectElement).value;
    config.update((c) => (c ? { ...c, profileName: name } : c));
    if (id != null) applyConfig(id);
  }

  // Pick the marker method for an anchor; clears any manual override (FR-Z3).
  function onMethod(which: "lt1" | "lt2", e: Event) {
    const id = $ui.activeTestId;
    if (id == null) return;
    setAnchorMethod(id, which, (e.target as HTMLSelectElement).value);
  }

  // Type an exact intensity for an anchor and apply on change/Enter (FR-Z3).
  function onIntensity(which: "lt1" | "lt2", e: Event) {
    const id = $ui.activeTestId;
    if (id == null) return;
    const v = (e.target as HTMLInputElement).valueAsNumber;
    if (!isFinite(v)) return;
    setAnchorIntensity(id, which, v);
  }

  // Revert a manual anchor back to its method value (drops the override).
  function autoAnchor(which: "lt1" | "lt2") {
    const id = $ui.activeTestId;
    if (id == null || !$config) return;
    const marker = which === "lt1" ? $config.lt1Anchor : $config.lt2Anchor;
    setAnchorMethod(id, which, marker);
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
          {num($analysis.lt1.lactate, 1)} mmol/L · {hr($analysis.lt1.heartRate)} bpm{#if $analysis.hasPace} ·
            {pace($analysis.lt1.pace)}{/if}
        </div>
      </div>
      <div class="card">
        <span class="eyebrow">IANS · anaerobic threshold</span>
        <div class="big mono">
          {num($analysis.lt2.intensity, 1)}<span class="u">{$analysis.unit}</span>
        </div>
        <div class="card-meta mono">
          {num($analysis.lt2.lactate, 1)} mmol/L · {hr($analysis.lt2.heartRate)} bpm{#if $analysis.hasPace} ·
            {pace($analysis.lt2.pace)}{/if}
        </div>
      </div>
    </div>

    {#if $config}
      <div class="controls">
        <span class="eyebrow">Calibration</span>

        <div class="profile-row">
          <Select
            label="Training profile"
            value={$config.profileName}
            options={$profileOptions.map((p) => ({
              value: p.name,
              label: `${p.name} · ${p.calibrated ? "calibrated" : "provisional"}`,
            }))}
            on:change={onProfile}
          />
        </div>

        <div class="anchor-grid">
          {#each anchors as a (a.which)}
            {@const anchorVal = a.which === "lt1" ? $config.lt1Anchor : $config.lt2Anchor}
            {@const ad = $analysis[a.which]}
            <div class="anchor-ctl">
              <div class="ctl-head">
                <span class="eyebrow">{a.name} · {a.full}</span>
                {#if ad.manual}<span class="manual-tag">manual</span>{/if}
              </div>
              <div class="ctl-inputs">
                <Select
                  label="Method"
                  value={anchorVal}
                  options={$markerOptions.map((m) => ({ value: m.name, label: m.name }))}
                  on:change={(e) => onMethod(a.which, e)}
                />
                <Field
                  label="Intensity"
                  type="number"
                  step="0.1"
                  suffix={$analysis.unit}
                  value={Math.round(ad.intensity * 10) / 10}
                  on:change={(e) => onIntensity(a.which, e)}
                />
                <Button on:click={() => autoAnchor(a.which)}>Auto</Button>
              </div>
            </div>
          {/each}
        </div>
      </div>
    {/if}

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

  /* profile + anchor controls (FR-Z3/Z5) */
  .controls {
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    background: var(--surface);
    padding: var(--space-4);
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }
  .profile-row {
    max-width: 340px;
  }
  .anchor-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-4);
  }
  .anchor-ctl {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    padding: var(--space-3);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    background: var(--surface-2);
  }
  .ctl-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-2);
  }
  .ctl-inputs {
    display: grid;
    grid-template-columns: 1.4fr 1fr auto;
    gap: var(--space-2);
    align-items: end;
  }
  .manual-tag {
    font-size: var(--fs-eyebrow);
    text-transform: uppercase;
    letter-spacing: 0.04em;
    padding: 2px 8px;
    border-radius: var(--radius-pill);
    background: color-mix(in srgb, var(--warn) 16%, transparent);
    color: var(--warn);
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
