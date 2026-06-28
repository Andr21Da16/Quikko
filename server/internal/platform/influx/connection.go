// Package influx gestiona el cliente compartido de InfluxDB.
package influx

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

const healthTimeout = 10 * time.Second

// Connect crea el cliente InfluxDB y verifica su salud. Falla rápido en boot.
func Connect(url, token, org string) (influxdb2.Client, error) {
	client := influxdb2.NewClient(url, token)

	ctx, cancel := context.WithTimeout(context.Background(), healthTimeout)
	defer cancel()

	health, err := client.Health(ctx)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("influx: health check fallido en %q: %w", url, err)
	}
	if health.Status != "pass" {
		client.Close()
		return nil, fmt.Errorf("influx: estado no saludable en %q: %s", url, health.Status)
	}

	slog.Info("conexión a InfluxDB establecida", "url", url, "org", org)
	return client, nil
}
