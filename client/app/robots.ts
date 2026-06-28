import type { MetadataRoute } from "next";
import { SITE_URL } from "@/lib/seo/metadata";

// Genera /robots.txt en build (convención App Router). El dashboard es privado: se
// deshabilita su indexación explícitamente aunque ya esté tras login.
export default function robots(): MetadataRoute.Robots {
  return {
    rules: { userAgent: "*", allow: "/", disallow: ["/dashboard/"] },
    sitemap: `${SITE_URL}/sitemap.xml`,
  };
}
