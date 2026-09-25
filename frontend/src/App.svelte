<script lang="ts">
  import { Events } from "@wailsio/runtime";
  import { fly } from "svelte/transition";
  import { cubicOut } from "svelte/easing";
  import TitleBar from "./lib/components/TitleBar.svelte";
  import TrackerView from "./lib/views/TrackerView.svelte";
  import SettingsView from "./lib/views/SettingsView.svelte";

  type View = "tracker" | "settings";
  let view = $state<View>("tracker");

  // The tray menu can open the settings.
  Events.On("navigate", (ev) => {
    if (ev.data === "settings" || ev.data === "tracker") view = ev.data;
  });

  const slide = (dir: 1 | -1) => ({ x: 40 * dir, duration: 260, easing: cubicOut, opacity: 0 });
</script>

<main class="window flex flex-col">
  <TitleBar
    title={view === "settings" ? "Einstellungen" : "Kimai Tracker"}
    canGoBack={view === "settings"}
    onBack={() => (view = "tracker")}
    onSettings={() => (view = "settings")}
  />

  <div class="relative min-h-0 flex-1">
    {#key view}
      <div
        class="absolute inset-0"
        in:fly={slide(view === "settings" ? 1 : -1)}
        out:fly={{ ...slide(view === "settings" ? -1 : 1), duration: 180 }}
      >
        {#if view === "tracker"}
          <TrackerView onOpenSettings={() => (view = "settings")} />
        {:else}
          <SettingsView />
        {/if}
      </div>
    {/key}
  </div>
</main>
