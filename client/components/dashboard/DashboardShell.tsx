"use client";

import { useEffect, useState, type ReactNode } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { motion, AnimatePresence } from "framer-motion";
import {
  LayoutDashboard,
  Link2,
  BarChart3,
  User,
  Menu,
  X,
  ChevronDown,
  LogOut,
  type LucideIcon,
} from "lucide-react";
import { cn } from "@/lib/utils/cn";
import { Badge } from "@/components/ui/Badge";
import { ThemeToggle } from "@/components/ThemeToggle";
import { useUIStore } from "@/store/ui";
import { useAuthStore } from "@/store/auth";

type NavItem = {
  href: string;
  label: string;
  icon: LucideIcon;
  exact?: boolean;
};

const NAV_ITEMS: NavItem[] = [
  { href: "/dashboard", label: "Overview", icon: LayoutDashboard, exact: true },
  { href: "/dashboard/urls", label: "Mis URLs", icon: Link2 },
  { href: "/dashboard/analytics", label: "Analytics", icon: BarChart3 },
  { href: "/dashboard/account", label: "Cuenta", icon: User },
];

function isActive(pathname: string, item: NavItem): boolean {
  return item.exact ? pathname === item.href : pathname.startsWith(item.href);
}

// Contenido del sidebar, reutilizado por el sidebar fijo (desktop) y el drawer (mobile).
function SidebarNav({ pathname }: { pathname: string }) {
  return (
    <nav className="flex flex-col gap-1 p-3">
      {NAV_ITEMS.map((item) => {
        const active = isActive(pathname, item);
        const Icon = item.icon;
        return (
          <Link
            key={item.href}
            href={item.href}
            aria-current={active ? "page" : undefined}
            className={cn(
              "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
              active
                ? "bg-brand-50 text-brand-700 dark:bg-brand-900/40 dark:text-brand-300"
                : "text-zinc-600 hover:bg-zinc-100 hover:text-zinc-900 dark:text-zinc-300 dark:hover:bg-zinc-800 dark:hover:text-zinc-50",
            )}
          >
            <Icon className="size-4 shrink-0" aria-hidden />
            {item.label}
          </Link>
        );
      })}
    </nav>
  );
}

function BrandHeader() {
  return (
    <div className="flex h-14 items-center px-4">
      <Link
        href="/dashboard"
        className="text-lg font-bold tracking-tight text-brand-600 dark:text-brand-400"
      >
        Quikko<span className="text-accent-400">.</span>
      </Link>
    </div>
  );
}

function UserMenu() {
  const user = useAuthStore((s) => s.user);
  const logout = useAuthStore((s) => s.logout);
  const router = useRouter();
  const [open, setOpen] = useState(false);

  const email = user?.email ?? "Cuenta";
  const initial = (user?.email?.[0] ?? "U").toUpperCase();

  const handleLogout = () => {
    setOpen(false);
    logout();
    router.push("/login");
  };

  return (
    <div className="relative">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        aria-haspopup="menu"
        aria-expanded={open}
        className="flex items-center gap-2 rounded-lg px-2 py-1.5 text-sm text-zinc-700 transition-colors hover:bg-zinc-100 dark:text-zinc-200 dark:hover:bg-zinc-800"
      >
        <span className="flex size-7 items-center justify-center rounded-full bg-brand-600 text-xs font-semibold text-white">
          {initial}
        </span>
        <span className="hidden max-w-[12rem] truncate sm:inline">{email}</span>
        <ChevronDown className="size-4 text-zinc-400" aria-hidden />
      </button>

      {open && (
        <>
          {/* Capa para cerrar al hacer click fuera. */}
          <button
            type="button"
            aria-hidden
            tabIndex={-1}
            onClick={() => setOpen(false)}
            className="fixed inset-0 z-40 cursor-default"
          />
          <div
            role="menu"
            className="absolute right-0 z-50 mt-2 w-48 overflow-hidden rounded-lg border border-zinc-200 bg-white py-1 shadow-lg dark:border-zinc-800 dark:bg-zinc-900"
          >
            <Link
              href="/dashboard/account"
              role="menuitem"
              onClick={() => setOpen(false)}
              className="flex items-center gap-2 px-3 py-2 text-sm text-zinc-700 hover:bg-zinc-100 dark:text-zinc-200 dark:hover:bg-zinc-800"
            >
              <User className="size-4" aria-hidden />
              Cuenta
            </Link>
            <button
              type="button"
              role="menuitem"
              onClick={handleLogout}
              className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/30"
            >
              <LogOut className="size-4" aria-hidden />
              Cerrar sesión
            </button>
          </div>
        </>
      )}
    </div>
  );
}

