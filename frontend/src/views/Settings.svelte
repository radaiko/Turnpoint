<script lang="ts">
  import { onMount } from "svelte";
  import { App, type Template } from "$lib/api";
  import { toast } from "$lib/stores/toast";
  import { ui, setRegion } from "$lib/stores/ui";
  import { formatDate, type Region } from "$lib/format";

  function onRegion(e: Event) {
    setRegion((e.currentTarget as HTMLSelectElement).value as Region);
  }
  import Button from "$lib/components/Button.svelte";
  import Field from "$lib/components/Field.svelte";
  import Select from "$lib/components/Select.svelte";
  import Modal from "$lib/components/Modal.svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";

  let templates: Template[] = [];
  let loaded = false;

  let showModal = false;
  let draft: Template | null = null;

  onMount(load);

  async function load() {
    templates = await App.ListTemplates();
    loaded = true;
  }

  function blankDraft(): Template {
    return {
      id: 0,
      name: "",
      sport: "running",
      stepDurationS: 180,
      increment: 0.5,
      startIntensity: 6,
      endIntensity: undefined,
      mode: "continuous",
      restDurationS: undefined,
      visibleColumns: "",
      isPredefined: false,
    } as unknown as Template;
  }

  function newTemplate() {
    draft = blankDraft();
    showModal = true;
  }

  function editTemplate(t: Template) {
    draft = { ...t };
    showModal = true;
  }

  function cloneTemplate(t: Template) {
    draft = { ...t, id: 0, name: `${t.name} (copy)`, isPredefined: false };
    showModal = true;
  }

  async function saveTemplate() {
    if (!draft) return;
    if (!draft.name.trim()) {
      toast("Name is required", "warn");
      return;
    }
    await App.SaveTemplate({
      id: draft.id,
      name: draft.name,
      sport: draft.sport,
      stepDurationS: draft.stepDurationS,
      increment: draft.increment,
      startIntensity: draft.startIntensity,
      endIntensity: draft.endIntensity,
      mode: draft.mode,
      isPredefined: false,
      visibleColumns: "",
    } as unknown as Template);
    showModal = false;
    await load();
    toast("Template saved", "ok");
  }

  async function deleteTemplate(t: Template) {
    if (!confirm(`Delete template "${t.name}"? This cannot be undone.`)) return;
    await App.DeleteTemplate(t.id);
    await load();
    toast("Template deleted", "ok");
  }

  async function backup() {
    const path = await App.BackupDatabase();
    if (path) toast(`Backup saved to ${path}`, "ok");
  }

  async function restore() {
    const path = await App.RestoreDatabase();
    if (path) {
      toast("Database restored — reloading…", "ok");
      location.reload();
    }
  }

  function unit(sport: string): string {
    return sport === "cycling" ? "W" : "km/h";
  }

  function summary(t: Template): string {
    const end =
      t.endIntensity !== undefined && t.endIntensity !== null && t.endIntensity !== 0
        ? String(t.endIntensity)
        : "—";
    return `step ${t.stepDurationS}s · +${t.increment} · ${t.startIntensity}..${end} ${unit(t.sport)}`;
  }
</script>

