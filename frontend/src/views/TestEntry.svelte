<script lang="ts">
  import { onMount } from "svelte";
  import { get } from "svelte/store";
  import { App, type Test, type Step } from "$lib/api";
  import { ui, setStage } from "$lib/stores/ui";
  import { runAnalysis } from "$lib/stores/analysis";
  import { toast } from "$lib/stores/toast";
  import Button from "$lib/components/Button.svelte";
  import Field from "$lib/components/Field.svelte";
  import DataGrid from "$lib/components/DataGrid.svelte";

  let test: Test | null = null;
  let steps: Step[] = [];
  let unit = "km/h";
  let dirty = false;

  onMount(load);

  async function load() {
    const id = get(ui).activeTestId;
    if (!id) return;
    test = await App.GetTest(id);
    unit = test.sport === "cycling" ? "W" : "km/h";
    steps = await App.GetSteps(id);
    if (!steps.length) steps = seedFromProtocol(test);
  }

  // Pre-fill the grid from the protocol (baseline + each step) for a fresh test.
  function seedFromProtocol(t: Test): Step[] {
    const rows: Step[] = [{ id: 0, testId: t.id, stepOrder: 0, intensity: 0, isBaseline: true, excluded: false, aborted: false } as unknown as Step];
    let v = t.startIntensity;
    let order = 1;
    const end = t.sport === "cycling" ? 440 : 22;
    while (v <= end && order < 20) {
      rows.push({ id: 0, testId: t.id, stepOrder: order, intensity: v, isBaseline: false, excluded: false, aborted: false } as unknown as Step);
      v += t.increment;
      order++;
    }
    return rows;
  }

  async function save(thenAnalyze = false) {
    if (!test) return;
    await App.SaveSteps(test.id, steps);
    dirty = false;
    toast("Saved", "ok");
    if (thenAnalyze) {
      await runAnalysis(test.id);
      setStage("analysis");
    }
  }
</script>

<div class="entry">
  {#if test}
    <header>
      <div>
        <span class="eyebrow">{test.sport} · {test.testDate}</span>
        <h2>Step data</h2>
      </div>
      <div class="protocol mono">
        {test.stepDurationS}s steps · +{test.increment}{unit} · start {test.startIntensity}{unit}
      </div>
    </header>

    <DataGrid bind:steps {unit} on:change={() => (dirty = true)} />

    <Field label="Pre-test note" bind:value={test.pretestNote} on:input={() => (dirty = true)} />
    <Field label="Remarks" bind:value={test.remarks} on:input={() => (dirty = true)} />

    <footer>
      <Button on:click={() => save(false)} disabled={!dirty && steps.length === 0}>Save</Button>
      <Button variant="primary" on:click={() => save(true)}>Save & Analyze →</Button>
    </footer>
  {/if}
</div>

<style>
  .entry {
    padding: var(--space-5);
    max-width: 900px;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }
  header {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
  }
  .protocol {
    font-size: var(--fs-caption);
    color: var(--text-muted);
  }
  footer {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-2);
  }
</style>
