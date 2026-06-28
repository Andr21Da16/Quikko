"use client";

import { useEffect, useRef, useState } from "react";
import * as d3 from "d3";
import { Card } from "@/components/ui";
import type { TimeBucket, TimeRange } from "@/types";

// Gráfico de clics en el tiempo con D3. D3 se usa AQUÍ (serie temporal, donde
// aporta: escalas de tiempo, ejes, área suavizada); los desgloses categóricos usan listas
// con barras simples (ver BreakdownCard) por ser más legibles y livianos. Colores vía
// CSS custom properties de los tokens de marca (Tailwind v4) para respetar el dark mode.

const HEIGHT = 240;
const MARGIN = { top: 12, right: 16, bottom: 26, left: 34 };

function tickFormatFor(range: TimeRange): (d: Date) => string {
  if (range === "24h") return d3.timeFormat("%H:%M");
  return d3.timeFormat("%d %b");
}

export function ClicksOverTimeChart({
  data,
  range,
}: {
  data: TimeBucket[];
  range: TimeRange;
}) {
  const containerRef = useRef<HTMLDivElement>(null);
  const svgRef = useRef<SVGSVGElement>(null);
  const [width, setWidth] = useState(0);

  // Ancho responsive: ResizeObserver actualiza el estado en su callback (no setState
  // síncrono en el cuerpo del efecto).
  useEffect(() => {
    const el = containerRef.current;
    if (!el) return;
    const ro = new ResizeObserver((entries) => {
      const w = entries[0]?.contentRect.width ?? 0;
      setWidth(w);
    });
    ro.observe(el);
    return () => ro.disconnect();
  }, []);

  useEffect(() => {
    const svg = d3.select(svgRef.current);
    svg.selectAll("*").remove();
    if (width <= 0 || data.length === 0) return;

    const innerW = width - MARGIN.left - MARGIN.right;
    const innerH = HEIGHT - MARGIN.top - MARGIN.bottom;

    const points = data
      .map((b) => ({ date: new Date(b.timestamp), count: b.count }))
      .sort((a, b) => a.date.getTime() - b.date.getTime());

    const x = d3
      .scaleTime()
      .domain(d3.extent(points, (d) => d.date) as [Date, Date])
      .range([0, innerW]);

    const maxCount = d3.max(points, (d) => d.count) ?? 0;
    const y = d3
      .scaleLinear()
      .domain([0, Math.max(maxCount, 1)])
      .nice()
      .range([innerH, 0]);

    const g = svg
      .append("g")
      .attr("transform", `translate(${MARGIN.left},${MARGIN.top})`);

    // Grid horizontal tenue.
    g.append("g")
      .attr("color", "currentColor")
      .style("color", "rgb(161 161 170 / 0.25)") // zinc-400 @ 25%
      .call(
        d3
          .axisLeft(y)
          .ticks(4)
          .tickSize(-innerW)
          .tickFormat(() => ""),
      )
      .call((sel) => sel.select(".domain").remove());

    // Eje Y (valores).
    g.append("g")
      .call(d3.axisLeft(y).ticks(4).tickSize(0))
      .call((sel) => sel.select(".domain").remove())
      .call((sel) =>
        sel
          .selectAll("text")
          .attr("class", "fill-zinc-400 text-[10px]")
          .attr("dx", "-2"),
      );

    // Eje X (tiempo).
    const ticks = Math.min(points.length, range === "24h" ? 6 : 7);
    g.append("g")
      .attr("transform", `translate(0,${innerH})`)
      .call(
        d3
          .axisBottom(x)
          .ticks(ticks)
          .tickSize(0)
          .tickPadding(8)
          .tickFormat((d) => tickFormatFor(range)(d as Date)),
      )
      .call((sel) => sel.select(".domain").remove())
      .call((sel) =>
        sel.selectAll("text").attr("class", "fill-zinc-400 text-[10px]"),
      );

    const brand = "var(--color-brand-600, #6D28D9)";

    // Área bajo la curva.
    const area = d3
      .area<{ date: Date; count: number }>()
      .x((d) => x(d.date))
      .y0(innerH)
      .y1((d) => y(d.count))
      .curve(d3.curveMonotoneX);

    g.append("path")
      .datum(points)
      .attr("fill", brand)
      .attr("fill-opacity", 0.12)
      .attr("d", area);

    // Línea.
    const line = d3
      .line<{ date: Date; count: number }>()
      .x((d) => x(d.date))
      .y((d) => y(d.count))
      .curve(d3.curveMonotoneX);

    g.append("path")
      .datum(points)
      .attr("fill", "none")
      .attr("stroke", brand)
      .attr("stroke-width", 2)
      .attr("d", line);

    // Punto final destacado (último valor).
    const last = points[points.length - 1];
    g.append("circle")
      .attr("cx", x(last.date))
      .attr("cy", y(last.count))
      .attr("r", 3.5)
      .attr("fill", brand);
  }, [data, width, range]);

  const totalInRange = data.reduce((acc, b) => acc + b.count, 0);

  return (
    <Card>
      <div className="mb-3 flex items-center justify-between">
        <h2 className="text-sm font-semibold text-zinc-900 dark:text-zinc-50">
          Clics en el tiempo
        </h2>
        <span className="text-xs text-zinc-400">
          {new Intl.NumberFormat("es").format(totalInRange)} en este rango
        </span>
      </div>
      <div ref={containerRef} className="w-full text-zinc-500 dark:text-zinc-400">
        {totalInRange === 0 ? (
          <div className="flex h-[240px] items-center justify-center text-sm text-zinc-400">
            Sin clics en este rango todavía.
          </div>
        ) : (
          <svg
            ref={svgRef}
            width={width}
            height={HEIGHT}
            role="img"
            aria-label="Gráfico de clics en el tiempo"
          />
        )}
      </div>
    </Card>
  );
}
