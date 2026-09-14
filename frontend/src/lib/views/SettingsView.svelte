<script lang="ts">
  import { AppService, type Info } from "../../../bindings/github.com/klein-it-solutions/kimai-desktop-app/internal/app";

  let info = $state<Info | null>(null);
  AppService.Info()
    .then((i) => (info = i))
    .catch(() => {});

  const sessionLabel = (t: string) => (t === "wayland" ? "Wayland" : t === "x11" ? "X11" : "–");
</script>

<div class="h-full overflow-y-auto px-4 pb-4">
  <h2 class="mb-1.5 px-1 text-[11px] font-medium tracking-wide text-fg-muted uppercase">Kimai</h2>
  <section class="rounded-[10px] bg-field p-3 text-fg-muted shadow-[inset_0_0_0_0.5px_var(--hairline)]">
    Server-URL, API-Token und „Verbindung testen“ folgen in Schritt 2.
  </section>

  <h2 class="mt-5 mb-1.5 px-1 text-[11px] font-medium tracking-wide text-fg-muted uppercase">Info</h2>
  <section
    class="overflow-hidden rounded-[10px] bg-field shadow-[inset_0_0_0_0.5px_var(--hairline)] [&>*+*]:border-t [&>*+*]:border-hairline"
  >
    {#each [["Version", info?.version ?? "–"], ["Plattform", info?.platform ?? "–"], ["Sitzung", sessionLabel(info?.sessionType ?? "")], ["Darstellung", info ? (info.darkMode ? "Dunkel" : "Hell") : "–"]] as [k, v]}
      <div class="flex h-9 items-center justify-between px-3">
        <span>{k}</span>
        <span class="text-fg-muted">{v}</span>
      </div>
    {/each}
  </section>
</div>
