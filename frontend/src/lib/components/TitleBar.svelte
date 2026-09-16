<script lang="ts">
  import { Application, Window } from "@wailsio/runtime";
  import Icon from "./Icon.svelte";

  let {
    title,
    canGoBack = false,
    onBack,
    onSettings,
  }: { title: string; canGoBack?: boolean; onBack?: () => void; onSettings?: () => void } = $props();

  // Step 1: closing quits. Once the tray exists (step 4) this hides the window instead.
  const close = () => Application.Quit();
  const minimise = () => Window.Minimise();
</script>

<header class="drag relative flex h-10 shrink-0 items-center px-3.5">
  <div class="no-drag group flex items-center gap-2">
    <button
      class="grid size-3 place-items-center rounded-full bg-[#ff5f57] shadow-[inset_0_0_0_0.5px_rgba(0,0,0,0.12)]"
      aria-label="Schließen"
      onclick={close}
    >
      <svg viewBox="0 0 12 12" class="size-2 opacity-0 transition-opacity group-hover:opacity-100" stroke="#4d0000" stroke-width="1.4" stroke-linecap="round"><path d="M3.5 3.5l5 5m0-5l-5 5" /></svg>
    </button>
    <button
      class="grid size-3 place-items-center rounded-full bg-[#febc2e] shadow-[inset_0_0_0_0.5px_rgba(0,0,0,0.12)]"
      aria-label="Minimieren"
      onclick={minimise}
    >
      <svg viewBox="0 0 12 12" class="size-2 opacity-0 transition-opacity group-hover:opacity-100" stroke="#5a3a00" stroke-width="1.4" stroke-linecap="round"><path d="M3 6h6" /></svg>
    </button>
  </div>

  <div class="pointer-events-none absolute inset-x-0 text-center text-[13px] font-semibold tracking-[-0.01em] text-fg/85">
    {title}
  </div>

  <div class="no-drag ml-auto flex items-center">
    {#if canGoBack}
      <button
        class="rounded-md px-2 py-1 text-[13px] font-medium text-accent transition-opacity hover:opacity-75"
        onclick={onBack}
      >
        Fertig
      </button>
    {:else}
      <button
        class="grid size-7 place-items-center rounded-md text-fg-muted transition-colors hover:bg-field-hover hover:text-fg"
        aria-label="Einstellungen"
        onclick={onSettings}
      >
        <Icon name="gear" size={16} />
      </button>
    {/if}
  </div>
</header>
