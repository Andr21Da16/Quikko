package analytics

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

// measurement es el nombre de la serie en InfluxDB (sección de mapeo en domain.go).
const measurement = "clicks"

// influxClickRepository implementa ClickRepository sobre el cliente oficial de
// InfluxDB. La escritura usa el Write API ASÍNCRONO (no el blocking): WritePoint
// solo encola en el buffer del cliente y retorna de inmediato, de modo que la
// goroutine de Agent 3 nunca se queda esperando I/O de red.
type influxClickRepository struct {
	writeAPI api.WriteAPI
	queryAPI api.QueryAPI
	bucket   string
}

// NewClickRepository obtiene los APIs de escritura/consulta del cliente compartido
// (cacheados por org/bucket dentro del cliente) y arranca un consumidor del canal
// de errores async, ya que con el Write API asíncrono los fallos de red no se
// devuelven por valor: llegan por Errors() y, si no se drenan, se pierden.
func NewClickRepository(client influxdb2.Client, org, bucket string) ClickRepository {
	writeAPI := client.WriteAPI(org, bucket)

	go func() {
		for err := range writeAPI.Errors() {
			slog.Error("analytics: error en escritura asíncrona a InfluxDB", "err", err)
		}
	}()

	return &influxClickRepository{
		writeAPI: writeAPI,
		queryAPI: client.QueryAPI(org),
		bucket:   bucket,
	}
}

func (r *influxClickRepository) Write(_ context.Context, e ClickEvent) error {
	// El Write API async no recibe context: WritePoint encola y retorna. El error
	// de red (si lo hay) llega por el canal Errors() que drenamos en el constructor.
	p := influxdb2.NewPoint(
		measurement,
		map[string]string{
			"shortCode":  e.ShortCode,
			"ownerId":    e.OwnerID,
			"country":    e.Country,
			"deviceType": e.DeviceType,
			"browser":    e.Browser,
		},
		map[string]interface{}{"count": int64(1)},
		e.Timestamp,
	)
	r.writeAPI.WritePoint(p)
	return nil
}

func (r *influxClickRepository) QueryStats(ctx context.Context, shortCode, ownerID string, rango TimeRange) (*ClickStats, error) {
	base := r.baseFilter(rango, shortCode, ownerID)

	stats := &ClickStats{
		ClicksByCountry: map[string]int64{},
		ClicksByDevice:  map[string]int64{},
		ClicksByBrowser: map[string]int64{},
		ClicksOverTime:  []TimeBucket{},
	}

	var err error
	if stats.TotalClicks, err = r.queryTotal(ctx, base); err != nil {
		return nil, err
	}
	if stats.ClicksByCountry, err = r.queryGroupedBy(ctx, base, "country"); err != nil {
		return nil, err
	}
	if stats.ClicksByDevice, err = r.queryGroupedBy(ctx, base, "deviceType"); err != nil {
		return nil, err
	}
	if stats.ClicksByBrowser, err = r.queryGroupedBy(ctx, base, "browser"); err != nil {
		return nil, err
	}
	if stats.ClicksOverTime, err = r.queryOverTime(ctx, base, rango.window()); err != nil {
		return nil, err
	}

	return stats, nil
}

// baseFilter construye el prefijo Flux común a todas las consultas: bucket, rango,
// measurement, field y los filtros de ownerId (y shortCode si se especifica).
//
// Los valores se interpolan con %q (string Go-quoted), que escapa comillas y
// barras igual que un literal de cadena Flux: así un shortCode/ownerID arbitrario
// no puede romper ni inyectar la query. rango.fluxStart() es una duración literal
// (-24h), no una cadena, así que va sin comillas.
func (r *influxClickRepository) baseFilter(rango TimeRange, shortCode, ownerID string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "from(bucket: %q)\n", r.bucket)
	fmt.Fprintf(&b, "  |> range(start: %s)\n", rango.fluxStart())
	fmt.Fprintf(&b, "  |> filter(fn: (r) => r._measurement == %q)\n", measurement)
	b.WriteString("  |> filter(fn: (r) => r._field == \"count\")\n")
	fmt.Fprintf(&b, "  |> filter(fn: (r) => r.ownerId == %q)\n", ownerID)
	if shortCode != "" {
		fmt.Fprintf(&b, "  |> filter(fn: (r) => r.shortCode == %q)\n", shortCode)
	}
	return b.String()
}

// queryTotal colapsa todo a una sola tabla y suma: el total de clics del rango.
func (r *influxClickRepository) queryTotal(ctx context.Context, base string) (int64, error) {
	q := base + "  |> group()\n  |> sum()\n"
	result, err := r.queryAPI.Query(ctx, q)
	if err != nil {
		return 0, fmt.Errorf("analytics: query de total: %w", err)
	}
	defer result.Close()

	var total int64
	for result.Next() {
		total += toInt64(result.Record().Value())
	}
	if result.Err() != nil {
		return 0, fmt.Errorf("analytics: leyendo total: %w", result.Err())
	}
	return total, nil
}

// queryGroupedBy reagrupa por una columna de tag y suma: un valor por categoría.
func (r *influxClickRepository) queryGroupedBy(ctx context.Context, base, column string) (map[string]int64, error) {
	q := fmt.Sprintf("%s  |> group(columns: [%q])\n  |> sum()\n", base, column)
	result, err := r.queryAPI.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("analytics: query agrupada por %s: %w", column, err)
	}
	defer result.Close()

	out := map[string]int64{}
	for result.Next() {
		rec := result.Record()
		key, _ := rec.ValueByKey(column).(string)
		if key == "" {
			key = "unknown"
		}
		out[key] += toInt64(rec.Value())
	}
	if result.Err() != nil {
		return nil, fmt.Errorf("analytics: leyendo agrupación por %s: %w", column, result.Err())
	}
	return out, nil
}

// queryOverTime arma la serie temporal: colapsa a una tabla y suma por ventana.
func (r *influxClickRepository) queryOverTime(ctx context.Context, base, window string) ([]TimeBucket, error) {
	q := fmt.Sprintf("%s  |> group()\n  |> aggregateWindow(every: %s, fn: sum, createEmpty: false)\n", base, window)
	result, err := r.queryAPI.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("analytics: query de serie temporal: %w", err)
	}
	defer result.Close()

	buckets := []TimeBucket{}
	for result.Next() {
		rec := result.Record()
		buckets = append(buckets, TimeBucket{Timestamp: rec.Time(), Count: toInt64(rec.Value())})
	}
	if result.Err() != nil {
		return nil, fmt.Errorf("analytics: leyendo serie temporal: %w", result.Err())
	}
	return buckets, nil
}

// toInt64 normaliza el _value de Flux a int64. La suma de un field entero vuelve
// como int64, pero blindamos los otros tipos numéricos por robustez.
func toInt64(v interface{}) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case float64:
		return int64(n)
	case uint64:
		return int64(n)
	default:
		return 0
	}
}
