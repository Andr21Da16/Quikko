"use client";

import { useRouter } from "next/navigation";
import { LogOut } from "lucide-react";
import { Badge, Button, Card, Spinner } from "@/components/ui";
import { useAuthStore } from "@/store/auth";
import type { AccountSummary } from "@/types";

// Tarjeta de identidad de la cuenta : email + plan actual. Incluye un botón de
// cerrar sesión por conveniencia (además del que ya está en el topbar del DashboardShell);
// el resumen de USO detallado se delega a AccountSummaryCard (reutilizado de Overview).
export function AccountInfoCard({
  summary,
  isLoading,
}: {
  summary: AccountSummary | null;
  isLoading: boolean;
}) {
  const router = useRouter();
  const logout = useAuthStore((s) => s.logout);

  const handleLogout = () => {
    logout();
    router.push("/login");
  };

  if (isLoading && !summary) {
    return (
      <Card className="flex h-24 items-center justify-center">
        <Spinner size={22} className="text-brand-600 dark:text-brand-400" />
      </Card>
    );
  }

  const email = summary?.email ?? "—";
  const plan = summary?.plan ?? null;
  const initial = (summary?.email?.[0] ?? "U").toUpperCase();

  return (
    <Card className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
      <div className="flex items-center gap-3">
        <span className="flex size-11 shrink-0 items-center justify-center rounded-full bg-brand-600 text-base font-semibold text-white">
          {initial}
        </span>
        <div className="min-w-0">
          <p className="truncate font-medium text-zinc-900 dark:text-zinc-50">
            {email}
          </p>
          <div className="mt-0.5 flex items-center gap-2">
            <span className="text-sm text-zinc-500 dark:text-zinc-400">Plan</span>
            {plan && (
              <Badge variant={plan === "pro" ? "brand" : "neutral"}>
                {plan === "pro" ? "Pro" : "Free"}
              </Badge>
            )}
          </div>
        </div>
      </div>

      <Button variant="ghost" onClick={handleLogout} className="shrink-0">
        <LogOut className="size-4" aria-hidden />
        Cerrar sesión
      </Button>
    </Card>
  );
}
