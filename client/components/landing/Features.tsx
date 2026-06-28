import {
  Zap,
  Radio,
  Tag,
  QrCode,
  FileDown,
  SlidersHorizontal,
  type LucideIcon,
} from "lucide-react";
import { Reveal } from "./Reveal";

// Features. Grid de 6 con BORDES COMPARTIDOS (no cajas aisladas con sombra
// propia) — el patrón que usan las referencias para que el grid no se sienta "plano".
type Feature = { icon: LucideIcon; title: string; desc: string };

const FEATURES: Feature[] = [
  {
    icon: Zap,
    title: "Redirección ultra rápida",
    desc: "Redis sirve cada redirección en milisegundos, sin tocar la base de datos en el camino caliente.",
  },
  {
    icon: Radio,
    title: "Analytics en tiempo real",
    desc: "Cada clic aparece al instante en tu dashboard vía WebSocket, sin recargar ni esperar reportes.",
  },
  {
    icon: Tag,
    title: "Alias personalizados",
    desc: "Elige un código propio y memorable en vez de uno aleatorio para tus campañas.",
  },
  {
    icon: QrCode,
    title: "QR automático",
    desc: "Cada enlace incluye su código QR, listo para imprimir o compartir donde lo necesites.",
  },
  {
    icon: FileDown,
    title: "Export CSV",
    desc: "Descarga tus métricas en CSV y analízalas en tu hoja de cálculo favorita.",
  },
  {
    icon: SlidersHorizontal,
    title: "Control total de tus enlaces",
    desc: "Activa, desactiva o elimina cualquier enlace corto cuando quieras, al instante.",
  },
];

export function Features() {
  return (
    <section
      id="features"
      className="border-b border-zinc-200 bg-white py-20 sm:py-28 dark:border-zinc-900 dark:bg-zinc-950"
    >
      <div className="mx-auto max-w-6xl px-4 sm:px-6">
        <Reveal className="mx-auto max-w-2xl text-center">
          <h2 className="text-3xl font-bold tracking-tight text-zinc-900 sm:text-4xl dark:text-white">
            Todo lo que necesita un enlace corto serio
          </h2>
          <p className="mt-4 text-lg text-zinc-600 dark:text-zinc-400">
            Velocidad, métricas en vivo y control — sin complicaciones.
          </p>
        </Reveal>

        <div className="mt-14 grid grid-cols-1 overflow-hidden rounded-2xl border border-zinc-200 bg-white sm:grid-cols-2 lg:grid-cols-3 dark:border-zinc-800 dark:bg-zinc-950">
          {FEATURES.map((feature, i) => {
            const Icon = feature.icon;
            return (
              <Reveal
                key={feature.title}
                delay={(i % 3) * 0.06}
                className="border-b border-r border-zinc-200 dark:border-zinc-800"
              >
                <div className="group h-full p-6 transition-colors hover:bg-zinc-50 dark:hover:bg-zinc-900/60">
                  <span className="inline-flex size-10 items-center justify-center rounded-lg bg-brand-50 text-brand-600 transition-colors group-hover:bg-brand-100 dark:bg-brand-900/40 dark:text-brand-300 dark:group-hover:bg-brand-900/70">
                    <Icon className="size-5" aria-hidden />
                  </span>
                  <h3 className="mt-4 font-semibold text-zinc-900 dark:text-white">
                    {feature.title}
                  </h3>
                  <p className="mt-1.5 text-sm leading-relaxed text-zinc-600 dark:text-zinc-400">
                    {feature.desc}
                  </p>
                </div>
              </Reveal>
            );
          })}
        </div>
      </div>
    </section>
  );
}
