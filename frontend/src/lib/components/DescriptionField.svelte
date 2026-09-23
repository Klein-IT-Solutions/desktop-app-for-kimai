<script lang="ts">
  import Icon from "./Icon.svelte";

  let {
    value = "",
    oninput,
    status = "",
    micAvailable = false,
  }: {
    value?: string;
    oninput: (text: string) => void;
    status?: "" | "saving" | "saved" | "error";
    micAvailable?: boolean;
  } = $props();
</script>

<div
  class="relative flex items-start gap-2 rounded-[10px] bg-field p-1.5 pl-3 shadow-[inset_0_0_0_0.5px_var(--hairline)] transition-colors focus-within:bg-field-hover"
>
  <textarea
    {value}
    oninput={(e) => oninput(e.currentTarget.value)}
    rows="2"
    placeholder="Woran arbeitest du?"
    class="min-h-[38px] flex-1 resize-none bg-transparent py-1 text-[13px] leading-snug text-fg outline-none placeholder:text-fg-subtle"
  ></textarea>
  <button
    class="grid size-8 shrink-0 place-items-center rounded-lg text-fg-muted transition-colors enabled:hover:bg-field-hover enabled:hover:text-accent disabled:opacity-40"
    aria-label="Spracheingabe"
    title={micAvailable ? "Gedrückt halten zum Sprechen" : "Spracheingabe folgt in Schritt 5"}
    disabled={!micAvailable}
  >
    <Icon name="mic" size={17} />
  </button>
  {#if status}
    <span
      class="absolute right-12 bottom-1.5 text-[10px] transition-opacity
        {status === 'error' ? 'text-danger' : status === 'saved' ? 'text-success' : 'text-fg-subtle'}"
    >
      {status === "saving" ? "Speichert …" : status === "saved" ? "In Kimai gespeichert" : "Nicht gespeichert"}
    </span>
  {/if}
</div>
