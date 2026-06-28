import { AppWindow, Globe, MonitorSmartphone } from "lucide-react";
import type { ClickStats, TimeRange } from "@/types";
import { ClicksOverTimeChart } from "./ClicksOverTimeChart";
import { BreakdownCard } from "./BreakdownCard";


const regionNames = new Intl.DisplayNames(["es"], { type: "region" });
function countryLabel(code: string): string {
  if (!code || code.toLowerCase() === "unknown") return "Desconocido";
  try {
    return regionNames.of(code.toUpperCase()) ?? code;
  } catch {
    return code;
  }
}

export function StatsPanels({
  stats,
  range,
}: {
  stats: ClickStats;
  range: TimeRange;
}) {
  return (
    <div className="space-y-6">
      <ClicksOverTimeChart data={stats.clicksOverTime} range={range} />

      <div className="grid gap-4 md:grid-cols-3">
        <BreakdownCard
          title="Países"
          icon={<Globe className="size-4" />}
          data={stats.clicksByCountry}
          labelMap={countryLabel}
        />
        <BreakdownCard
          title="Dispositivos"
          icon={<MonitorSmartphone className="size-4" />}
          data={stats.clicksByDevice}
        />
        <BreakdownCard
          title="Navegadores"
          icon={<AppWindow className="size-4" />}
          data={stats.clicksByBrowser}
        />
      </div>
    </div>
  );
}
