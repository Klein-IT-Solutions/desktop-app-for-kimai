import "@fontsource-variable/inter";
import "./app.css";
import { mount } from "svelte";
import App from "./App.svelte";
import { initTheme } from "./lib/theme";
import { initConnection } from "./lib/connection.svelte";
import { initPlatform } from "./lib/platform";

initTheme();
initConnection();

initPlatform().finally(() => {
  mount(App, { target: document.getElementById("app")! });
});
