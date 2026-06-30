<script lang="ts">
  import { onMount } from "svelte";
  import { get } from "svelte/store";
  import { ui, toggleTheme, setSection, type Section } from "$lib/stores/ui";
  import { checkForUpdate, update } from "$lib/stores/update";
  import { getPlatform, type Platform } from "$lib/window";
  import { VERSION } from "$lib/version";
  import WindowControls from "$lib/components/WindowControls.svelte";
  import UpdateBanner from "$lib/components/UpdateBanner.svelte";
  import Toasts from "$lib/components/Toasts.svelte";
  import Athletes from "./views/Athletes.svelte";
  import TestWorkspace from "./views/TestWorkspace.svelte";
  import Comparison from "./views/Comparison.svelte";
  import Settings from "./views/Settings.svelte";
  import WhatsNew from "./views/WhatsNew.svelte";
  import Updates from "./views/Updates.svelte";

  const navItems: { id: Section; label: string }[] = [
    { id: "athletes", label: "Athletes" },
    { id: "comparison", label: "Comparison" },
    { id: "settings", label: "Settings" },
  ];
  const metaItems: { id: Section; label: string }[] = [
    { id: "whatsnew", label: "What's New" },
    { id: "updates", label: "Updates" },
  ];

  let platform: Platform = "windows";
  $: isMac = platform === "darwin";
  onMount(async () => {
    platform = await getPlatform();
    // Best-effort update check (the app's only network call); silent on failure.
    if (get(ui).autoCheckUpdates) checkForUpdate();
  });
</script>

<div class="shell">
<div class="titlebar" class:mac={isMac} style="--wails-draggable: drag">
  {#if isMac}
    <WindowControls {platform} />
  {/if}
  <div class="brand">
    <span class="logo">▲</span>
    <span class="name">Turnpoint</span>
  </div>
  <div class="spacer" />
  <button class="theme" on:click={toggleTheme} title="Toggle theme" style="--wails-draggable: no-drag">
    {$ui.theme === "dark" ? "☾" : "☀"}
  </button>
  {#if !isMac}
    <span class="divider" />
    <WindowControls {platform} />
  {/if}
</div>

<UpdateBanner />

<div class="layout">
  <nav class="rail">
    <div class="nav-group">
      {#each navItems as item}
        <button
          class="nav-item"
          class:active={$ui.section === item.id}
          on:click={() => setSection(item.id)}
        >
          <span class="marker" />
          <span>{item.label}</span>
        </button>
      {/each}
    </div>
    <div class="nav-spacer" />
    <div class="nav-group">
      {#each metaItems as item}
        <button
          class="nav-item"
          class:active={$ui.section === item.id}
          on:click={() => setSection(item.id)}
        >
          <span class="marker" />
          <span>{item.label}</span>
          {#if item.id === "updates" && $update.info?.available}
            <span class="badge" title="Update available" />
          {/if}
        </button>
      {/each}
    </div>
    <div class="nav-foot">
      <span class="eyebrow">Local · Offline</span>
      <span class="ver mono">v{VERSION}</span>
    </div>
  </nav>

  <main class="stage">
    {#if $ui.section === "settings"}
      <Settings />
    {:else if $ui.section === "comparison"}
      <Comparison />
    {:else if $ui.section === "whatsnew"}
      <WhatsNew />
    {:else if $ui.section === "updates"}
      <Updates />
    {:else if $ui.activeTestId}
      <TestWorkspace />
    {:else}
      <Athletes />
    {/if}
  </main>
</div>
</div>

<Toasts />

<style>
  .titlebar {
    height: var(--titlebar-h);
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: 0 var(--space-3) 0 var(--space-4);
    border-bottom: 1px solid var(--border);
    background: var(--surface);
  }
  /* macOS traffic lights sit at the far left with a standard inset */
  .titlebar.mac {
    padding-left: var(--space-3);
  }
  .spacer {
    flex: 1;
  }
  .brand {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .logo {
    color: var(--signal);
    font-size: 13px;
  }
  .name {
    font-weight: 600;
    letter-spacing: -0.01em;
  }
  .theme {
    width: 28px;
    height: 28px;
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    background: var(--surface);
    color: var(--text-muted);
  }
  .theme:hover {
    background: var(--surface-2);
  }
  .divider {
    width: 1px;
    height: 18px;
    background: var(--border);
    margin: 0 var(--space-1);
  }
  .shell {
    height: 100vh;
    display: flex;
    flex-direction: column;
  }
  .layout {
    display: flex;
    flex: 1;
    min-height: 0;
  }
  .rail {
    width: var(--rail-w);
    flex-shrink: 0;
    border-right: 1px solid var(--border);
    background: var(--surface);
    display: flex;
    flex-direction: column;
    padding: var(--space-3);
  }
  .nav-group {
    display: flex;
    flex-direction: column;
    gap: 1px;
  }
  .nav-item {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    height: 32px;
    padding: 0 var(--space-2);
    border: none;
    background: transparent;
    color: var(--text-muted);
    text-align: left;
    font-weight: 500;
    letter-spacing: -0.006em;
  }
  .nav-item .marker {
    width: 2px;
    height: 13px;
    background: transparent;
    flex-shrink: 0;
  }
  .nav-item:hover {
    color: var(--text);
  }
  .nav-item.active {
    color: var(--text);
  }
  .nav-item.active .marker {
    background: var(--signal);
  }
  .badge {
    margin-left: auto;
    width: 7px;
    height: 7px;
    border-radius: var(--radius-pill);
    background: var(--signal);
    flex: none;
  }
  .nav-spacer {
    flex: 1;
  }
  .nav-foot {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-2);
    padding: var(--space-2);
    margin-top: var(--space-2);
  }
  .ver {
    font-size: var(--fs-eyebrow);
    color: var(--text-faint);
  }
  .stage {
    flex: 1;
    overflow: auto;
    background: var(--bg);
  }
</style>
