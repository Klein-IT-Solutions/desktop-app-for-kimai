import {
  AssistantService,
  type AssistantSettings,
} from "../../bindings/github.com/klein-it-solutions/kimai-desktop-app/internal/app";
import { errorMessage } from "./connection.svelte";
import { setDescription, tracker } from "./tracker.svelte";

/** Local AI (bullet points) state. */
export const assistant = $state({
  settings: null as AssistantSettings | null,
  busy: false,
  error: "",
  // Previous description for "Rückgängig".
  undo: null as string | null,
});

export const assistantReady = () => !!assistant.settings?.enabled && !!assistant.settings?.ready;

export function initAssistant() {
  refreshAssistant();
}

export async function refreshAssistant() {
  try {
    assistant.settings = await AssistantService.GetSettings();
  } catch {
    /* plain browser */
  }
}

let undoTimer: ReturnType<typeof setTimeout> | undefined;

/** Turns the whole description into bullet points (undoable). */
export async function summarizeDescription() {
  const text = tracker.state.description.trim();
  if (!text || assistant.busy) return;
  const bullets = await run(text);
  if (bullets === null) return;
  offerUndo(tracker.state.description);
  setDescription(bullets);
}

/** Converts dictated text to bullets before it is appended (auto mode). */
export async function bulletsForDictation(text: string): Promise<string> {
  if (!assistantReady() || !assistant.settings?.autoDictation) return text;
  const bullets = await run(text);
  return bullets ?? text;
}

async function run(text: string): Promise<string | null> {
  assistant.busy = true;
  assistant.error = "";
  try {
    return await AssistantService.Summarize(text);
  } catch (e) {
    assistant.error = errorMessage(e);
    return null;
  } finally {
    assistant.busy = false;
  }
}

function offerUndo(previous: string) {
  assistant.undo = previous;
  clearTimeout(undoTimer);
  undoTimer = setTimeout(() => (assistant.undo = null), 15000);
}

export function undoSummary() {
  if (assistant.undo === null) return;
  setDescription(assistant.undo);
  assistant.undo = null;
}
