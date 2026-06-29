<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import type { Step } from "$lib/api";
  import { App } from "$lib/api";
  import { mmss, parseMMSS } from "$lib/format";
  import { toast } from "$lib/stores/toast";

  export let steps: Step[] = [];
  export let unit = "km/h";

  const dispatch = createEventDispatcher();
  function changed() {
    steps = steps;
    dispatch("change", steps);
  }

  function blank(order: number): Step {
    return { id: 0, testId: 0, stepOrder: order, intensity: 0, isBaseline: order === 0, excluded: false, aborted: false } as unknown as Step;
  }

  function addRow() {
    steps = [...steps, blank(steps.length)];
    changed();
  }
  function removeRow(i: number) {
    steps = steps.filter((_, idx) => idx !== i).map((s, idx) => ({ ...s, stepOrder: idx }));
    changed();
  }

  function setNum(i: number, key: keyof Step, v: string) {
    const n = v === "" ? undefined : parseFloat(v);
    (steps[i] as any)[key] = isNaN(n as number) ? undefined : n;
    if (key === "intensity") steps[i].isBaseline = steps[i].intensity === 0;
    changed();
  }
  function setTime(i: number, v: string) {
    const sec = parseMMSS(v);
    (steps[i] as any).timePointS = sec ?? undefined;
    changed();
  }

  async function onPaste(e: ClipboardEvent) {
    const text = e.clipboardData?.getData("text") ?? "";
    if (!text.includes("\n") && !text.includes("\t")) return; // single cell paste — let it through
    e.preventDefault();
    try {
      const parsed = await App.ParsePaste(text);
      if (parsed.length) {
        steps = parsed;
        changed();
        toast(`Pasted ${parsed.length} rows`, "ok");
      }
    } catch (err) {
      toast("Could not parse pasted data", "warn");
    }
  }
</script>

<div class="grid" on:paste={onPaste}>
  <div class="head row">
    <span>#</span>
    <span>Intensity [{unit}]</span>
    <span>Time</span>
    <span>HR</span>
    <span>Lactate</span>
    <span>RPE</span>
    <span></span>
  </div>
  {#each steps as s, i (i)}
    <div class="row" class:baseline={s.isBaseline}>
      <span class="idx mono">{s.isBaseline ? "0" : i}</span>
      <input class="num" type="number" step="0.1" value={s.intensity ?? ""} on:input={(e) => setNum(i, "intensity", e.currentTarget.value)} />
      <input class="num" type="text" placeholder="mm:ss" value={s.timePointS != null ? mmss(s.timePointS) : ""} on:change={(e) => setTime(i, e.currentTarget.value)} />
      <input class="num" type="number" value={s.heartRate ?? ""} on:input={(e) => setNum(i, "heartRate", e.currentTarget.value)} />
      <input class="num" type="number" step="0.01" value={s.lactate ?? ""} on:input={(e) => setNum(i, "lactate", e.currentTarget.value)} />
      <input class="num" type="number" value={s.rpe ?? ""} on:input={(e) => setNum(i, "rpe", e.currentTarget.value)} />
      <button class="del" on:click={() => removeRow(i)} title="Remove row">×</button>
    </div>
  {/each}
  <button class="add" on:click={addRow}>+ Add row</button>
</div>

<style>
  .grid {
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    overflow: hidden;
    background: var(--surface);
  }
  .row {
    display: grid;
    grid-template-columns: 40px 1.4fr 1fr 1fr 1fr 0.8fr 36px;
    align-items: center;
    border-bottom: 1px solid var(--border);
  }
  .head {
    background: var(--surface-2);
    font-size: var(--fs-eyebrow);
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--text-faint);
  }
  .head span {
    padding: var(--space-2) var(--space-3);
  }
  .baseline {
    background: color-mix(in srgb, var(--accent) 5%, transparent);
  }
  .idx {
    padding: 0 var(--space-3);
    color: var(--text-faint);
  }
  input {
    height: 34px;
    border: none;
    border-left: 1px solid var(--border);
    background: transparent;
    padding: 0 var(--space-3);
    width: 100%;
    outline: none;
    font-variant-numeric: tabular-nums;
  }
  input:focus {
    background: color-mix(in srgb, var(--accent) 8%, transparent);
  }
  .del {
    border: none;
    background: transparent;
    color: var(--text-faint);
    height: 100%;
    font-size: 16px;
  }
  .del:hover {
    color: var(--danger);
  }
  .add {
    width: 100%;
    height: 34px;
    border: none;
    background: transparent;
    color: var(--accent);
    font-weight: 500;
  }
  .add:hover {
    background: var(--surface-2);
  }
</style>
