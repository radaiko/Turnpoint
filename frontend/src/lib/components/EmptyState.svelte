<script lang="ts">
  export let title = "Nothing here yet";
  export let hint = "";
</script>

<div class="empty">
  <div class="rule" />
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
  .rule {
    width: 28px;
    height: 2px;
    background: var(--signal);
    margin-bottom: var(--space-1);
  }
  .title {
    font-size: var(--fs-body);
    font-weight: 600;
    color: var(--text);
    letter-spacing: -0.01em;
  }
  .hint {
    font-size: var(--fs-caption);
    max-width: 38ch;
    line-height: 1.55;
    color: var(--text-faint);
  }
</style>
