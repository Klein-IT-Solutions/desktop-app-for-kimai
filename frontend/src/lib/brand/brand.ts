/**
 * Publisher branding. Replace logo.svg with the official Klein IT Solutions
 * logo; until then BrandLogo renders a text wordmark.
 */
export const BRAND = {
  company: "Klein IT Solutions",
  website: "https://klein-it.solutions",
  websiteLabel: "klein-it.solutions",
  repository: "https://github.com/klein-it-solutions/kimai-desktop-app",
  license: "https://github.com/klein-it-solutions/kimai-desktop-app/blob/main/LICENSE",
};

// Vite resolves this at build time; an absent file yields an empty record.
const logos = import.meta.glob("./logo.svg", { eager: true, query: "?url", import: "default" }) as Record<string, string>;
export const brandLogoUrl: string | undefined = logos["./logo.svg"];
