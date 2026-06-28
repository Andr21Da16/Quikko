package analytics

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"strings"
	"testing"
	"time"

	"urlshortener/internal/auth"
	"urlshortener/internal/redirect"
	"urlshortener/internal/shortener"
)

// --- Mocks ---

type mockClickRepo struct {
	writeErr   error
	writeCalls int
	written    []ClickEvent

	queryCalls int
	queryStats *ClickStats
	queryErr   error
	lastCode   string
	lastOwner  string
	lastRange  TimeRange
}

func (m *mockClickRepo) Write(_ context.Context, e ClickEvent) error {
	m.writeCalls++
	m.written = append(m.written, e)
	return m.writeErr
}

func (m *mockClickRepo) QueryStats(_ context.Context, code, owner string, r TimeRange) (*ClickStats, error) {
	m.queryCalls++
	m.lastCode, m.lastOwner, m.lastRange = code, owner, r
	if m.queryErr != nil {
		return nil, m.queryErr
	}
	if m.queryStats != nil {
		return m.queryStats, nil
	}
	return &ClickStats{}, nil
}

type mockPublisher struct {
	calls  int
	events []ClickEvent
}

func (m *mockPublisher) PublishClickEvent(e ClickEvent) {
	m.calls++
	m.events = append(m.events, e)
}

// mockURLRepo implementa shortener.URLRepository; solo FindByCode es relevante
// para la validación de ownership, el resto satisface la interfaz.
type mockURLRepo struct {
	url *shortener.ShortURL
	err error
}

func (m *mockURLRepo) FindByCode(_ context.Context, _ string) (*shortener.ShortURL, error) {
	return m.url, m.err
}
func (m *mockURLRepo) Create(context.Context, *shortener.ShortURL) error { return nil }
func (m *mockURLRepo) FindByID(context.Context, string) (*shortener.ShortURL, error) {
	return nil, shortener.ErrURLNotFound
}
func (m *mockURLRepo) FindByOwner(context.Context, string, shortener.ListFilter, int, int) ([]*shortener.ShortURL, int64, error) {
	return nil, 0, nil
}
func (m *mockURLRepo) ExistsByCode(context.Context, string) (bool, error)        { return false, nil }
func (m *mockURLRepo) CountActiveByOwner(context.Context, string) (int64, error) { return 0, nil }
func (m *mockURLRepo) Update(context.Context, *shortener.ShortURL) error         { return nil }
func (m *mockURLRepo) Delete(context.Context, string, string) error              { return nil }
func (m *mockURLRepo) IncrementClicks(context.Context, string) error             { return nil }
func (m *mockURLRepo) SumClicksByOwner(context.Context, string) (int64, error)   { return 0, nil }
func (m *mockURLRepo) FindShortCodesByOwner(context.Context, string) ([]string, error) {
	return nil, nil
}
func (m *mockURLRepo) DeleteAllByOwner(context.Context, string) error { return nil }

// mockUserRepo implementa auth.UserRepository; solo FindByID importa para el plan.
type mockUserRepo struct {
	plan auth.Plan
}

func (m *mockUserRepo) Create(context.Context, *auth.User) error { return nil }
func (m *mockUserRepo) FindByEmail(context.Context, string) (*auth.User, error) {
	return nil, auth.ErrUserNotFound
}
func (m *mockUserRepo) FindByID(_ context.Context, id string) (*auth.User, error) {
	return &auth.User{ID: id, Plan: m.plan}, nil
}
func (m *mockUserRepo) UpdatePlan(context.Context, string, auth.Plan) error  { return nil }
func (m *mockUserRepo) UpdatePassword(context.Context, string, string) error { return nil }
func (m *mockUserRepo) Delete(context.Context, string) error                 { return nil }

// proUser/freeUser son helpers de legibilidad para los tests.
func proUser() *mockUserRepo  { return &mockUserRepo{plan: auth.PlanPro} }
func freeUser() *mockUserRepo { return &mockUserRepo{plan: auth.PlanFree} }

