<script lang="ts">
  // Minimal, trusted markdown renderer for build-time content (the changelog)
  // and GitHub release notes. Handles h2/h3, paragraphs and bullet lists.
  export let src = "";
  export let dropTitle = true; // drop the leading H1 document title

  type Block =
    | { kind: "h2"; text: string }
    | { kind: "h3"; text: string }
    | { kind: "p"; text: string }
    | { kind: "ul"; items: string[] };

  function parse(md: string): Block[] {
    const blocks: Block[] = [];
    let list: string[] | null = null;
    const flush = () => {
      if (list) {
        blocks.push({ kind: "ul", items: list });
        list = null;
      }
    };
    for (const raw of (md ?? "").split("\n")) {
      const line = raw.trimEnd();
      if (line.startsWith("- ") || line.startsWith("* ")) {
        (list ??= []).push(inline(line.slice(2)));
      } else if (line.startsWith("### ")) {
        flush();
        blocks.push({ kind: "h3", text: line.slice(4) });
      } else if (line.startsWith("## ")) {
        flush();
        blocks.push({ kind: "h2", text: line.slice(3) });
      } else if (line.startsWith("# ")) {
        flush();
        if (!dropTitle) blocks.push({ kind: "h2", text: line.slice(2) });
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

  // strip bold/links/code to plain text
  function inline(s: string): string {
    return s
      .replace(/\*\*(.+?)\*\*/g, "$1")
      .replace(/_(.+?)_/g, "$1")
      .replace(/`(.+?)`/g, "$1")
      .replace(/\[(.+?)\]\((.+?)\)/g, "$1");
  }

  $: blocks = parse(src);
</script>

<div class="md">
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

<style>
  .md {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .version {
    margin-top: var(--space-4);
    padding-bottom: var(--space-1);
    border-bottom: 1px solid var(--border);
  }
  .version:first-child {
    margin-top: 0;
  }
  .cat {
    margin-top: var(--space-2);
    color: var(--signal);
  }
  .para {
    color: var(--text-muted);
    font-size: var(--fs-caption);
    line-height: 1.55;
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
    line-height: 1.5;
  }
</style>
