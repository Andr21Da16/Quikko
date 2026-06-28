import { SITE_NAME, SITE_URL } from "@/lib/seo/metadata";
import { LandingNav } from "@/components/landing/LandingNav";
import { Hero } from "@/components/landing/Hero";
import { Features } from "@/components/landing/Features";
import { HowItWorks } from "@/components/landing/HowItWorks";
import { Pricing } from "@/components/landing/Pricing";
import { FinalCta } from "@/components/landing/FinalCta";
import { Footer } from "@/components/landing/Footer";

// JSON-LD (structured data) de la home pública. Tipo SoftwareApplication, simple y
// veraz: NO se inventan aggregateRating ni datos inexistentes. price 0 = plan Free.

const jsonLd = {
  "@context": "https://schema.org",
  "@type": "SoftwareApplication",
  name: SITE_NAME,
  url: SITE_URL,
  applicationCategory: "BusinessApplication",
  operatingSystem: "Web",
  offers: { "@type": "Offer", price: "0", priceCurrency: "USD" },
};

// Landing pública de Quikko (Agent 16), en el grupo de rutas (public). Reemplaza el
// placeholder del Agent 11. Orden fijo de secciones: Nav → Hero → Features → Cómo
// funciona → Pricing → CTA final → Footer. Sin testimonios (no hay usuarios reales).
export default function LandingPage() {
  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
      />
      <LandingNav />
      <main className="flex-1">
        <Hero />
        <Features />
        <HowItWorks />
        <Pricing />
        <FinalCta />
      </main>
      <Footer />
    </>
  );
}