func sampleEvent() redirect.ClickEvent {
	return redirect.ClickEvent{
		ShortCode:  "abc123",
		OwnerID:    "owner1",
		Country:    "PE",
		DeviceType: "desktop",
		Browser:    "Chrome",
		Timestamp:  time.Now().UTC(),
	}
}

// --- RecordClick: escritura + republicación al hub ---

func TestRecordClick_WritesAndPublishesOnce(t *testing.T) {
	repo := &mockClickRepo{}
	pub := &mockPublisher{}
	svc := NewAnalyticsService(repo, &mockURLRepo{}, proUser(), pub)

	if err := svc.RecordClick(context.Background(), sampleEvent()); err != nil {
		t.Fatalf("esperaba éxito, obtuvo %v", err)
	}
	if repo.writeCalls != 1 {
		t.Fatalf("esperaba 1 Write, hubo %d", repo.writeCalls)
	}
	if pub.calls != 1 {
		t.Fatalf("un clic escrito debe publicarse exactamente una vez, hubo %d", pub.calls)
	}
	// El evento traducido conserva todos los campos del de Agent 3.
	if got := pub.events[0]; got.ShortCode != "abc123" || got.OwnerID != "owner1" || got.Country != "PE" {
		t.Fatalf("evento publicado no conserva los datos: %+v", got)
	}
}

func TestRecordClick_WriteError_DoesNotPublish(t *testing.T) {
	repo := &mockClickRepo{writeErr: errors.New("influx caído")}
	pub := &mockPublisher{}
	svc := NewAnalyticsService(repo, &mockURLRepo{}, proUser(), pub)

	if err := svc.RecordClick(context.Background(), sampleEvent()); err == nil {
		t.Fatal("esperaba que el error de escritura se propagara")
	}
	if pub.calls != 0 {
		t.Fatalf("no se debe publicar si la escritura falló, hubo %d", pub.calls)
	}
}

// --- GetStats: ownership + validación de rango ---

func TestGetStats_ForeignShortCode_Forbidden(t *testing.T) {
	repo := &mockClickRepo{}
	// La URL existe pero es de otro dueño.
	urls := &mockURLRepo{url: &shortener.ShortURL{ShortCode: "abc123", OwnerID: "someoneElse"}}
	svc := NewAnalyticsService(repo, urls, proUser(), &mockPublisher{})

	code := "abc123"
	_, err := svc.GetStats(context.Background(), "intruder", &code, "7d")
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("esperaba ErrForbidden, obtuvo %v", err)
	}
	if repo.queryCalls != 0 {
		t.Fatalf("no debe consultar InfluxDB para una URL ajena, hubo %d queries", repo.queryCalls)
	}
}

func TestGetStats_NonexistentShortCode_Forbidden(t *testing.T) {
	// URL inexistente se trata como ajena: no filtra existencia.
	repo := &mockClickRepo{}
	urls := &mockURLRepo{err: shortener.ErrURLNotFound}
	svc := NewAnalyticsService(repo, urls, proUser(), &mockPublisher{})

	code := "missing"
	_, err := svc.GetStats(context.Background(), "owner1", &code, "7d")
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("esperaba ErrForbidden para código inexistente, obtuvo %v", err)
	}
	if repo.queryCalls != 0 {
		t.Fatalf("no debe consultar InfluxDB, hubo %d queries", repo.queryCalls)
	}
}

func TestGetStats_InvalidRange_NoQuery(t *testing.T) {
	repo := &mockClickRepo{}
	svc := NewAnalyticsService(repo, &mockURLRepo{}, proUser(), &mockPublisher{})

	_, err := svc.GetStats(context.Background(), "owner1", nil, "1y")
	if !errors.Is(err, ErrInvalidRange) {
		t.Fatalf("esperaba ErrInvalidRange, obtuvo %v", err)
	}
	if repo.queryCalls != 0 {
		t.Fatalf("un rango inválido no debe llegar a InfluxDB, hubo %d queries", repo.queryCalls)
	}
}

