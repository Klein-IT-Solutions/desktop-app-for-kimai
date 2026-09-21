<script lang="ts">
  import { onDestroy } from "svelte";
  import SelectRow from "../components/SelectRow.svelte";
  import TimerButton from "../components/TimerButton.svelte";
  import DescriptionField from "../components/DescriptionField.svelte";
  import { connection } from "../connection.svelte";

  let { onOpenSettings }: { onOpenSettings: () => void } = $props();

  // Step 1: a purely local demo timer to try out the look & feel.
  // Step 3 replaces this with the Kimai-backed tracker state.
  let startedAt = $state<number | null>(null);
  let now = $state(Date.now());
  let description = $state("");

  const running = $derived(startedAt !== null);
  const elapsed = $derived(startedAt ? (now - startedAt) / 1000 : 0);

  const tick = setInterval(() => (now = Date.now()), 250);
  onDestroy(() => clearInterval(tick));

  const status = $derived.by(() => {
    const r = connection.result;
    if (connection.checking && !r) return { tone: "bg-fg-subtle", text: "Verbindung wird geprüft …" };
    if (!r || r.status === "not_configured") return { tone: "bg-warning", text: "Kimai verbinden" };
    if (r.ok) return { tone: "bg-success", text: `${r.userName} · ${host(r.baseUrl)}` };
    if (r.status === "offline") return { tone: "bg-danger", text: "Offline – Kimai nicht erreichbar" };
    if (r.status === "unauthorized") return { tone: "bg-danger", text: "API-Token ungültig" };
    if (r.status === "keyring") return { tone: "bg-danger", text: "Schlüsselbund nicht verfügbar" };
    return { tone: "bg-danger", text: "Verbindungsproblem" };
  });

  function host(url: string) {
    try {
      return new URL(url).host;
    } catch {
      return url;
    }
  }

  function toggle() {
    startedAt = running ? null : Date.now();
    now = Date.now();
  }
</script>

<div class="flex h-full flex-col px-4 pt-1 pb-3">
  <section
    class="overflow-hidden rounded-[10px] bg-field shadow-[inset_0_0_0_0.5px_var(--hairline)] [&>*+*]:border-t [&>*+*]:border-hairline"
  >
    <SelectRow icon="building" label="Kunde" disabled />
    <SelectRow icon="folder" label="Projekt" disabled />
    <SelectRow icon="tag" label="Tätigkeit" disabled />
  </section>

  <div class="flex flex-1 items-center justify-center py-2">
    <TimerButton {running} {elapsed} ontoggle={toggle} />
  </div>

  <DescriptionField bind:value={description} />

  <button
    class="mt-2.5 flex min-w-0 items-center justify-center gap-1.5 text-[11px] text-fg-subtle transition-colors hover:text-fg-muted"
    onclick={onOpenSettings}
  >
    <span class="size-1.5 shrink-0 rounded-full transition-colors {status.tone}"></span>
    <span class="truncate">{status.text}</span>
  </button>
</div>
