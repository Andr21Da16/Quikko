package redirect

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// withStubAPI levanta un servidor de prueba como ipapi.co, apunta geoIPAPIBaseURL a él,
// y devuelve un contador de hits + el cleanup. rdb=nil en los resolvers de estos tests:
// así se ejercita además el camino de "sin cache" (degradación ante Redis ausente/caído),
// que debe seguir resolviendo el país yendo directo al servicio externo.
func withStubAPI(t *testing.T, handler http.HandlerFunc) *int64 {
	t.Helper()
	var hits int64
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&hits, 1)
		handler(w, r)
	}))
	prev := geoIPAPIBaseURL
	geoIPAPIBaseURL = ts.URL + "/"
	t.Cleanup(func() {
		geoIPAPIBaseURL = prev
		ts.Close()
	})
	return &hits
}

func TestResolveCountry_PublicIP_ReturnsRealCountry(t *testing.T) {
	hits := withStubAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("US\n")) // ipapi.co devuelve el ISO en texto plano
	})

	r := NewCountryResolver(nil, 8*time.Second)
	if got := r.ResolveCountry("8.8.8.8"); got != "US" {
		t.Fatalf("país = %q, se esperaba %q", got, "US")
	}
	if *hits != 1 {
		t.Fatalf("hits al servicio externo = %d, se esperaba 1", *hits)
	}
}

func TestResolveCountry_PrivateAndLoopback_NoExternalCall(t *testing.T) {
	hits := withStubAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("US"))
	})

	r := NewCountryResolver(nil, 8*time.Second)
	for _, ip := range []string{"127.0.0.1", "192.168.1.10", "10.0.0.1", "::1", "fe80::1"} {
		if got := r.ResolveCountry(ip); got != countryUnknown {
			t.Errorf("IP %s: país = %q, se esperaba %q", ip, got, countryUnknown)
		}
	}
	if *hits != 0 {
		t.Fatalf("hits al servicio externo = %d, se esperaba 0 (IPs privadas/loopback no llaman)", *hits)
	}
}

func TestResolveCountry_InvalidIPString_NoExternalCall(t *testing.T) {
	hits := withStubAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("US"))
	})

	r := NewCountryResolver(nil, 8*time.Second)
	if got := r.ResolveCountry("no-es-una-ip"); got != countryUnknown {
		t.Fatalf("país = %q, se esperaba %q", got, countryUnknown)
	}
	if *hits != 0 {
		t.Fatalf("hits = %d, se esperaba 0", *hits)
	}
}

func TestResolveCountry_Timeout_DegradesToUnknown(t *testing.T) {
	withStubAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(300 * time.Millisecond) // más que el timeout del cliente
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("US"))
	})

	// Timeout corto a propósito para no demorar el test; valida que Client.Timeout corta.
	r := NewCountryResolver(nil, 50*time.Millisecond)

	start := time.Now()
	got := r.ResolveCountry("8.8.8.8")
	elapsed := time.Since(start)

	if got != countryUnknown {
		t.Fatalf("país = %q, se esperaba %q ante timeout", got, countryUnknown)
	}
	if elapsed > 250*time.Millisecond {
		t.Fatalf("la resolución tardó %v: el timeout no cortó la espera", elapsed)
	}
}

func TestResolveCountry_Non200_DegradesToUnknown(t *testing.T) {
	withStubAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden) // ej. rate limit / bloqueo
	})

	r := NewCountryResolver(nil, 8*time.Second)
	if got := r.ResolveCountry("8.8.8.8"); got != countryUnknown {
		t.Fatalf("país = %q, se esperaba %q ante status no-200", got, countryUnknown)
	}
}

func TestResolveCountry_UnexpectedBody_DegradesToUnknown(t *testing.T) {
	withStubAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("RateLimited")) // ipapi.co manda texto, no un ISO válido
	})

	r := NewCountryResolver(nil, 8*time.Second)
	if got := r.ResolveCountry("8.8.8.8"); got != countryUnknown {
		t.Fatalf("país = %q, se esperaba %q ante body inesperado", got, countryUnknown)
	}
}

func TestIsValidCountryCode(t *testing.T) {
	valid := []string{"US", "PE", "DE", "JP"}
	invalid := []string{"", "U", "USA", "us", "U1", "Undefined", "12"}
	for _, s := range valid {
		if !isValidCountryCode(s) {
			t.Errorf("isValidCountryCode(%q) = false, se esperaba true", s)
		}
	}
	for _, s := range invalid {
		if isValidCountryCode(s) {
			t.Errorf("isValidCountryCode(%q) = true, se esperaba false", s)
		}
	}
}
