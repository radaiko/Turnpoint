<script lang="ts">
  import { createEventDispatcher } from "svelte";
  export let title = "";
  export let open = false;
  const dispatch = createEventDispatcher();
  function close() {
    dispatch("close");
  }
</script>

{#if open}
  <div class="backdrop" on:click|self={close} on:keydown={(e) => e.key === "Escape" && close()} role="presentation">
    <div class="modal" role="dialog" aria-modal="true">
      <header>
        <h3>{title}</h3>
      </header>
      <div class="body"><slot /></div>
      <footer><slot name="footer" /></footer>
    </div>
  </div>
{/if}

<style>
  .backdrop {
    position: fixed;
    inset: 0;
    background: rgba(8, 10, 14, 0.5);
    display: grid;
    place-items: center;
    z-index: 100;
  }
  .modal {
    width: min(560px, 92vw);
    max-height: 86vh;
    overflow: auto;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-modal);
  }
  header {
    padding: var(--space-4);
    border-bottom: 1px solid var(--border);
  }
  .body {
    padding: var(--space-4);
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }
  footer {
    padding: var(--space-4);
    border-top: 1px solid var(--border);
    display: flex;
    justify-content: flex-end;
    gap: var(--space-2);
  }
</style>
