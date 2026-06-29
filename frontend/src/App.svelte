<script lang="ts">
  import { onMount } from "svelte";
  import { ui, toggleTheme, setSection } from "$lib/stores/ui";
  import { getPlatform, type Platform } from "$lib/window";
  import WindowControls from "$lib/components/WindowControls.svelte";
  import Toasts from "$lib/components/Toasts.svelte";
  import Athletes from "./views/Athletes.svelte";
  import TestWorkspace from "./views/TestWorkspace.svelte";
  import Comparison from "./views/Comparison.svelte";

  const navItems: { id: "athletes" | "comparison"; label: string; icon: string }[] = [
    { id: "athletes", label: "Athletes", icon: "◍" },
    { id: "comparison", label: "Comparison", icon: "≋" },
  ];

  let platform: Platform = "windows";
  $: isMac = platform === "darwin";
  onMount(async () => {
    platform = await getPlatform();
  });
</script>

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

<div class="layout">
  <nav class="rail">
    <div class="nav-group">
      {#each navItems as item}
        <button
          class="nav-item"
          class:active={$ui.section === item.id}
          on:click={() => setSection(item.id)}
        >
          <span class="icon">{item.icon}</span>
          <span>{item.label}</span>
        </button>
      {/each}
    </div>
    <div class="nav-spacer" />
    <div class="nav-foot eyebrow">Local · Offline</div>
  </nav>

  <main class="stage">
    {#if $ui.section === "comparison"}
      <Comparison />
    {:else if $ui.activeTestId}
      <TestWorkspace />
    {:else}
      <Athletes />
    {/if}
  </main>
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
    color: var(--accent);
    font-size: 14px;
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
  .layout {
    display: flex;
    height: calc(100vh - var(--titlebar-h));
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
    gap: 2px;
  }
  .nav-item {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    height: 36px;
    padding: 0 var(--space-3);
    border: none;
    border-radius: var(--radius-md);
    background: transparent;
    color: var(--text-muted);
    text-align: left;
    font-weight: 500;
  }
  .nav-item:hover {
    background: var(--surface-2);
    color: var(--text);
  }
  .nav-item.active {
    background: color-mix(in srgb, var(--accent) 12%, transparent);
    color: var(--accent);
  }
  .icon {
    width: 18px;
    text-align: center;
  }
  .nav-spacer {
    flex: 1;
  }
  .nav-foot {
    padding: var(--space-2);
  }
  .stage {
    flex: 1;
    overflow: auto;
    background: var(--bg);
  }
</style>