func TestGetStats_Overview_NoOwnershipCheck(t *testing.T) {
	// Sin shortCode: vista overview, agrega todas las URLs del owner sin validar
	// ownership de una URL puntual. Rango vacío -> default 24h.
	repo := &mockClickRepo{queryStats: &ClickStats{TotalClicks: 5}}
	svc := NewAnalyticsService(repo, &mockURLRepo{}, proUser(), &mockPublisher{})

	stats, err := svc.GetStats(context.Background(), "owner1", nil, "")
	if err != nil {
		t.Fatalf("esperaba éxito, obtuvo %v", err)
	}
	if stats.TotalClicks != 5 {
		t.Fatalf("stats inesperadas: %+v", stats)
	}
	if repo.lastCode != "" {
		t.Fatalf("overview debe consultar con shortCode vacío, fue %q", repo.lastCode)
	}
	if repo.lastOwner != "owner1" {
		t.Fatalf("debe consultar acotado al owner, fue %q", repo.lastOwner)
	}
	if repo.lastRange != Range24h {
		t.Fatalf("rango vacío debe defaultear a 24h, fue %q", repo.lastRange)
	}
}

// Free + 7d/30d => PLAN_RANGE_NOT_ALLOWED, sin tocar InfluxDB.
func TestGetStats_FreePlan_HistoricRangeRejected(t *testing.T) {
	for _, rango := range []string{"7d", "30d"} {
		repo := &mockClickRepo{}
		svc := NewAnalyticsService(repo, &mockURLRepo{}, freeUser(), &mockPublisher{})

		_, err := svc.GetStats(context.Background(), "owner1", nil, rango)
		if !errors.Is(err, ErrRangeNotAllowedForPlan) {
			t.Fatalf("range %s en Free: esperaba ErrRangeNotAllowedForPlan, obtuvo %v", rango, err)
		}
		if repo.queryCalls != 0 {
			t.Fatalf("range %s en Free no debe consultar InfluxDB, hubo %d queries", rango, repo.queryCalls)
		}
	}
}

// Free + 24h => permitido (su rango por defecto).
func TestGetStats_FreePlan_24hAllowed(t *testing.T) {
	repo := &mockClickRepo{queryStats: &ClickStats{TotalClicks: 3}}
	svc := NewAnalyticsService(repo, &mockURLRepo{}, freeUser(), &mockPublisher{})

	stats, err := svc.GetStats(context.Background(), "owner1", nil, "24h")
	if err != nil {
		t.Fatalf("Free con 24h debe funcionar, obtuvo %v", err)
	}
	if stats.TotalClicks != 3 || repo.queryCalls != 1 {
		t.Fatalf("esperaba 1 query con datos, stats=%+v queries=%d", stats, repo.queryCalls)
	}
}

// Pro + 7d => permitido.
func TestGetStats_ProPlan_HistoricRangeAllowed(t *testing.T) {
	repo := &mockClickRepo{queryStats: &ClickStats{TotalClicks: 7}}
	svc := NewAnalyticsService(repo, &mockURLRepo{}, proUser(), &mockPublisher{})

	if _, err := svc.GetStats(context.Background(), "owner1", nil, "7d"); err != nil {
		t.Fatalf("Pro con 7d debe funcionar, obtuvo %v", err)
	}
	if repo.queryCalls != 1 || repo.lastRange != Range7d {
		t.Fatalf("esperaba 1 query con rango 7d, queries=%d range=%q", repo.queryCalls, repo.lastRange)
	}
}

// --- Export CSV (Agent 8) ---

func sampleStats() *ClickStats {
	return &ClickStats{
		TotalClicks:     6,
		ClicksByCountry: map[string]int64{"PE": 4, "US": 2},
		ClicksByDevice:  map[string]int64{"mobile": 5, "desktop": 1},
		ClicksByBrowser: map[string]int64{"Chrome": 6},
		ClicksOverTime: []TimeBucket{
			{Timestamp: time.Date(2026, 6, 25, 0, 0, 0, 0, time.UTC), Count: 2},
			{Timestamp: time.Date(2026, 6, 26, 0, 0, 0, 0, time.UTC), Count: 4},
		},
	}
}

