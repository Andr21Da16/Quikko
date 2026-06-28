import type { ReactNode } from "react";
import { DashboardShell } from "@/components/dashboard/DashboardShell";

// Layout del grupo (dashboard): envuelve TODA página de /dashboard/* en el shell de
// navegación compartido. 
export default function DashboardLayout({ children }: { children: ReactNode }) {
  return <DashboardShell>{children}</DashboardShell>;
}
