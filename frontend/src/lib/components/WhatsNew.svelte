<script lang="ts">
  import changelogRaw from "$root/CHANGELOG.md?raw";
  import Modal from "./Modal.svelte";

  export let open = false;

  type Block =
    | { kind: "h2"; text: string }
    | { kind: "h3"; text: string }
    | { kind: "p"; text: string }
    | { kind: "ul"; items: string[] };

  // Minimal markdown parse tailored to the changelog format (h1 intro dropped).
  function parse(md: string): Block[] {
    const blocks: Block[] = [];
    let list: string[] | null = null;
    const flush = () => {
      if (list) {
        blocks.push({ kind: "ul", items: list });
        list = null;
      }
    };
    for (const raw of md.split("\n")) {
      const line = raw.trimEnd();
      if (line.startsWith("- ")) {
        (list ??= []).push(inline(line.slice(2)));
      } else if (line.startsWith("### ")) {
        flush();
        blocks.push({ kind: "h3", text: line.slice(4) });
      } else if (line.startsWith("## ")) {
        flush();
        blocks.push({ kind: "h2", text: line.slice(3) });
      } else if (line.startsWith("# ")) {
        flush(); // drop the document title
      } else if (line.trim() === "") {
        flush();
      } else {
        flush();
        blocks.push({ kind: "p", text: inline(line) });
      }
    }
    flush();
    return blocks;
  }

  // strip bold/links/code to plain text (trusted, build-time content)
  function inline(s: string): string {
    return s
      .replace(/\*\*(.+?)\*\*/g, "$1")
      .replace(/`(.+?)`/g, "$1")
      .replace(/\[(.+?)\]\((.+?)\)/g, "$1");
  }

  $: blocks = parse(changelogRaw);
</script>

<Modal title="What's New" {open} on:close>
  <div class="log">
    {#each blocks as b}
      {#if b.kind === "h2"}
        <h3 class="version">{b.text}</h3>
      {:else if b.kind === "h3"}
        <p class="eyebrow cat">{b.text}</p>
      {:else if b.kind === "p"}
        <p class="para">{b.text}</p>
      {:else if b.kind === "ul"}
        <ul>
          {#each b.items as item}
            <li>{item}</li>
          {/each}
        </ul>
      {/if}
    {/each}
  </div>
</Modal>

<style>
  .log {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    max-height: 60vh;
    overflow: auto;
    padding-right: var(--space-2);
  }
  .version {
    margin-top: var(--space-3);
    padding-bottom: var(--space-1);
    border-bottom: 1px solid var(--border);
  }
  .version:first-child {
    margin-top: 0;
  }
  .cat {
    margin-top: var(--space-2);
    color: var(--accent);
  }
  .para {
    color: var(--text-muted);
    font-size: var(--fs-caption);
  }
  ul {
    margin: 0;
    padding-left: var(--space-4);
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }
  li {
    font-size: var(--fs-caption);
    color: var(--text);
  }
</style>
