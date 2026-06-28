//go:build geoiplive

// Verificación REAL del resolver de país contra ipapi.co y un Redis real.
// Excluido del `go test ./...` normal por el build tag `geoiplive` (requiere red + Redis),
// así no rompe CI. Ejecutar con:
//
//	docker compose up -d redis
//	go test -tags geoiplive -run TestLive_GeoIP -v ./internal/redirect/
//
// REDIS_ADDR opcional (default localhost:6379).
package redirect

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func liveRedis(t *testing.T) *redis.Client {
	t.Helper()
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Fatalf("Redis no disponible en %s: %v", addr, err)
	}
	return rdb
}

func TestLive_GeoIP(t *testing.T) {
	rdb := liveRedis(t)
	ctx := context.Background()
	const publicIP = "8.8.8.8"
	_ = rdb.Del(ctx, geoIPCacheKeyPrefix+publicIP).Err() // partir sin cache previo

	// CRITERIO 1: IP pública nueva (sin cache) resuelve un país real contra ipapi.co.
	r := NewCountryResolver(rdb, 8*time.Second)
	country := r.ResolveCountry(publicIP)
	if !isValidCountryCode(country) {
		t.Fatalf("país real esperado para %s, se obtuvo %q", publicIP, country)
	}
	t.Logf("CRITERIO 1 OK: %s -> %q (país real desde ipapi.co)", publicIP, country)

	// Y se cacheó en Redis con TTL ~48h.
	ttl, err := rdb.TTL(ctx, geoIPCacheKeyPrefix+publicIP).Result()
	if err != nil || ttl <= 47*time.Hour || ttl > 48*time.Hour {
		t.Fatalf("TTL del cache = %v (err %v), se esperaba ~48h", ttl, err)
	}
	t.Logf("cache escrito con TTL %v (~48h)", ttl)

	// CRITERIO 2: repetir la misma IP resuelve desde Redis SIN 2ª llamada HTTP. Prueba:
	// apuntamos la base de la API a una URL rota; si igual devuelve el país, vino del cache.
	prev := geoIPAPIBaseURL
	geoIPAPIBaseURL = "http://127.0.0.1:1/" // puerto cerrado: cualquier HTTP fallaría
	again := r.ResolveCountry(publicIP)
	geoIPAPIBaseURL = prev
	if again != country {
		t.Fatalf("2ª resolución = %q, se esperaba %q desde cache (no debió llamar HTTP)", again, country)
	}
	t.Logf("CRITERIO 2 OK: 2ª resolución %q vino del cache (API rota y aun así resolvió)", again)

	// CRITERIO 3: IP privada/loopback no llama al servicio externo (devuelve unknown ya;
	// además verificamos que NO se escribe cache para ella).
	_ = rdb.Del(ctx, geoIPCacheKeyPrefix+"127.0.0.1").Err()
	if got := r.ResolveCountry("127.0.0.1"); got != countryUnknown {
		t.Fatalf("IP loopback -> %q, se esperaba %q", got, countryUnknown)
	}
	if n, _ := rdb.Exists(ctx, geoIPCacheKeyPrefix+"127.0.0.1").Result(); n != 0 {
		t.Fatalf("no debió escribirse cache para una IP loopback")
	}
	t.Logf("CRITERIO 3 OK: 127.0.0.1 -> unknown sin llamada externa ni cache")

	// CRITERIO 4: ipapi.co que "no responde" -> el timeout corta y degrada a unknown, sin
	// colgar. Simulamos un host no enrutable (blackhole) con un timeout corto para el test.
	const freshIP = "1.1.1.1"
	_ = rdb.Del(ctx, geoIPCacheKeyPrefix+freshIP).Err()
	geoIPAPIBaseURL = "http://10.255.255.1/" // IP no enrutable: la conexión cuelga hasta timeout
	rTimeout := NewCountryResolver(rdb, 2*time.Second)
	start := time.Now()
	got := rTimeout.ResolveCountry(freshIP)
	elapsed := time.Since(start)
	geoIPAPIBaseURL = prev
	if got != countryUnknown {
		t.Fatalf("ante servicio que no responde -> %q, se esperaba %q", got, countryUnknown)
	}
	if elapsed > 3*time.Second {
		t.Fatalf("la resolución tardó %v: el timeout de 2s no cortó", elapsed)
	}
	t.Logf("CRITERIO 4 OK: servicio no responde -> unknown en %v (timeout cortó la espera)", elapsed)

	// CRITERIO 5: Redis caído -> se salta el cache y se resuelve igual contra ipapi.co,
	// sin error fatal. Simulamos "caído" con un cliente apuntado a un puerto cerrado.
	const ip5 = "9.9.9.9"
	downRdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6390"}) // nada escucha ahí
	rDown := NewCountryResolver(downRdb, 8*time.Second)
	got5 := rDown.ResolveCountry(ip5)
	if !isValidCountryCode(got5) {
		t.Fatalf("con Redis caído, país real esperado para %s, se obtuvo %q", ip5, got5)
	}
	t.Logf("CRITERIO 5 OK: con Redis caído, %s -> %q (degradó a llamar directo a ipapi.co)", ip5, got5)
}
