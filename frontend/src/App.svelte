<script lang="ts">
  import { ui, toggleTheme, setSection } from "$lib/stores/ui";
  import { minimise, toggleMaximise, close } from "$lib/window";
  import Toasts from "$lib/components/Toasts.svelte";
  import Athletes from "./views/Athletes.svelte";
  import TestWorkspace from "./views/TestWorkspace.svelte";
  import Comparison from "./views/Comparison.svelte";

  const navItems: { id: "athletes" | "comparison"; label: string; icon: string }[] = [
    { id: "athletes", label: "Athletes", icon: "◍" },
    { id: "comparison", label: "Comparison", icon: "≋" },
  ];
</script>

<div class="titlebar" style="--wails-draggable: drag">
  <div class="brand">
    <span class="logo">▲</span>
    <span class="name">Turnpoint</span>
  </div>
  <div class="win-controls" style="--wails-draggable: no-drag">
    <button class="theme" on:click={toggleTheme} title="Toggle theme">
      {$ui.theme === "dark" ? "☾" : "☀"}
    </button>
    <span class="divider" />
    <button class="wc" on:click={minimise} title="Minimise" aria-label="Minimise">
      <svg width="11" height="11" viewBox="0 0 11 11"><line x1="1.5" y1="5.5" x2="9.5" y2="5.5" stroke="currentColor" stroke-width="1.2" /></svg>
    </button>
    <button class="wc" on:click={toggleMaximise} title="Maximise" aria-label="Maximise">
      <svg width="11" height="11" viewBox="0 0 11 11"><rect x="1.6" y="1.6" width="7.8" height="7.8" rx="1.2" fill="none" stroke="currentColor" stroke-width="1.2" /></svg>
    </button>
    <button class="wc close" on:click={close} title="Close" aria-label="Close">
      <svg width="11" height="11" viewBox="0 0 11 11"><line x1="2" y1="2" x2="9" y2="9" stroke="currentColor" stroke-width="1.2" /><line x1="9" y1="2" x2="2" y2="9" stroke="currentColor" stroke-width="1.2" /></svg>
    </button>
  </div>
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
    justify-content: space-between;
    padding: 0 var(--space-3) 0 var(--space-4);
    border-bottom: 1px solid var(--border);
    background: var(--surface);
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
  .win-controls {
    display: flex;
    align-items: center;
    gap: var(--space-1);
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
  .wc {
    display: grid;
    place-items: center;
    width: 28px;
    height: 28px;
    border: none;
    border-radius: var(--radius-md);
    background: transparent;
    color: var(--text-muted);
  }
  .wc:hover {
    background: var(--surface-2);
    color: var(--text);
  }
  .wc.close:hover {
    background: var(--danger);
    color: #fff;
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
