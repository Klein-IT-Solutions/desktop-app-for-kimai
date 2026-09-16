import "@fontsource-variable/inter";
import "./app.css";
import { mount } from "svelte";
import App from "./App.svelte";
import { initTheme } from "./lib/theme";
import { initConnection } from "./lib/connection.svelte";

initTheme();
initConnection();

mount(App, { target: document.getElementById("app")! });
