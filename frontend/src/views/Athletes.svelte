<script lang="ts">
  import { onMount } from "svelte";
  import { App, type AthleteSummary, type Athlete, type Test, type Template } from "$lib/api";
  import { openTest } from "$lib/stores/ui";
  import { toast } from "$lib/stores/toast";
  import { ageFromDOB } from "$lib/format";
  import Button from "$lib/components/Button.svelte";
  import Field from "$lib/components/Field.svelte";
  import Select from "$lib/components/Select.svelte";
  import Modal from "$lib/components/Modal.svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";

  let athletes: AthleteSummary[] = [];
  let search = "";
  let selectedId: number | null = null;
  let selected: Athlete | null = null;
  let tests: Test[] = [];
  let templates: Template[] = [];

  let editing: Athlete | null = null;
  let showAthleteModal = false;
  let showTestModal = false;
  let newTest = { templateId: 0, date: new Date().toISOString().slice(0, 10) };

  onMount(async () => {
    await load();
    templates = (await App.ListTemplates()) ?? [];
  });

  async function load() {
    athletes = (await App.ListAthletes(search)) ?? [];
  }

  async function selectAthlete(id: number) {
    selectedId = id;
    selected = await App.GetAthlete(id);
    tests = (await App.ListTests(id)) ?? [];
  }

  function newAthlete() {
    editing = { id: 0, name: "", dob: "", sex: "unspecified", bodyMassKg: undefined, primarySport: "running", notes: "" } as unknown as Athlete;
    showAthleteModal = true;
  }
  function editAthlete() {
    if (selected) {
      editing = { ...selected };
      showAthleteModal = true;
    }
  }

  async function saveAthlete() {
    if (!editing) return;
    if (!editing.name.trim()) {
      toast("Name is required", "warn");
      return;
    }
    // Coerce the optional numeric body mass: "" / null → omitted (Go nil), else a number.
    const bm = (editing as any).bodyMassKg;
    const payload: any = {
      ...editing,
      bodyMassKg: bm === "" || bm === null || bm === undefined || isNaN(Number(bm)) ? undefined : Number(bm),
    };
    const id = await App.SaveAthlete(payload);
    showAthleteModal = false;
    await load();
    await selectAthlete(id);
    toast("Athlete saved", "ok");
  }

  async function deleteAthlete() {
    if (!selected) return;
    if (!confirm(`Delete ${selected.name} and all their tests? This cannot be undone.`)) return;
    await App.DeleteAthlete(selected.id);
    selectedId = null;
    selected = null;
    tests = [];
    await load();
    toast("Athlete deleted", "ok");
  }

  function openNewTest() {
    if (!templates.length) return;
    newTest = { templateId: templates[0].id, date: new Date().toISOString().slice(0, 10) };
    showTestModal = true;
  }

  async function createTest() {
    if (!selected) return;
    const tpl = templates.find((t) => t.id === newTest.templateId) ?? templates[0];
    const id = await App.SaveTest({
      id: 0,
      athleteId: selected.id,
      testDate: newTest.date,
      sport: tpl.sport,
      stepDurationS: tpl.stepDurationS,
      increment: tpl.increment,
      startIntensity: tpl.startIntensity,
      mode: tpl.mode,
      templateId: tpl.id,
    } as unknown as Test);
    showTestModal = false;
    openTest(id);
  }
</script>