<div class="page">
  <section class="block">
    <header class="block-head">
      <div>
        <h2>Templates</h2>
        <p class="sub">Reusable step-test protocols for new tests.</p>
      </div>
      <Button variant="primary" on:click={newTemplate}>+ New template</Button>
    </header>

    <div class="cards">
      {#each templates as t (t.id)}
        <div class="card">
          <div class="card-main">
            <div class="title-row">
              <span class="nm">{t.name}</span>
              <span class="tag">{t.sport}</span>
              {#if t.isPredefined}<span class="tag muted">predefined</span>{/if}
            </div>
            <span class="meta mono">{summary(t)}</span>
          </div>
          <div class="card-actions">
            {#if t.isPredefined}
              <Button variant="ghost" on:click={() => cloneTemplate(t)}>Clone</Button>
            {:else}
              <Button variant="ghost" on:click={() => editTemplate(t)}>Edit</Button>
              <Button variant="danger" on:click={() => deleteTemplate(t)}>Delete</Button>
            {/if}
          </div>
        </div>
      {/each}
      {#if loaded && !templates.length}
        <EmptyState title="No templates" hint="Create a template to speed up adding tests." />
      {/if}
    </div>
  </section>

  <section class="block">
    <header class="block-head">
      <div>
        <h2>Region</h2>
        <p class="sub">How dates and numbers are written.</p>
      </div>
    </header>

    <div class="card region-card">
      <Select
        label="Format"
        value={$ui.region}
        on:change={onRegion}
        options={[
          { value: "system", label: "Follow system" },
          { value: "eu", label: "Europe — 31.12.2025 · 1,5" },
          { value: "us", label: "United States — 12/31/2025 · 1.5" },
        ]}
      />
      <p class="caption">
        Today reads <span class="mono">{formatDate(new Date().toISOString().slice(0, 10), $ui.region)}</span>.
        Native date pickers follow your operating system.
      </p>
    </div>
  </section>

  <section class="block">
    <header class="block-head">
      <div>
        <h2>Data</h2>
        <p class="sub">Back up or restore your local database.</p>
      </div>
    </header>

    <div class="card data-card">
      <div class="data-actions">
        <Button on:click={backup}>Back up database…</Button>
        <Button on:click={restore}>Restore from backup…</Button>
      </div>
      <p class="caption">
        Back up writes a single portable <span class="mono">.db</span> file you can copy anywhere.
        All data stays local on this machine — nothing is ever sent to a server.
      </p>
    </div>
  </section>
</div>

<Modal title={draft?.id ? "Edit template" : "New template"} open={showModal} on:close={() => (showModal = false)}>
  {#if draft}
    <Field label="Name" bind:value={draft.name} />
    <Select
      label="Sport"
      bind:value={draft.sport}
      options={[
        { value: "running", label: "Running" },
        { value: "cycling", label: "Cycling" },
      ]}
    />
    <div class="grid2">
      <Field label="Step duration" type="number" step="1" suffix="s" bind:value={draft.stepDurationS} />
      <Field label="Increment" type="number" step="0.1" suffix={unit(draft.sport)} bind:value={draft.increment} />
      <Field label="Start intensity" type="number" step="0.1" suffix={unit(draft.sport)} bind:value={draft.startIntensity} />
      <Field label="End intensity" type="number" step="0.1" suffix={unit(draft.sport)} bind:value={draft.endIntensity} />
    </div>
    <Select
      label="Mode"
      bind:value={draft.mode}
      options={[
        { value: "continuous", label: "Continuous" },
        { value: "intermittent", label: "Intermittent" },
      ]}
    />
  {/if}
  <svelte:fragment slot="footer">
    <Button on:click={() => (showModal = false)}>Cancel</Button>
    <Button variant="primary" on:click={saveTemplate}>Save</Button>
  </svelte:fragment>
</Modal>

<style>
  .page {
    max-width: 900px;
    margin: 0 auto;
    padding: var(--space-6) var(--space-5) var(--space-8);
    display: flex;
    flex-direction: column;
    gap: var(--space-6);
  }
  .block {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }
  .block-head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--space-4);
  }
  h2 {
    font-size: var(--fs-h2);
    font-weight: 600;
    color: var(--text);
  }
  .sub {
    font-size: var(--fs-caption);
    color: var(--text-muted);
    margin-top: 2px;
  }
  .cards {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .card {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-4);
    padding: var(--space-3) var(--space-4);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    background: var(--surface);
  }
  .card-main {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    min-width: 0;
  }
  .title-row {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .nm {
    font-weight: 500;
    color: var(--text);
  }
  .meta {
    font-size: var(--fs-caption);
    color: var(--text-faint);
  }
  .card-actions {
    display: flex;
    gap: var(--space-2);
    flex-shrink: 0;
  }
  .tag {
    font-size: var(--fs-eyebrow);
    text-transform: uppercase;
    letter-spacing: 0.04em;
    padding: 2px 8px;
    border-radius: var(--radius-pill);
    background: var(--surface-2);
    color: var(--text-muted);
  }
  .tag.muted {
    background: transparent;
    border: 1px solid var(--border);
    color: var(--text-faint);
  }
  .data-card,
  .region-card {
    flex-direction: column;
    align-items: stretch;
    gap: var(--space-3);
  }
  .region-card {
    max-width: 440px;
  }
  .data-actions {
    display: flex;
    gap: var(--space-2);
  }
  .caption {
    font-size: var(--fs-caption);
    color: var(--text-muted);
    line-height: 1.5;
    max-width: 60ch;
  }
  .grid2 {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-3);
  }
</style>
