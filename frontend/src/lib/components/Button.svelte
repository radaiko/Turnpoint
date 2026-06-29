<script lang="ts">
  export let variant: "primary" | "ghost" | "danger" | "subtle" = "subtle";
  export let type: "button" | "submit" = "button";
  export let disabled = false;
</script>

<button {type} {disabled} class={variant} on:click>
  <slot />
</button>

<style>
  button {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    height: 32px;
    padding: 0 var(--space-3);
    border-radius: var(--radius-md);
    border: 1px solid var(--border);
    background: var(--surface);
    color: var(--text);
    font-weight: 500;
    transition: background 150ms ease, border-color 150ms ease;
  }
  button:hover:not(:disabled) {
    background: var(--surface-2);
  }
  button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .primary {
    background: var(--accent);
    border-color: var(--accent);
    color: var(--accent-contrast);
  }
  .primary:hover:not(:disabled) {
    background: var(--accent-hover);
    border-color: var(--accent-hover);
  }
  .ghost {
    background: transparent;
    border-color: transparent;
  }
  .ghost:hover:not(:disabled) {
    background: var(--surface-2);
  }
  .danger {
    color: var(--danger);
    border-color: var(--border);
  }
  .danger:hover:not(:disabled) {
    background: color-mix(in srgb, var(--danger) 10%, transparent);
    border-color: var(--danger);
  }
</style>
