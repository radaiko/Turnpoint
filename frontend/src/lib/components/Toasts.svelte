<script lang="ts">
  import { toasts, dismiss } from "$lib/stores/toast";
</script>

<div class="stack">
  {#each $toasts as t (t.id)}
    <button class="toast {t.kind}" on:click={() => dismiss(t.id)}>
      {t.message}
    </button>
  {/each}
</div>

<style>
  .stack {
    position: fixed;
    bottom: var(--space-4);
    right: var(--space-4);
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    z-index: 200;
  }
  .toast {
    padding: var(--space-2) var(--space-4);
    border-radius: var(--radius-md);
    background: var(--surface);
    border: 1px solid var(--border);
    box-shadow: var(--shadow-pop);
    font-size: var(--fs-caption);
    color: var(--text);
    cursor: pointer;
    border-left-width: 3px;
    text-align: left;
  }
  .ok {
    border-left-color: var(--ok);
  }
  .warn {
    border-left-color: var(--warn);
  }
  .danger {
    border-left-color: var(--danger);
  }
  .info {
    border-left-color: var(--accent);
  }
</style>
