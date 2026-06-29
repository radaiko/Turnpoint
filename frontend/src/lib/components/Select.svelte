<script lang="ts">
  export let label = "";
  export let value: any = "";
  export let options: { value: string | number; label: string }[] = [];
</script>

<label class="field">
  {#if label}<span class="lbl">{label}</span>{/if}
  <div class="wrap">
    <select bind:value on:change>
      {#each options as o}
        <option value={o.value}>{o.label}</option>
      {/each}
    </select>
  </div>
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
    height: 32px;
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    background: var(--surface);
  }
  .wrap:focus-within {
    border-color: var(--accent);
  }
  select {
    width: 100%;
    height: 100%;
    border: none;
    background: transparent;
    outline: none;
    padding: 0 var(--space-3);
    color: var(--text);
  }
</style>
