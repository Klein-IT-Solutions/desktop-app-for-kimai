import { Events } from "@wailsio/runtime";
import {
  VoiceService,
  type VoiceState,
} from "../../bindings/github.com/klein-it-solutions/kimai-desktop-app/internal/app";
import { setDescription, tracker } from "./tracker.svelte";
import { bulletsForDictation } from "./assistant.svelte";

/** Speech input state shared by the mic button and the settings. */
export const voice = $state({
  state: "idle" as "idle" | "recording" | "transcribing",
  level: 0,
  error: "",
});

export function initVoice() {
  VoiceService.State()
    .then(apply)
    .catch(() => {});
  Events.On("voice:state", (ev) => apply(ev.data as VoiceState));
  Events.On("voice:level", (ev) => (voice.level = Number(ev.data) || 0));
  // Recognised text is appended to the description and stays editable.
  Events.On("voice:result", async (ev) => {
    let text = String(ev.data ?? "").trim();
    if (!text) return;
    text = await bulletsForDictation(text);
    const current = tracker.state.description.trimEnd();
    const sep = text.startsWith("- ") || current.includes("\n- ") || current.startsWith("- ") ? "\n" : " ";
    setDescription(current ? `${current}${sep}${text}` : text);
  });
}

function apply(s: VoiceState) {
  voice.state = (s.state as typeof voice.state) || "idle";
  voice.error = s.error ?? "";
  if (voice.state !== "recording") voice.level = 0;
}

export const startRecording = () => VoiceService.Start().catch(() => {});
export const stopRecording = () => VoiceService.Stop().catch(() => {});
export const cancelRecording = () => VoiceService.Cancel().catch(() => {});