function PlanBadge() {
  const plan = useAuthStore((s) => s.user?.plan);
  if (!plan) return null;
  return (
    <Badge variant={plan === "pro" ? "brand" : "neutral"}>
      {plan === "pro" ? "Pro" : "Free"}
    </Badge>
  );
}

export function DashboardShell({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const sidebarOpen = useUIStore((s) => s.sidebarOpen);
  const toggleSidebar = useUIStore((s) => s.toggleSidebar);
  const setSidebarOpen = useUIStore((s) => s.setSidebarOpen);

  // Cierre automático del drawer al navegar a otra ruta (mobile): no debe quedar
  // tapando la página recién cargada.
  useEffect(() => {
    setSidebarOpen(false);
  }, [pathname, setSidebarOpen]);

  // Hidratar el `user` al entrar al dashboard: solo se persisten los tokens (no el user),
  // así que tras un F5 el badge de plan / email del topbar estarían vacíos. fetchCurrentUser
  // es idempotente y no hace nada si ya hay user o no hay token. (setState va dentro del
  // store, no síncrono en este effect.)
  useEffect(() => {
    if (!useAuthStore.getState().user) {
      void useAuthStore.getState().fetchCurrentUser();
    }
  }, []);

  return (
    <div className="flex min-h-dvh bg-white text-zinc-900 dark:bg-zinc-950 dark:text-zinc-100">
      {/* Sidebar fijo (desktop, lg+). */}
      <aside className="hidden w-60 shrink-0 flex-col border-r border-zinc-200 dark:border-zinc-800 lg:flex">
        <BrandHeader />
        <SidebarNav pathname={pathname} />
      </aside>

      {/* Drawer + overlay (mobile/tablet, < lg). */}
      <AnimatePresence>
        {sidebarOpen && (
          <motion.div
            key="overlay"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={() => setSidebarOpen(false)}
            className="fixed inset-0 z-40 bg-black/50 lg:hidden"
            aria-hidden
          />
        )}
        {sidebarOpen && (
          <motion.aside
            key="drawer"
            initial={{ x: "-100%" }}
            animate={{ x: 0 }}
            exit={{ x: "-100%" }}
            transition={{ type: "spring", stiffness: 360, damping: 36 }}
            className="fixed inset-y-0 left-0 z-50 flex w-64 flex-col border-r border-zinc-200 bg-white dark:border-zinc-800 dark:bg-zinc-950 lg:hidden"
          >
            <div className="flex items-center justify-between pr-2">
              <BrandHeader />
              <button
                type="button"
                onClick={() => setSidebarOpen(false)}
                aria-label="Cerrar menú"
                className="inline-flex size-9 items-center justify-center rounded-lg text-zinc-500 hover:bg-zinc-100 dark:text-zinc-400 dark:hover:bg-zinc-800"
              >
                <X className="size-5" aria-hidden />
              </button>
            </div>
            <SidebarNav pathname={pathname} />
          </motion.aside>
        )}
      </AnimatePresence>

      {/* Columna principal. */}
      <div className="flex min-w-0 flex-1 flex-col">
        <header className="flex h-14 items-center gap-3 border-b border-zinc-200 px-4 dark:border-zinc-800">
          <button
            type="button"
            onClick={toggleSidebar}
            aria-label="Abrir menú"
            className="inline-flex size-9 items-center justify-center rounded-lg text-zinc-600 hover:bg-zinc-100 dark:text-zinc-300 dark:hover:bg-zinc-800 lg:hidden"
          >
            <Menu className="size-5" aria-hidden />
          </button>

          <div className="ml-auto flex items-center gap-2">
            <PlanBadge />
            <ThemeToggle />
            <UserMenu />
          </div>
        </header>

        <main className="flex-1 p-4 lg:p-6">{children}</main>
      </div>
    </div>
  );
}
