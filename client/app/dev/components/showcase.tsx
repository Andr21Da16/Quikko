"use client";

import type { ReactNode } from "react";
import { Mail, Search } from "lucide-react";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Card } from "@/components/ui/Card";
import { Badge } from "@/components/ui/Badge";
import { Spinner } from "@/components/ui/Spinner";
import { ThemeToggle } from "@/components/ThemeToggle";
import { useNotificationsStore } from "@/store/notifications";

// Página interna de desarrollo

function Section({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section className="flex flex-col gap-4">
      <h2 className="text-sm font-semibold uppercase tracking-wide text-zinc-500 dark:text-zinc-400">
        {title}
      </h2>
      {children}
    </section>
  );
}

export function ComponentsShowcase() {
  const notify = useNotificationsStore((s) => s.notify);

  return (
    <div className="min-h-dvh bg-white text-zinc-900 dark:bg-zinc-950 dark:text-zinc-100">
      <div className="mx-auto flex max-w-4xl flex-col gap-12 p-6 lg:p-10">
        <header className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">Componentes UI</h1>
            <p className="text-sm text-zinc-500 dark:text-zinc-400">
              Catálogo interno de los átomos de{" "}
              <code className="text-brand-600 dark:text-brand-400">components/ui/</code>.
            </p>
          </div>
          <ThemeToggle />
        </header>

        <Section title="Button — variantes">
          <div className="flex flex-wrap items-center gap-3">
            <Button variant="primary">Primary</Button>
            <Button variant="secondary">Secondary</Button>
            <Button variant="ghost">Ghost</Button>
            <Button variant="danger">Danger</Button>
          </div>
        </Section>

        <Section title="Button — tamaños y estados">
          <div className="flex flex-wrap items-center gap-3">
            <Button size="sm">Small</Button>
            <Button size="md">Medium</Button>
            <Button size="lg">Large</Button>
            <Button isLoading>Cargando…</Button>
            <Button disabled>Disabled</Button>
          </div>
        </Section>

        <Section title="Input">
          <div className="grid max-w-md gap-4">
            <Input label="Email" type="email" placeholder="tu@email.com" icon={<Mail className="size-4" />} />
            <Input label="Buscar" placeholder="Filtrar URLs…" icon={<Search className="size-4" />} />
            <Input label="Sin icono" placeholder="Texto libre" />
            <Input
              label="Con error"
              placeholder="alias-tomado"
              defaultValue="promo"
              error="El alias ya está en uso. Elige otro."
            />
          </div>
        </Section>

        <Section title="Card">
          <div className="grid gap-4 sm:grid-cols-2">
            <Card>
              <h3 className="font-semibold">Una Card simple</h3>
              <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
                Contenedor base con padding, borde y sombra consistentes.
              </p>
            </Card>
            <Card className="border-brand-200 dark:border-brand-900">
              <h3 className="font-semibold text-brand-700 dark:text-brand-300">
                Card con className extendido
              </h3>
              <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
                El className se mergea con tailwind-merge sin conflictos.
              </p>
            </Card>
          </div>
        </Section>

        <Section title="Badge — variantes semánticas">
          <div className="flex flex-wrap items-center gap-2">
            <Badge variant="neutral">Neutral</Badge>
            <Badge variant="success">Activa</Badge>
            <Badge variant="danger">Inactiva</Badge>
            <Badge variant="warning">Warning</Badge>
            <Badge variant="brand">Pro</Badge>
            <Badge variant="accent">Accent</Badge>
          </div>
        </Section>

        <Section title="Spinner">
          <div className="flex items-center gap-6 text-brand-600 dark:text-brand-400">
            <Spinner size={16} />
            <Spinner size={24} />
            <Spinner size={32} />
          </div>
        </Section>

        <Section title="Toast — dispara una notificación">
          <div className="flex flex-wrap items-center gap-3">
            <Button
              variant="secondary"
              onClick={() => notify("success", "URL creada correctamente.")}
            >
              Success
            </Button>
            <Button
              variant="secondary"
              onClick={() => notify("error", "El alias ya está en uso.")}
            >
              Error
            </Button>
            <Button
              variant="secondary"
              onClick={() => notify("info", "Conexión en tiempo real activa.")}
            >
              Info
            </Button>
          </div>
        </Section>
      </div>
    </div>
  );
}
