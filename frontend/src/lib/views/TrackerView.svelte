<script lang="ts">
  import { onDestroy } from "svelte";
  import { slide } from "svelte/transition";
  import { TrackerService, type Option } from "../../../bindings/github.com/klein-it-solutions/kimai-desktop-app/internal/app";
  import SelectRow from "../components/SelectRow.svelte";
  import TimerButton from "../components/TimerButton.svelte";
  import DescriptionField from "../components/DescriptionField.svelte";
  import Picker from "../components/Picker.svelte";
  import { connection } from "../connection.svelte";
  import { voice } from "../voice.svelte";
  import { voiceEnabled } from "../voiceSettings.svelte";
  import { assistant, undoSummary } from "../assistant.svelte";
  import {
    canStart,
    selectActivity,
    selectCustomer,
    selectProject,
    setDescription,
    toggleTimer,
    tracker,
  } from "../tracker.svelte";

  let { onOpenSettings }: { onOpenSettings: () => void } = $props();

  type PickerKind = "customer" | "project" | "activity" | null;
  let picker = $state<PickerKind>(null);

  let now = $state(Date.now());
  const tick = setInterval(() => (now = Date.now()), 250);
  onDestroy(() => clearInterval(tick));

  const s = $derived(tracker.state);
  const connected = $derived(connection.result?.ok ?? false);
  const elapsed = $derived(s.running && s.begin ? Math.max(0, (now - Date.parse(s.begin)) / 1000) : 0);
  const locked = $derived(s.running || tracker.busy);

  // After a choice, open the next picker – or pick automatically when there is only one option.
  async function afterCustomer() {
    const items = (await TrackerService.Projects(s.customer.id, false).catch(() => null))?.items ?? [];
    if (items.length === 1) {
      selectProject(items[0]);
      return afterProject();
    }
    picker = "project";
  }

  async function afterProject() {
    const items = (await TrackerService.Activities(s.project.id, false).catch(() => null))?.items ?? [];
    if (items.length === 1) {
      selectActivity(items[0]);
      picker = null;
      return;
    }
    picker = s.activity.id ? null : "activity";
  }

  function onCustomer(o: Option) {
    const changed = o.id !== s.customer.id;
    selectCustomer(o);
    picker = null;
    if (changed || !s.project.id) afterCustomer();
  }

  function onProject(o: Option) {
    const changed = o.id !== s.project.id;
    selectProject(o);
    picker = null;
    if (changed || !s.activity.id) afterProject();
  }

  function onActivity(o: Option) {
    selectActivity(o);
    picker = null;
  }

  function onKey(e: KeyboardEvent) {
    // Ctrl/Cmd+Enter starts or stops the timer from anywhere in the window.
    if (e.key === "Enter" && (e.ctrlKey || e.metaKey) && !picker) {
      if (s.running || (canStart() && connected)) toggleTimer();
      e.preventDefault();
    }
  }

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

  const hint = $derived.by(() => {
    if (s.running) return "";
    if (!connected) return "";
    if (!s.customer.id) return "Kunde auswählen";
    if (!s.project.id) return "Projekt auswählen";
    if (!s.activity.id) return "Tätigkeit auswählen";
    return "";
  });
</script>

<svelte:window onkeydown={onKey} />

<div class="relative flex h-full flex-col px-4 pt-1 pb-3">
  <section
    class="overflow-hidden rounded-[10px] bg-field shadow-[inset_0_0_0_0.5px_var(--hairline)] transition-opacity [&>*+*]:border-t [&>*+*]:border-hairline
      {s.running ? 'opacity-80' : ''}"
  >
    <SelectRow icon="building" label="Kunde" value={s.customer.name} disabled={locked} onclick={() => (picker = "customer")} />
    <SelectRow
      icon="folder"
      label="Projekt"
      value={s.project.name}
      placeholder={s.customer.id ? "Auswählen" : "–"}
      disabled={locked || !s.customer.id}
      onclick={() => (picker = "project")}
    />
    <SelectRow
      icon="tag"
      label="Tätigkeit"
      value={s.activity.name}
      placeholder={s.project.id ? "Auswählen" : "–"}
      disabled={locked || !s.project.id}
      onclick={() => (picker = "activity")}
    />
  </section>

  <div class="flex flex-1 flex-col items-center justify-center gap-2 py-2">
    <TimerButton
      running={s.running}
      {elapsed}
      disabled={tracker.busy || (!s.running && (!canStart() || !connected))}
      ontoggle={toggleTimer}
    />
    <div class="h-4 text-[11px] text-fg-subtle">
      {#if tracker.busy}
        {s.running ? "Wird gestoppt …" : "Wird gestartet …"}
      {:else if s.running && s.begin}
        seit {new Date(s.begin).toLocaleTimeString("de-DE", { hour: "2-digit", minute: "2-digit" })} Uhr
      {:else}
        {hint}
      {/if}
    </div>
  </div>

  {#if tracker.error}
    <button
      transition:slide={{ duration: 180 }}
      class="mb-2 rounded-[10px] bg-danger/12 px-3 py-2 text-left text-[12px] leading-snug"
      title="Ausblenden"
      onclick={() => (tracker.error = "")}
    >
      {tracker.error}
    </button>
  {/if}

  {#if assistant.error}
    <button
      transition:slide={{ duration: 180 }}
      class="mb-2 rounded-[10px] bg-danger/12 px-3 py-2 text-left text-[12px] leading-snug"
      onclick={() => (assistant.error = "")}
    >
      {assistant.error}
    </button>
  {:else if assistant.undo !== null}
    <div transition:slide={{ duration: 180 }} class="mb-2 flex items-center justify-between rounded-[10px] bg-accent/10 px-3 py-1.5 text-[12px]">
      <span>In Stichpunkte umgewandelt</span>
      <button class="font-medium text-accent hover:opacity-75" onclick={undoSummary}>Rückgängig</button>
    </div>
  {/if}

  {#if voice.error}
    <button
      transition:slide={{ duration: 180 }}
      class="mb-2 rounded-[10px] bg-warning/15 px-3 py-2 text-left text-[12px] leading-snug"
      title="Ausblenden"
      onclick={() => (voice.error = "")}
    >
      {voice.error}
    </button>
  {/if}

  <DescriptionField
    value={s.description}
    oninput={setDescription}
    status={s.running ? tracker.descriptionStatus : ""}
    micAvailable={voiceEnabled.value}
  />

  <button
    class="mt-2.5 flex min-w-0 items-center justify-center gap-1.5 text-[11px] text-fg-subtle transition-colors hover:text-fg-muted"
    onclick={onOpenSettings}
  >
    <span class="size-1.5 shrink-0 rounded-full transition-colors {status.tone}"></span>
    <span class="truncate">{status.text}</span>
  </button>

  {#if picker === "customer"}
    <Picker
      title="Kunde"
      placeholder="Kunde suchen …"
      selectedId={s.customer.id}
      load={(r) => TrackerService.Customers(r)}
      onselect={onCustomer}
      onclose={() => (picker = null)}
    />
  {:else if picker === "project"}
    <Picker
      title="Projekt"
      placeholder="Projekt suchen …"
      selectedId={s.project.id}
      load={(r) => TrackerService.Projects(s.customer.id, r)}
      onselect={onProject}
      onclose={() => (picker = null)}
    />
  {:else if picker === "activity"}
    <Picker
      title="Tätigkeit"
      placeholder="Tätigkeit suchen …"
      selectedId={s.activity.id}
      load={(r) => TrackerService.Activities(s.project.id, r)}
      onselect={onActivity}
      onclose={() => (picker = null)}
    />
  {/if}
</div>
