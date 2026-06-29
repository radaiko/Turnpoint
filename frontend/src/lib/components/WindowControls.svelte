<script lang="ts">
  import { minimise, toggleMaximise, close, type Platform } from "$lib/window";
  export let platform: Platform = "windows";
  $: isMac = platform === "darwin";
</script>

{#if isMac}
  <!-- macOS traffic lights: close · minimise · maximise (left, glyphs on hover) -->
  <div class="mac" style="--wails-draggable: no-drag">
    <button class="dot close" on:click={close} aria-label="Close"><span>×</span></button>
    <button class="dot min" on:click={minimise} aria-label="Minimise"><span>−</span></button>
    <button class="dot max" on:click={toggleMaximise} aria-label="Maximise"><span>+</span></button>
  </div>
{:else}
  <!-- Windows / Linux: minimise · maximise · close (right) -->
  <div class="win" style="--wails-draggable: no-drag">
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
{/if}

<style>
  /* ── macOS traffic lights ── */
  .mac {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .dot {
    width: 12px;
    height: 12px;
    border-radius: var(--radius-pill);
    border: none;
    padding: 0;
    display: grid;
    place-items: center;
    line-height: 1;
  }
  .dot span {
    font-size: 9px;
    font-weight: 700;
    color: rgba(0, 0, 0, 0.55);
    opacity: 0;
    transition: opacity 120ms ease;
  }
  .mac:hover .dot span {
    opacity: 1;
  }
  .dot.close {
    background: #ff5f57;
  }
  .dot.min {
    background: #febc2e;
  }
  .dot.max {
    background: #28c840;
  }

  /* ── Windows / Linux ── */
  .win {
    display: flex;
    align-items: center;
    gap: 2px;
  }
  .wc {
    display: grid;
    place-items: center;
    width: 30px;
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
</style>
