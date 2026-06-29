<script lang="ts">
  import { onMount } from "svelte";
  import { get } from "svelte/store";
  import { ui, setStage, type Stage } from "$lib/stores/ui";
  import { analysis, ensureAnalysis } from "$lib/stores/analysis";
  import { loadConfig } from "$lib/stores/config";
  import { App } from "$lib/api";
  import Button from "$lib/components/Button.svelte";

  // Run the analysis once when the workspace opens so Analysis/Zones/Report all
  // have data regardless of which stage the user visits first, and load the
  // per-test analysis configuration.
  onMount(async () => {
    const id = get(ui).activeTestId;
    if (!id) return;
    const test = await App.GetTest(id);
    await loadConfig(id, test.sport);
    await ensureAnalysis(id);
  });
  import TestEntry from "./TestEntry.svelte";
  import Analysis from "./Analysis.svelte";
  import Zones from "./Zones.svelte";
  import Report from "./Report.svelte";

  const stages: { id: Stage; label: string }[] = [
    { id: "entry", label: "Entry" },
    { id: "analysis", label: "Analysis" },
    { id: "zones", label: "Zones" },
    { id: "report", label: "Report" },
  ];

  function back() {
    ui.update((s) => ({ ...s, activeTestId: null }));
    analysis.set(null);
  }
</script>

<div class="workspace">
  <header class="bar">
    <Button variant="ghost" on:click={back}>← Athletes</Button>
    <div class="tabs">
      {#each stages as s}
        <button class="tab" class:active={$ui.stage === s.id} on:click={() => setStage(s.id)}>
          {s.label}
        </button>
      {/each}
    </div>
    <div class="spacer" />
  </header>

  <div class="content">
    {#if $ui.stage === "entry"}
      <TestEntry />
    {:else if $ui.stage === "analysis"}
      <Analysis />
    {:else if $ui.stage === "zones"}
      <Zones />
    {:else if $ui.stage === "report"}
      <Report />
    {/if}
  </div>
</div>

<style>
  .workspace {
    display: flex;
    flex-direction: column;
    height: 100%;
  }
  .bar {
    display: flex;
    align-items: center;
    gap: var(--space-4);
    padding: var(--space-2) var(--space-4);
    border-bottom: 1px solid var(--border);
    background: var(--surface);
  }
  .tabs {
    display: flex;
    gap: 2px;
    background: var(--inset);
    padding: 3px;
    border-radius: var(--radius-md);
  }
  .tab {
    height: 28px;
    padding: 0 var(--space-4);
    border: none;
    border-radius: var(--radius-sm);
    background: transparent;
    color: var(--text-muted);
    font-weight: 500;
  }
  .tab.active {
    background: var(--surface);
    color: var(--text);
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.08);
  }
  .spacer {
    flex: 1;
  }
  .content {
    flex: 1;
    overflow: auto;
  }
</style>