// El CSV exitoso lleva BOM, secciones por dimensión, cabeceras y CRLF, y es
// parseable por encoding/csv (abre en Excel/Sheets sin problemas de encoding).
func TestExportStatsCSV_Structure(t *testing.T) {
	repo := &mockClickRepo{queryStats: sampleStats()}
	svc := NewAnalyticsService(repo, &mockURLRepo{}, proUser(), &mockPublisher{})

	out, err := svc.ExportStatsCSV(context.Background(), "owner1", nil, "7d")
	if err != nil {
		t.Fatalf("export falló: %v", err)
	}

	if !bytes.HasPrefix(out, []byte("\xEF\xBB\xBF")) {
		t.Fatal("el CSV debe empezar con el BOM UTF-8 para Excel")
	}
	if !bytes.Contains(out, []byte("\r\n")) {
		t.Fatal("el CSV debe usar terminador CRLF")
	}

	s := string(out)
	for _, want := range []string{
		"# Totales por país", "country,clicks", "PE,4", "US,2",
		"# Totales por dispositivo", "deviceType,clicks", "mobile,5", "desktop,1",
		"# Totales por navegador", "browser,clicks", "Chrome,6",
		"# Clics por día", "date,clicks", "2026-06-25,2", "2026-06-26,4",
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("el CSV debe contener %q\n---\n%s", want, s)
		}
	}

	// Orden por clics descendente dentro de la sección de país (PE antes que US).
	if strings.Index(s, "PE,4") > strings.Index(s, "US,2") {
		t.Fatal("la sección de país debe ordenarse por clics descendente")
	}

	// El BOM + las líneas de datos deben ser parseables por un lector CSV estándar
	// (saltando los comentarios "#" y las líneas en blanco) sin error de formato.
	body := strings.TrimPrefix(s, "\xEF\xBB\xBF")
	r := csv.NewReader(strings.NewReader(body))
	r.FieldsPerRecord = -1 // las secciones tienen distinto número de columnas
	if _, err := r.ReadAll(); err != nil {
		t.Fatalf("el CSV no es parseable por encoding/csv: %v", err)
	}
}

// La restricción de ownership no se puede evadir por el export.
func TestExportStatsCSV_ForeignShortCode_Forbidden(t *testing.T) {
	repo := &mockClickRepo{}
	urls := &mockURLRepo{url: &shortener.ShortURL{ShortCode: "abc123", OwnerID: "someoneElse"}}
	svc := NewAnalyticsService(repo, urls, proUser(), &mockPublisher{})

	code := "abc123"
	if _, err := svc.ExportStatsCSV(context.Background(), "intruder", &code, "7d"); !errors.Is(err, ErrForbidden) {
		t.Fatalf("esperaba ErrForbidden, obtuvo %v", err)
	}
	if repo.queryCalls != 0 {
		t.Fatalf("una URL ajena no debe consultar InfluxDB, hubo %d queries", repo.queryCalls)
	}
}

// La restricción de plan/rango (Agent 7) no se puede evadir por el export.
func TestExportStatsCSV_FreePlan_HistoricRangeRejected(t *testing.T) {
	repo := &mockClickRepo{}
	svc := NewAnalyticsService(repo, &mockURLRepo{}, freeUser(), &mockPublisher{})

	if _, err := svc.ExportStatsCSV(context.Background(), "owner1", nil, "7d"); !errors.Is(err, ErrRangeNotAllowedForPlan) {
		t.Fatalf("esperaba ErrRangeNotAllowedForPlan en el export, obtuvo %v", err)
	}
	if repo.queryCalls != 0 {
		t.Fatalf("Free con 7d no debe consultar InfluxDB ni generar CSV, hubo %d queries", repo.queryCalls)
	}
}
