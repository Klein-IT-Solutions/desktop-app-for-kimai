<script lang="ts">
  import { onDestroy } from "svelte";
  import SelectRow from "../components/SelectRow.svelte";
  import TimerButton from "../components/TimerButton.svelte";
  import DescriptionField from "../components/DescriptionField.svelte";

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

  function toggle() {
    startedAt = running ? null : Date.now();
    now = Date.now();
  }
</script>

<div class="flex h-full flex-col px-4 pb-3">
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
    class="mt-2.5 flex items-center justify-center gap-1.5 text-[11px] text-fg-subtle transition-colors hover:text-fg-muted"
    onclick={onOpenSettings}
  >
    <span class="size-1.5 rounded-full bg-warning"></span>
    Demo-Modus · Kimai verbinden
  </button>
</div>
