<script lang="ts">
  import { BRAND, brandLogoDarkUrl, brandLogoUrl } from "./brand";

  let { height = 20 }: { height?: number } = $props();
</script>

{#if brandLogoUrl}
  {#if brandLogoDarkUrl}
    <img src={brandLogoUrl} alt={BRAND.company} style="height:{height}px" class="w-auto dark:hidden" />
    <img src={brandLogoDarkUrl} alt={BRAND.company} style="height:{height}px" class="hidden w-auto dark:block" />
  {:else}
    <!-- Without a dark variant the logo sits on a light badge in dark mode, so colours stay intact. -->
    <span class="inline-block rounded-md dark:bg-white dark:px-2 dark:py-1">
      <img src={brandLogoUrl} alt={BRAND.company} style="height:{height}px" class="w-auto" />
    </span>
  {/if}
{:else}
  <!-- Text wordmark until an official logo file is added. -->
  <span class="inline-flex items-baseline gap-1 leading-none font-semibold tracking-[-0.01em]" style="font-size:{height * 0.7}px">
    <span class="text-fg">Klein</span><span class="text-accent">IT</span><span class="font-normal text-fg-muted">Solutions</span>
  </span>
{/if}
