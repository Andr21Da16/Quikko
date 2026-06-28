import {
  LayoutDashboard,
  Link2,
  BarChart3,
  User,
  TrendingUp,
} from "lucide-react";
import { cn } from "@/lib/utils/cn";

// Mockup estático y fiel del dashboard real (Agent 16): sidebar, tarjeta de stats,
// gráfico de clics con una barra "en vivo" en acento lima, y lista de URLs. Todo en
// HTML/SVG/Tailwind, sin imágenes rasterizadas.
//
// TODO(dashboard-real): cuando exista el dashboard real, reemplazar este mockup por un
// screenshot del producto.

// Alturas relativas de las barras del gráfico (la última es la "en vivo").
const BARS = [38, 52, 44, 66, 58, 80, 100];

const NAV = [
  { label: "Overview", icon: LayoutDashboard, active: true },
  { label: "Mis URLs", icon: Link2, active: false },
  { label: "Analytics", icon: BarChart3, active: false },
  { label: "Cuenta", icon: User, active: false },
];

const URLS = [
  { code: "quikko.io/lanzamiento", clicks: "1,284" },
  { code: "quikko.io/promo-q3", clicks: "642" },
  { code: "quikko.io/docs", clicks: "318" },
];

export function DashboardMockup() {
  return (
    <div className="overflow-hidden rounded-xl border border-zinc-200 bg-white shadow-2xl shadow-zinc-900/10 dark:border-zinc-800 dark:bg-zinc-900 dark:shadow-black/40">
      {/* Barra de ventana */}
      <div className="flex items-center gap-2 border-b border-zinc-200 bg-zinc-50 px-4 py-2.5 dark:border-zinc-800 dark:bg-zinc-950/60">
        <span className="size-2.5 rounded-full bg-red-400/80" />
        <span className="size-2.5 rounded-full bg-amber-400/80" />
        <span className="size-2.5 rounded-full bg-green-400/80" />
        <span className="ml-3 truncate text-xs text-zinc-400 dark:text-zinc-500">
          app.quikko.io/dashboard
        </span>
      </div>

      <div className="flex">
        {/* Sidebar */}
        <aside className="hidden w-44 shrink-0 flex-col border-r border-zinc-200 p-3 sm:flex dark:border-zinc-800">
          <div className="px-2 pb-3 text-sm font-bold text-brand-600 dark:text-brand-400">
            Quikko<span className="text-accent-400">.</span>
          </div>
          <nav className="flex flex-col gap-1">
            {NAV.map((item) => {
              const Icon = item.icon;
              return (
                <div
                  key={item.label}
                  className={cn(
                    "flex items-center gap-2 rounded-md px-2 py-1.5 text-xs font-medium",
                    item.active
                      ? "bg-brand-50 text-brand-700 dark:bg-brand-900/40 dark:text-brand-300"
                      : "text-zinc-500 dark:text-zinc-400",
                  )}
                >
                  <Icon className="size-3.5" aria-hidden />
                  {item.label}
                </div>
              );
            })}
          </nav>
        </aside>

        {/* Main */}
        <div className="min-w-0 flex-1 space-y-4 p-4">
          {/* Stats + chart */}
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-5">
            {/* Stat card */}
            <div className="rounded-lg border border-zinc-200 p-3 sm:col-span-2 dark:border-zinc-800">
              <div className="flex items-center gap-1.5 text-xs text-zinc-500 dark:text-zinc-400">
                <TrendingUp className="size-3.5 text-accent-500" aria-hidden />
                Total de clics
              </div>
              <div className="mt-1 text-2xl font-bold text-zinc-900 dark:text-zinc-50">
                2,244
              </div>
              <div className="mt-0.5 text-[11px] font-medium text-accent-600 dark:text-accent-400">
                +18% esta semana
              </div>
            </div>

            {/* Bar chart */}
            <div className="rounded-lg border border-zinc-200 p-3 sm:col-span-3 dark:border-zinc-800">
              <div className="mb-2 flex items-center justify-between">
                <span className="text-xs text-zinc-500 dark:text-zinc-400">
                  Clics por hora
                </span>
                <span className="inline-flex items-center gap-1 text-[11px] font-medium text-accent-600 dark:text-accent-400">
                  <span className="relative flex size-1.5">
                    <span className="absolute inline-flex size-full animate-ping rounded-full bg-accent-400 opacity-75" />
                    <span className="relative inline-flex size-1.5 rounded-full bg-accent-500" />
                  </span>
                  en vivo
                </span>
              </div>
              <div className="flex h-20 items-end gap-1.5">
                {BARS.map((h, i) => {
                  const live = i === BARS.length - 1;
                  return (
                    <div
                      key={i}
                      style={{ height: `${h}%` }}
                      className={cn(
                        "flex-1 rounded-sm",
                        live
                          ? "bg-accent-400"
                          : "bg-brand-200 dark:bg-brand-900",
                      )}
                    />
                  );
                })}
              </div>
            </div>
          </div>

          {/* URL list */}
          <div className="rounded-lg border border-zinc-200 dark:border-zinc-800">
            {URLS.map((url, i) => (
              <div
                key={url.code}
                className={cn(
                  "flex items-center justify-between px-3 py-2.5",
                  i !== URLS.length - 1 &&
                    "border-b border-zinc-200 dark:border-zinc-800",
                )}
              >
                <div className="flex min-w-0 items-center gap-2">
                  <span className="flex size-6 shrink-0 items-center justify-center rounded-md bg-brand-50 dark:bg-brand-900/40">
                    <Link2
                      className="size-3 text-brand-600 dark:text-brand-300"
                      aria-hidden
                    />
                  </span>
                  <span className="truncate text-xs font-medium text-zinc-700 dark:text-zinc-200">
                    {url.code}
                  </span>
                </div>
                <span className="shrink-0 text-xs tabular-nums text-zinc-500 dark:text-zinc-400">
                  {url.clicks} clics
                </span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
