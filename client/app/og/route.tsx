import { ImageResponse } from "next/og";

// Imagen Open Graph (1200x630) generada on-demand con next/og (HTML→imagen, patrón
// nativo de Next, sin dependencias extra ni PNG estático). Compone el wordmark Quikko
// sobre fondo de marca con "aire" alrededor: la última "o" es el isotipo (anillo +
// punto) en acento lima, representando el destino de un enlace.
export const runtime = "nodejs";

const BRAND = "#6D28D9";
const BRAND_DARK = "#4B1B98";
const ACCENT = "#A3E635";

export async function GET() {
  return new ImageResponse(
    (
      <div
        style={{
          width: "1200px",
          height: "630px",
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          gap: "32px",
          background: `linear-gradient(135deg, ${BRAND} 0%, ${BRAND_DARK} 100%)`,
          color: "#ffffff",
          fontFamily: "sans-serif",
        }}
      >
        {/* Wordmark: "Quikk" + nodo de red (anillo + punto) como última "o". */}
        <div style={{ display: "flex", alignItems: "center" }}>
          <span style={{ fontSize: "150px", fontWeight: 700, letterSpacing: "-4px" }}>
            Quikk
          </span>
          <div
            style={{
              width: "104px",
              height: "104px",
              marginLeft: "8px",
              marginTop: "36px",
              borderRadius: "9999px",
              border: `18px solid ${ACCENT}`,
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
            }}
          >
            <div
              style={{
                width: "30px",
                height: "30px",
                borderRadius: "9999px",
                background: ACCENT,
              }}
            />
          </div>
        </div>

        <div style={{ display: "flex", height: "6px", width: "120px", background: ACCENT, borderRadius: "9999px" }} />

        <div
          style={{
            display: "flex",
            fontSize: "40px",
            fontWeight: 500,
            color: "#EDE5FB",
            maxWidth: "880px",
            textAlign: "center",
          }}
        >
          Acorta, comparte y mide tus enlaces en tiempo real
        </div>
      </div>
    ),
    { width: 1200, height: 630 },
  );
}
