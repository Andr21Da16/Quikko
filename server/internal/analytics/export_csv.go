package analytics

import (
	"bytes"
	"encoding/csv"
	"sort"
	"strconv"
)

// utf8BOM se antepone al CSV para que Excel (Windows) detecte la codificación
// UTF-8 y no corrompa caracteres no ASCII. Google Sheets / LibreOffice lo ignoran.
const utf8BOM = "\xEF\xBB\xBF"

// statsToCSV serializa un ClickStats a CSV con secciones por dimensión (decisión
// (b) de la spec: reutiliza el ClickStats tal cual, sin desagregar la combinación
// de las 4 dimensiones). Cada bloque lleva un título "# ..." y su propia cabecera.
//
// Formato pensado para abrir directo en Excel/Sheets/LibreOffice: BOM UTF-8,
// separador de coma estándar y terminador CRLF.
func statsToCSV(stats *ClickStats, tr TimeRange) []byte {
	var buf bytes.Buffer
	buf.WriteString(utf8BOM)

	w := csv.NewWriter(&buf)
	w.UseCRLF = true // CRLF: máxima compatibilidad con Excel en Windows

	writeCountSection(w, "# Totales por país", "country", stats.ClicksByCountry)
	_ = w.Write(nil) // línea en blanco separadora
	writeCountSection(w, "# Totales por dispositivo", "deviceType", stats.ClicksByDevice)
	_ = w.Write(nil)
	writeCountSection(w, "# Totales por navegador", "browser", stats.ClicksByBrowser)
	_ = w.Write(nil)
	writeTimeSection(w, stats.ClicksOverTime, tr)

	w.Flush()
	return buf.Bytes()
}

// writeCountSection escribe un bloque "título / cabecera / filas" ordenado por
// clics descendente (desempate por nombre ascendente, para salida determinista).
func writeCountSection(w *csv.Writer, title, dimHeader string, counts map[string]int64) {
	_ = w.Write([]string{title})
	_ = w.Write([]string{dimHeader, "clicks"})

	type kv struct {
		key string
		val int64
	}
	rows := make([]kv, 0, len(counts))
	for k, v := range counts {
		rows = append(rows, kv{k, v})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].val != rows[j].val {
			return rows[i].val > rows[j].val
		}
		return rows[i].key < rows[j].key
	})

	for _, r := range rows {
		_ = w.Write([]string{r.key, strconv.FormatInt(r.val, 10)})
	}
}

// writeTimeSection escribe la serie temporal en orden cronológico. El bucket se
// formatea como fecha en rangos diarios (7d/30d) y con hora en 24h (buckets
// horarios), para no perder granularidad.
func writeTimeSection(w *csv.Writer, buckets []TimeBucket, tr TimeRange) {
	_ = w.Write([]string{"# Clics por día"})
	_ = w.Write([]string{"date", "clicks"})

	layout := "2006-01-02"
	if tr == Range24h {
		layout = "2006-01-02 15:04"
	}
	for _, b := range buckets {
		_ = w.Write([]string{b.Timestamp.UTC().Format(layout), strconv.FormatInt(b.Count, 10)})
	}
}
