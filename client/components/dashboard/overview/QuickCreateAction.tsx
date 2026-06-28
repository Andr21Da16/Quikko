import Link from "next/link";
import { ArrowRight, Plus } from "lucide-react";
import { Card } from "@/components/ui";

export function QuickCreateAction() {
  return (
    <Card className="flex flex-col items-start gap-4 bg-brand-600 p-6 text-white sm:flex-row sm:items-center sm:justify-between dark:bg-brand-700 dark:border-brand-700">
      <div>
        <h2 className="text-base font-semibold">Acorta un nuevo enlace</h2>
        <p className="mt-0.5 text-sm text-brand-100">
          Crea una URL corta con su QR y empieza a medir clics al instante.
        </p>
      </div>
      <Link
        href="/dashboard/urls"
        className="inline-flex h-10 shrink-0 items-center gap-2 rounded-lg bg-white px-4 text-sm font-medium text-brand-700 transition-colors hover:bg-brand-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-offset-2 focus-visible:ring-offset-brand-600"
      >
        <Plus className="size-4" aria-hidden />
        Crear URL
        <ArrowRight className="size-4" aria-hidden />
      </Link>
    </Card>
  );
}
