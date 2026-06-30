<script lang="ts">
  import { onMount } from "svelte";
  import { App } from "$lib/api";
  import { VERSION } from "$lib/version";
  import { ui, setAutoCheckUpdates } from "$lib/stores/ui";
  import { update, checkForUpdate, setInstalling } from "$lib/stores/update";
  import { toast } from "$lib/stores/toast";
  import Button from "$lib/components/Button.svelte";
  import Markdown from "$lib/components/Markdown.svelte";

  // Check on first open if we don't already have a result this session.
  onMount(() => {
    if (!$update.info && !$update.checking) checkForUpdate();
  });

  async function check() {
    const info = await checkForUpdate();
    if (!info || info.error) {
      toast(info?.error || "Could not check — are you online?", "warn");
      return;
    }
    if (!info.available) toast("You're on the latest version", "ok");
  }

  async function updateNow() {
    const info = $update.info;
    if (!info) return;
    setInstalling(true);
    try {
      await App.DownloadAndRunUpdate(info.downloadUrl || info.releaseUrl);
      toast("Starting the update…", "ok");
    } catch (e: any) {
      toast(typeof e === "string" ? e : e?.message || "Update could not start", "danger");
      setInstalling(false);
    }
  }

  function onAutoToggle(e: Event) {
    setAutoCheckUpdates((e.currentTarget as HTMLInputElement).checked);
  }

  $: info = $update.info;
</script>

<div class="page">
  <section class="block">
    <header class="head">
      <p class="eyebrow">Software</p>
      <h1>Updates</h1>
    </header>

    <div class="card status">
      <div class="line">
        <span class="mark">▲</span>
        <span class="prod">Turnpoint</span>
        <span class="ver mono">v{VERSION}</span>
        <span
          class="state"
          class:ok={info && !info.available && !info.error}
          class:up={info?.available}
        >
          {#if $update.checking}
            Checking…
          {:else if info?.error}
            Couldn't check
          {:else if info?.available}
            Update available
          {:else if info}
            Up to date
          {:else}
            —
          {/if}
        </span>
      </div>
      <div class="actions">
        <Button on:click={check} disabled={$update.checking}>
          {$update.checking ? "Checking…" : "Check for updates"}
        </Button>
      </div>
      <label class="auto">
        <input type="checkbox" checked={$ui.autoCheckUpdates} on:change={onAutoToggle} />
        <span>Check automatically when the app launches</span>
      </label>
      <p class="caption">
        Checking for updates is the only time Turnpoint goes online — your data never leaves this
        machine.
      </p>
    </div>

    {#if info?.available}
      <div class="card release">
        <div class="rel-head">
          <div>
            <p class="eyebrow">New release</p>
            <h2>Version {info.latestVersion}</h2>
          </div>
          <Button variant="primary" on:click={updateNow} disabled={$update.installing}>
            {$update.installing ? "Updating…" : "Update now"}
          </Button>
        </div>
        {#if info.notes}
          <div class="notes"><Markdown src={info.notes} dropTitle={false} /></div>
        {/if}
        <button class="link" on:click={() => info && App.OpenReleasePage(info.releaseUrl)}>
          View release on GitHub
        </button>
      </div>
    {/if}
  </section>
</div>

<style>
  .page {
    max-width: 640px;
    margin: 0 auto;
    padding: var(--space-6) var(--space-6) var(--space-8);
  }
  .head {
    margin-bottom: var(--space-5);
  }
  .card {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    padding: var(--space-5);
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
  }
  .card + .card {
    margin-top: var(--space-4);
  }
  .line {
    display: flex;
    align-items: baseline;
    gap: var(--space-2);
  }
  .mark {
    color: var(--signal);
    font-size: 12px;
  }
  .prod {
    font-weight: 600;
  }
  .ver {
    color: var(--text-muted);
    font-size: var(--fs-caption);
  }
  .state {
    margin-left: auto;
    font-family: var(--font-mono);
    font-size: var(--fs-micro);
    letter-spacing: var(--track-micro);
    text-transform: uppercase;
    color: var(--text-faint);
  }
  .state.ok {
    color: var(--ok);
  }
  .state.up {
    color: var(--signal);
  }
  .actions {
    display: flex;
    gap: var(--space-2);
  }
  .auto {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    font-size: var(--fs-caption);
    color: var(--text-muted);
    cursor: pointer;
  }
  .auto input {
    accent-color: var(--accent);
  }
  .caption {
    font-size: var(--fs-caption);
    color: var(--text-faint);
    line-height: 1.5;
  }
  .release {
    border-color: color-mix(in srgb, var(--signal) 35%, var(--border));
  }
  .rel-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-4);
  }
  .notes {
    padding-top: var(--space-2);
    border-top: 1px solid var(--border);
  }
  .link {
    align-self: flex-start;
    background: none;
    border: none;
    padding: 0;
    color: var(--text-muted);
    text-decoration: underline;
    text-underline-offset: 2px;
    cursor: pointer;
    font-size: var(--fs-caption);
  }
  .link:hover {
    color: var(--text);
  }
</style>
