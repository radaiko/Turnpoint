<script lang="ts">
  export let title = "Nothing here yet";
  export let hint = "";
</script>

<div class="empty">
  <div class="mark">⊘</div>
  <p class="title">{title}</p>
  {#if hint}<p class="hint">{hint}</p>{/if}
  <slot />
</div>

<style>
  .empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--space-3);
    padding: var(--space-8) var(--space-4);
    color: var(--text-muted);
    text-align: center;
  }
  .mark {
    font-size: 32px;
    color: var(--text-faint);
  }
  .title {
    font-size: var(--fs-h3);
    font-weight: 600;
    color: var(--text);
  }
  .hint {
    font-size: var(--fs-caption);
    max-width: 40ch;
  }
</style>
