<script lang="ts">
  import {
    AppService,
    SettingsService,
    type ConnectionResult,
    type Info,
    type Settings,
  } from "../../../bindings/github.com/klein-it-solutions/kimai-desktop-app/internal/app";
  import Button from "../components/Button.svelte";
  import Icon from "../components/Icon.svelte";
  import StatusBanner from "../components/StatusBanner.svelte";
  import TextField from "../components/TextField.svelte";
  import { connection, errorMessage, refreshConnection } from "../connection.svelte";

  const TOKEN_DOCS = "https://www.kimai.org/documentation/rest-api.html";
  const LANGUAGES = [
    ["de", "Deutsch"],
    ["en", "English"],
    ["fr", "Français"],
    ["es", "Español"],
    ["it", "Italiano"],
    ["nl", "Nederlands"],
  ] as const;

  let settings = $state<Settings | null>(null);
  let info = $state<Info | null>(null);

  let url = $state("");
  let token = $state("");
  let language = $state("de");
  let showToken = $state(false);

  let testing = $state(false);
  let saving = $state(false);
  let result = $state<ConnectionResult | null>(null);
  let error = $state("");
  let keyringFailure = $state("");
  // Only shown while the keyring is the selected storage.
  const keyringProblem = $derived(
    keyringFailure || (settings?.tokenStorage !== "file" ? (settings?.keyringError ?? "") : ""),
  );

  const dirty = $derived(
    settings !== null &&
      (url.trim() !== settings.kimaiUrl || token.trim() !== "" || language !== settings.language),
  );

  load();

  async function load() {
    try {
      settings = await SettingsService.GetSettings();
      url = settings.kimaiUrl;
      language = settings.language || "de";
      info = await AppService.Info();
      // Show the state of the stored connection right away.
      const r = connection.result;
      if (r && r.status !== "not_configured" && r.status !== "keyring") result = r;
    } catch (e) {
      error = errorMessage(e);
    }
  }

  async function test() {
    error = "";
    testing = true;
    try {
      result = await SettingsService.TestConnection(url, token);
    } catch (e) {
      error = errorMessage(e);
    } finally {
      testing = false;
    }
  }

  /** Saves the form. tokenStorage switches between keyring and file ("" keeps it). */
  async function save(tokenStorage = "") {
    error = "";
    keyringFailure = "";
    saving = true;
    try {
      const res = await SettingsService.SaveSettings({ kimaiUrl: url, token, language, tokenStorage });
      settings = res.settings;
      url = settings.kimaiUrl;
      if (res.keyringFailed) {
        // Keep the typed token so the user can choose the file fallback.
        keyringFailure = res.message ?? "";
        result = null;
        return;
      }
      if (res.message) error = res.message;
      token = "";
      showToken = false;
      result = await refreshConnection();
    } catch (e) {
      error = errorMessage(e);
    } finally {
      saving = false;
    }
  }

  async function removeToken() {
    error = "";
    try {
      await SettingsService.RemoveToken();
      settings = await SettingsService.GetSettings();
      result = null;
      await refreshConnection();
    } catch (e) {
      error = errorMessage(e);
    }
  }

  const sessionLabel = (t: string) => (t === "wayland" ? "Wayland" : t === "x11" ? "X11" : "–");
  const group =
    "overflow-hidden rounded-[10px] bg-field shadow-[inset_0_0_0_0.5px_var(--hairline)] [&>*+*]:border-t [&>*+*]:border-hairline";
  const heading = "mb-1.5 px-1 text-[11px] font-medium tracking-wide text-fg-muted uppercase";
</script>

