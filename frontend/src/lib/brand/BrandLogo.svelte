<script lang="ts">
  import { BRAND, brandLogoUrl } from "./brand";

  let { height = 20 }: { height?: number } = $props();
</script>

{#if brandLogoUrl}
  <img src={brandLogoUrl} alt={BRAND.company} style="height:{height}px" class="w-auto dark:brightness-0 dark:invert" />
{:else}
  <!-- Text wordmark until the official logo file (logo.svg) is added. -->
  <span class="inline-flex items-baseline gap-1 leading-none font-semibold tracking-[-0.01em]" style="font-size:{height * 0.7}px">
    <span class="text-fg">Klein</span><span class="text-accent">IT</span><span class="font-normal text-fg-muted">Solutions</span>
  </span>
{/if}