<div class="cols">
  <section class="list">
    <header>
      <h2>Athletes</h2>
      <Button variant="primary" on:click={newAthlete}>+ New</Button>
    </header>
    <input class="search" placeholder="Search…" bind:value={search} on:input={load} />
    <div class="rows">
      {#each athletes as a (a.id)}
        <button class="row" class:active={a.id === selectedId} on:click={() => selectAthlete(a.id)}>
          <span class="nm">{a.name}</span>
          <span class="meta mono">{a.primarySport ?? "—"} · {a.testCount} tests</span>
        </button>
      {/each}
      {#if !athletes.length}
        <EmptyState title="No athletes" hint="Create your first athlete to begin." />
      {/if}
    </div>
  </section>

  <section class="detail">
    {#if selected}
      <header>
        <div>
          <h2>{selected.name}</h2>
          <p class="sub mono">
            {selected.sex} · age {ageFromDOB(selected.dob)} ·
            {selected.bodyMassKg ? selected.bodyMassKg + " kg" : "no mass"}
          </p>
        </div>
        <div class="actions">
          <Button on:click={editAthlete}>Edit</Button>
          <Button variant="danger" on:click={deleteAthlete}>Delete</Button>
        </div>
      </header>

      <div class="tests-head">
        <h3>Tests</h3>
        <Button variant="primary" on:click={openNewTest}>+ New test</Button>
      </div>
      <div class="tests">
        {#each tests as t (t.id)}
          <button class="test-row" on:click={() => openTest(t.id)}>
            <span class="mono">{t.testDate}</span>
            <span class="tag">{t.sport}</span>
            <span class="meta">{t.startIntensity}{t.sport === "cycling" ? "W" : " km/h"} start · +{t.increment}</span>
          </button>
        {/each}
        {#if !tests.length}
          <EmptyState title="No tests yet" hint="Add a step test for this athlete." />
        {/if}
      </div>
    {:else}
      <EmptyState title="Select an athlete" hint="Pick someone on the left, or create a new athlete." />
    {/if}
  </section>
</div>

<Modal title={editing?.id ? "Edit athlete" : "New athlete"} open={showAthleteModal} on:close={() => (showAthleteModal = false)}>
  {#if editing}
    <Field label="Name" bind:value={editing.name} />
    <Field label="Date of birth" type="date" bind:value={editing.dob} />
    <Select label="Sex" bind:value={editing.sex} options={[
      { value: "unspecified", label: "Unspecified" },
      { value: "male", label: "Male" },
      { value: "female", label: "Female" },
    ]} />
    <Field label="Body mass" type="number" step="0.1" suffix="kg" bind:value={editing.bodyMassKg} />
    <Select label="Primary sport" bind:value={editing.primarySport} options={[
      { value: "running", label: "Running" },
      { value: "cycling", label: "Cycling" },
    ]} />
    <Field label="Notes" bind:value={editing.notes} />
  {/if}
  <svelte:fragment slot="footer">
    <Button on:click={() => (showAthleteModal = false)}>Cancel</Button>
    <Button variant="primary" on:click={saveAthlete}>Save</Button>
  </svelte:fragment>
</Modal>

<Modal title="New test" open={showTestModal} on:close={() => (showTestModal = false)}>
  <Select label="Template" bind:value={newTest.templateId}
    options={templates.map((t) => ({ value: t.id, label: t.name }))} />
  <Field label="Date" type="date" bind:value={newTest.date} />
  <svelte:fragment slot="footer">
    <Button on:click={() => (showTestModal = false)}>Cancel</Button>
    <Button variant="primary" on:click={createTest}>Create</Button>
  </svelte:fragment>
</Modal>

<style>
  .cols {
    display: grid;
    grid-template-columns: 320px 1fr;
    height: 100%;
  }
  .list {
    border-right: 1px solid var(--border);
    display: flex;
    flex-direction: column;
    background: var(--surface);
  }
  .list header,
  .detail header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-4);
  }
  .search {
    margin: 0 var(--space-4) var(--space-3);
    height: 32px;
    padding: 0 var(--space-3);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    background: var(--bg);
    outline: none;
  }
  .rows {
    flex: 1;
    overflow: auto;
    padding: 0 var(--space-2) var(--space-2);
  }
  .row {
    width: 100%;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 2px;
    padding: var(--space-2) var(--space-3);
    border: none;
    border-radius: var(--radius-md);
    background: transparent;
    text-align: left;
  }
  .row:hover {
    background: var(--surface-2);
  }
  .row.active {
    background: color-mix(in srgb, var(--accent) 12%, transparent);
  }
  .nm {
    font-weight: 500;
    color: var(--text);
  }
  .meta {
    font-size: var(--fs-caption);
    color: var(--text-faint);
  }
  .detail {
    overflow: auto;
  }
  .sub {
    font-size: var(--fs-caption);
    color: var(--text-muted);
    margin-top: 2px;
  }
  .actions {
    display: flex;
    gap: var(--space-2);
  }
  .tests-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 var(--space-4) var(--space-2);
  }
  .tests {
    padding: 0 var(--space-4) var(--space-4);
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }
  .test-row {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-3);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    background: var(--surface);
    text-align: left;
  }
  .test-row:hover {
    border-color: var(--accent);
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
</style>