<div class="h-full overflow-y-auto px-4 pt-1 pb-4">
  <h2 class={heading}>Kimai</h2>
  <section class={group}>
    <TextField label="Server" type="url" placeholder="https://kimai.example.com" bind:value={url} />
    <TextField
      label="API-Token"
      type={showToken ? "text" : "password"}
      placeholder={settings?.hasToken
        ? settings.tokenStorage === "file"
          ? "In Datei gespeichert"
          : "Im Schlüsselbund gespeichert"
        : "Token einfügen"}
      bind:value={token}
    >
      {#snippet trailing()}
        {#if token}
          <button
            class="no-drag -mr-1 grid size-6 place-items-center rounded text-fg-subtle hover:text-fg"
            aria-label={showToken ? "Token verbergen" : "Token anzeigen"}
            onclick={(e) => {
              e.preventDefault();
              showToken = !showToken;
            }}
          >
            <Icon name={showToken ? "eyeOff" : "eye"} size={14} />
          </button>
        {/if}
      {/snippet}
    </TextField>
  </section>

  <div class="mt-1.5 flex items-center justify-between px-1 text-[11px] text-fg-muted">
    <span>Kimai → Profil → API-Zugang</span>
    <div class="flex gap-3">
      {#if settings?.hasToken}
        <button class="text-danger hover:opacity-75" onclick={removeToken}>Token entfernen</button>
      {/if}
      <button class="flex items-center gap-1 text-accent hover:opacity-75" onclick={() => AppService.OpenURL(TOKEN_DOCS)}>
        Hilfe <Icon name="external" size={11} stroke={2} />
      </button>
    </div>
  </div>

  <div class="mt-3 flex gap-2">
    <Button onclick={test} busy={testing} disabled={!url.trim() || (!token.trim() && !settings?.hasToken)}>
      Verbindung testen
    </Button>
    <Button variant="primary" onclick={() => save()} busy={saving} disabled={!dirty}>Speichern</Button>
  </div>

  {#if keyringProblem}
    <div class="mt-3 rounded-[10px] bg-danger/12 px-3 py-2.5 text-[12px] leading-snug">
      <div class="font-medium">
        {keyringFailure ? "Token konnte nicht im Schlüsselbund gespeichert werden." : "Der Schlüsselbund ist nicht verfügbar."}
      </div>
      <details class="mt-1 text-[11px] text-fg-muted">
        <summary class="cursor-pointer">Details</summary>
        <div class="mt-1 break-words select-text">{keyringProblem}</div>
      </details>
      {#if token.trim()}
        <button class="mt-2 font-medium text-accent hover:opacity-75" onclick={() => save("file")}>
          Stattdessen unverschlüsselt in Datei speichern
        </button>
      {:else}
        <div class="mt-2 text-fg-muted">Token eingeben, um ihn stattdessen in einer Datei zu speichern.</div>
      {/if}
    </div>
  {/if}

  {#if settings?.tokenStorage === "file"}
    <div class="mt-3 rounded-[10px] bg-warning/15 px-3 py-2.5 text-[12px] leading-snug">
      Der Token liegt unverschlüsselt (nur für dich lesbar) in
      <span class="break-all text-fg-muted select-text">{settings.secretsPath}</span>.
      {#if settings.hasToken}
        <button class="mt-1 block font-medium text-accent hover:opacity-75" onclick={() => save("keyring")}>
          In den Schlüsselbund verschieben
        </button>
      {/if}
    </div>
  {/if}

  {#if error}
    <div class="mt-3 rounded-[10px] bg-danger/12 px-3 py-2.5 text-[12px] leading-snug">{error}</div>
  {:else if result && result.status !== "keyring"}
    {#key result.checkedAt}
      <div class="mt-3"><StatusBanner {result} /></div>
    {/key}
  {/if}

  <h2 class="mt-5 {heading}">Allgemein</h2>
  <section class={group}>
    <label class="flex h-10 items-center justify-between gap-2 px-3">
      <span>Sprache</span>
      <select
        bind:value={language}
        class="cursor-pointer appearance-none bg-transparent pr-1 text-right text-fg-muted outline-none"
      >
        {#each LANGUAGES as [code, name]}
          <option value={code}>{name}</option>
        {/each}
      </select>
    </label>
  </section>
  <p class="mt-1.5 px-1 text-[11px] text-fg-muted">Wird für die Spracheingabe verwendet.</p>

  <h2 class="mt-5 {heading}">Info</h2>
  <section class={group}>
    {#each [["Version", info?.version ?? "–"], ["Plattform", info?.platform ?? "–"], ["Sitzung", sessionLabel(info?.sessionType ?? "")]] as [k, v]}
      <div class="flex h-9 items-center justify-between px-3">
        <span>{k}</span>
        <span class="text-fg-muted">{v}</span>
      </div>
    {/each}
    <div class="flex h-9 items-center justify-between gap-3 px-3">
      <span class="shrink-0">Konfiguration</span>
      <span class="truncate text-fg-muted" title={settings?.configPath}>{settings?.configPath ?? "–"}</span>
    </div>
  </section>
</div>
