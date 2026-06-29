<script lang="ts">
  export let label = "";
  // generic form wrapper: value may be string | number | undefined depending on the bound field
  export let value: any = "";
  export let type: "text" | "number" | "date" = "text";
  export let placeholder = "";
  export let suffix = "";
  export let step: string | undefined = undefined;
</script>

<label class="field">
  {#if label}<span class="lbl">{label}</span>{/if}
  <span class="wrap">
    {#if type === "number"}
      <input type="number" {step} {placeholder} bind:value on:input on:change class="num" />
    {:else if type === "date"}
      <input type="date" {placeholder} bind:value on:input on:change />
    {:else}
      <input type="text" {placeholder} bind:value on:input on:change />
    {/if}
    {#if suffix}<span class="suffix mono">{suffix}</span>{/if}
  </span>
</label>

<style>
  .field {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }
  .lbl {
    font-size: var(--fs-label);
    font-weight: 500;
    color: var(--text-muted);
  }
  .wrap {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    height: 32px;
    padding: 0 var(--space-3);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    background: var(--surface);
    transition: border-color 150ms ease;
  }
  .wrap:focus-within {
    border-color: var(--accent);
  }
  input {
    flex: 1;
    border: none;
    background: transparent;
    outline: none;
    width: 100%;
    min-width: 0;
  }
  .suffix {
    color: var(--text-faint);
    font-size: var(--fs-caption);
  }
</style>
