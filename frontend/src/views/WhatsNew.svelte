<script lang="ts">
  import changelogRaw from "$root/CHANGELOG.md?raw";
  import { VERSION } from "$lib/version";
  import Markdown from "$lib/components/Markdown.svelte";
</script>

<div class="page">
  <section class="block">
    <header class="head">
      <div>
        <p class="eyebrow">Changelog</p>
        <h1>What's New</h1>
      </div>
      <span class="ver mono">v{VERSION}</span>
    </header>
    <div class="card">
      <Markdown src={changelogRaw} />
    </div>
  </section>
</div>

<style>
  .page {
    max-width: 760px;
    margin: 0 auto;
    padding: var(--space-6) var(--space-6) var(--space-8);
  }
  .head {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: var(--space-4);
    margin-bottom: var(--space-5);
  }
  .ver {
    color: var(--text-faint);
    font-size: var(--fs-caption);
  }
  .card {
    padding: var(--space-5) var(--space-6);
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
  }
</style>
