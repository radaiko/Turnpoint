<script lang="ts">
  import { App } from "$lib/api";
  import { update, dismissUpdate, setInstalling } from "$lib/stores/update";
  import { toast } from "$lib/stores/toast";

  $: info = $update.info;
  $: show = !!info?.available && !$update.dismissed;

  async function updateNow() {
    if (!info) return;
    setInstalling(true);
    try {
      // Windows: downloads + runs the installer and quits the app. Elsewhere:
      // opens the release page.
      await App.DownloadAndRunUpdate(info.downloadUrl || info.releaseUrl);
      toast("Starting the update…", "ok");
    } catch (e: any) {
      toast(typeof e === "string" ? e : e?.message || "Update could not start", "danger");
      setInstalling(false);
    }
  }

  function details() {
    if (info) App.OpenReleasePage(info.releaseUrl);
  }
</script>

{#if show && info}
  <div class="banner" role="status">
    <span class="pip" />
    <span class="msg">
      <strong>Update available</strong>
      <span class="ver mono">v{info.currentVersion} → v{info.latestVersion}</span>
    </span>
    <div class="actions">
      <button class="link" on:click={details}>Release notes</button>
      <button class="cta" on:click={updateNow} disabled={$update.installing}>
        {$update.installing ? "Updating…" : "Update now"}
      </button>
      <button class="x" title="Dismiss" aria-label="Dismiss" on:click={dismissUpdate}>✕</button>
    </div>
  </div>
{/if}

<style>
  .banner {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-2) var(--space-4);
    background: var(--signal-soft);
    border-bottom: 1px solid color-mix(in srgb, var(--signal) 35%, transparent);
    color: var(--text);
    font-size: var(--fs-caption);
  }
  .pip {
    width: 7px;
    height: 7px;
    border-radius: var(--radius-pill);
    background: var(--signal);
    flex: none;
  }
  .msg {
    display: flex;
    align-items: baseline;
    gap: var(--space-2);
  }
  .ver {
    color: var(--text-muted);
    font-size: var(--fs-micro);
  }
  .actions {
    margin-left: auto;
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .link {
    background: none;
    border: none;
    color: var(--text-muted);
    text-decoration: underline;
    text-underline-offset: 2px;
    cursor: pointer;
    font-size: var(--fs-caption);
  }
  .link:hover {
    color: var(--text);
  }
  .cta {
    border: none;
    border-radius: var(--radius-md);
    background: var(--signal);
    color: #fff;
    padding: var(--space-1) var(--space-3);
    font-weight: 600;
    font-size: var(--fs-caption);
    cursor: pointer;
  }
  .cta:hover {
    filter: brightness(1.06);
  }
  .cta:disabled {
    opacity: 0.6;
    cursor: default;
  }
  .x {
    background: none;
    border: none;
    color: var(--text-faint);
    cursor: pointer;
    padding: 0 var(--space-1);
    font-size: var(--fs-caption);
  }
  .x:hover {
    color: var(--text);
  }
</style>
