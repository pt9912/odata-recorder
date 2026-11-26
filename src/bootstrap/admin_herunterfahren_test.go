package bootstrap_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
	"github.com/pt9912/odata-recorder/src/hexagon/services/session"
)

// zweiLiteral ist ein hochgefahrener Handle mit zwei Fakes der Server auf einem Protokoll.
type zweiLiteral struct {
	h          *bootstrap.Handle
	p          *protokoll
	odata, adm *probe
	s          *fakeSessions
	uhr        *supporttest.FestUhr
}

func neuZwei(modus string, frist time.Duration, setze func(*zweiLiteral), ablage error) *zweiLiteral {
	z := &zweiLiteral{p: &protokoll{}, uhr: supporttest.NewFestUhr(time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC))}
	z.odata, z.adm, z.s = neuProbe("odata", z.p), neuProbe("admin", z.p), &fakeSessions{p: z.p}
	if setze != nil {
		setze(z)
	}
	z.h = bootstrap.LiteralHandleZwei(modus, frist, z.uhr, z.s, z.odata, z.adm, &protSchliesser{h: &haken{p: z.p}, c: nichtsSchliessen{ablage}})
	return z
}

// TestAdminHerunterfahren (B09): 4 Modi × {kein Fehler, CloseAll, Schliessen OData, Schliessen Admin, Ablage, alle} = 24
// Zellen: die Folge der Schritte, das Warten auf beide Serve vor der Ablage und die Folge der Fehler im errors.Join.
func TestAdminHerunterfahren(t *testing.T) {
	errC, errO, errA, errD := errors.New("c"), errors.New("o"), errors.New("a"), errors.New("d")
	zellen := 0
	for _, modus := range vierModi() {
		for _, quelle := range []string{"keine", "CloseAll", "OData", "Admin", "Ablage", "alle"} {
			zellen++
			waehle := func(name string, e error) error {
				if quelle == name || quelle == "alle" {
					return e
				}
				return nil
			}
			z := neuZwei(modus, time.Second, func(z *zweiLiteral) {
				z.s.fehler = waehle("CloseAll", errC)
				z.odata.schliessFehler, z.adm.schliessFehler = waehle("OData", errO), waehle("Admin", errA)
				z.odata.nachStop = func() { time.Sleep(5 * time.Millisecond) }
				z.adm.nachStop = z.odata.nachStop
			}, waehle("Ablage", errD))
			err := z.h.Herunterfahren(context.Background())
			folge := []string{"odata:anhalten", "admin:anhalten", pCloseAll, "odata:schliessen", "admin:schliessen", pAblageZu}
			fehler := []error{waehle("CloseAll", errC), waehle("OData", errO), waehle("Admin", errA), waehle("Ablage", errD)}
			if modus != modusRecord {
				folge = slices.Delete(folge, 2, 3)
				fehler[0] = nil
			}
			var got []error
			if j, ok := err.(interface{ Unwrap() []error }); ok {
				got = j.Unwrap()
			}
			if soll := slices.DeleteFunc(fehler, func(e error) bool { return e == nil }); !slices.Equal(got, soll) {
				t.Errorf("%s %s: Fehler %v, erwartet %v", modus, quelle, got, soll)
			}
			ev := z.p.liste()
			ohne := slices.DeleteFunc(slices.Clone(ev), func(e string) bool { return strings.HasSuffix(e, ":serve") || strings.HasSuffix(e, ":ende") })
			ende := slices.IndexFunc(ev, func(e string) bool { return e == pAblageZu })
			if !slices.Equal(ohne, folge) || slices.Index(ev, "odata:ende") > ende || slices.Index(ev, "admin:ende") > ende ||
				!slices.Contains(ev, "odata:ende") || !slices.Contains(ev, "admin:ende") {
				t.Errorf("%s %s: %v, erwartet %v mit beiden Serve vor der Ablage", modus, quelle, ev, folge)
			}
		}
	}
	if zellen != 24 {
		t.Errorf("%d Zellen, erwartet 24", zellen)
	}
}

// TestAdminFristRest (B10): Dauer von Schliessen OData × Dauer von CloseAll (record); Schliessen des Admin-Servers erhält den
// Rest der Frist von 10 s.
func TestAdminFristRest(t *testing.T) {
	const s = time.Second
	dauern := []time.Duration{0, 3 * s, 10 * s, 12 * s}
	zellen := 0
	pruefe := func(modus string, closeAll, odata, soll time.Duration) {
		zellen++
		z := neuZwei(modus, 10*s, func(z *zweiLiteral) {
			z.s.beiCloseAll = func() { z.uhr.Advance(closeAll) }
			z.odata.beiSchliessen = func() { z.uhr.Advance(odata) }
		}, nil)
		if err := z.h.Herunterfahren(context.Background()); err != nil {
			t.Fatal(err)
		}
		if got := z.adm.frist(); !slices.Equal(got, []time.Duration{soll}) {
			t.Errorf("%s CloseAll %s, OData %s: Admin %v, erwartet %s", modus, closeAll, odata, got, soll)
		}
	}
	for i, soll := range [][]time.Duration{{10 * s, 7 * s, 0, 0}, {7 * s, 4 * s, 0, 0}, {0, 0, 0, 0}} {
		for j, o := range dauern {
			pruefe(modusRecord, []time.Duration{0, 3 * s, 10 * s}[i], o, soll[j])
		}
	}
	for _, modus := range []string{modusReplay, modusMock, modusHybrid} {
		for j, o := range dauern {
			pruefe(modus, 0, o, []time.Duration{10 * s, 7 * s, 0, 0}[j])
		}
	}
	if zellen != 24 {
		t.Errorf("%d Zellen, erwartet 24", zellen)
	}
}

// haltendeSessions hält CloseAll an, bis der Test frei schließt.
type haltendeSessions struct {
	inbound.SessionUseCase
	drin, frei chan struct{}
}

func (s haltendeSessions) CloseAll(ctx context.Context) (inbound.CloseBericht, error) {
	close(s.drin)
	<-s.frei
	return s.SessionUseCase.CloseAll(ctx)
}

// sende stellt eine Anfrage an den Admin-Listener mit dem Header Authorization auth (leer: keiner) und liefert Status, Body und
// Header; bei einem Fehler der Leitung Status 0.
func sende(c *http.Client, methode, url, auth, body string) (int, string, http.Header) {
	req, err := http.NewRequest(methode, url, strings.NewReader(body))
	if err != nil {
		return 0, err.Error(), nil
	}
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	resp, err := c.Do(req)
	if err != nil {
		return 0, err.Error(), nil
	}
	defer func() { _ = resp.Body.Close() }()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(b), resp.Header
}

func neuerAdminClient() *http.Client {
	return &http.Client{Transport: &http.Transport{}, Timeout: 10 * time.Second}
}

func bearer(t *testing.T, pfad string) string {
	t.Helper()
	b, err := os.ReadFile(pfad)
	if err != nil {
		t.Fatal(err)
	}
	return "Bearer " + strings.TrimSpace(string(b))
}

// TestAdminZulassung (B11): 4 Modi × {zugelassen, später}: während CloseAll (record) bzw. Schliessen des OData-Servers
// (sonst) läuft, ist ein späterer Request 503 und legt keine Session an; der vor Anhalten zugelassene endet mit 200.
func TestAdminZulassung(t *testing.T) {
	zellen := 0
	for _, modus := range vierModi() {
		zellen += 2
		k := listenerKonfig(t.TempDir(), modus)
		drin, frei := make(chan struct{}), make(chan struct{})
		angenommen, loslassen := make(chan struct{}), make(chan struct{})
		h, db := adminBau(t, k, nil, func(b *bootstrap.Bauteile) {
			p := &protokoll{}
			odata := neuProbe("odata", p)
			mitProben(b, odata, nil)
			neu := bootstrap.StandardBauteile()
			b.NeuerAdmin = func(c admin.Konfig) (bootstrap.Server, error) {
				c.Routen = append(slices.Clone(c.Routen), admin.Route{Methode: "GET", Muster: "/__test/haengt",
					Handler: func(context.Context, admin.Anfrage) (admin.Antwort, error) {
						close(angenommen)
						<-loslassen
						return admin.Antwort{Status: 200}, nil
					}})
				return neu.NeuerAdmin(c)
			}
			if modus != modusRecord {
				odata.beiSchliessen = func() { close(drin); <-frei }
				return
			}
			b.NeueSession = func(c session.Konfig) (inbound.SessionUseCase, error) {
				s, err := neu.NeueSession(c)
				return haltendeSessions{s, drin, frei}, err
			}
		})
		freigeben, halten := sync.OnceFunc(func() { close(loslassen) }), sync.OnceFunc(func() { close(frei) })
		t.Cleanup(halten) // vor dem Cleanup von adminBau, damit ein Abbruch kein Herunterfahren hängen lässt
		t.Cleanup(freigeben)
		if err := h.Hochfahren(context.Background()); err != nil {
			t.Fatalf("%s: %v", modus, err)
		}
		auth, url := bearer(t, k.AdminTokenFile), "http://"+h.AdminAdresse()
		spaeter := []*http.Client{neuerAdminClient(), neuerAdminClient()}
		for _, c := range spaeter {
			if status, body, _ := sende(c, "GET", url+"/healthz", "", ""); status != 200 {
				t.Fatalf("%s: healthz %d %s", modus, status, body)
			}
		}
		zugelassen := make(chan int, 1)
		go func() {
			status, _, _ := sende(neuerAdminClient(), "GET", url+"/__test/haengt", auth, "")
			zugelassen <- status
		}()
		<-angenommen
		ergebnis := make(chan error, 1)
		go func() { ergebnis <- h.Herunterfahren(context.Background()) }()
		select {
		case <-drin:
		case <-time.After(10 * time.Second):
			t.Fatalf("%s: Herunterfahren erreicht den Halt nicht", modus)
		}
		davor, err := db.Sessions().List(context.Background(), "x1")
		if err != nil {
			t.Fatal(err)
		}
		post, postBody, _ := sende(spaeter[0], "POST", url+"/__admin/sessions", auth, `{"service_id":"x1"}`)
		health, healthBody, _ := sende(spaeter[1], "GET", url+"/healthz", "", "")
		danach, err := db.Sessions().List(context.Background(), "x1")
		if err != nil {
			t.Fatal(err)
		}
		if post != 503 || !strings.Contains(postBody, "ODR-E5010") || health != 503 || len(danach) != len(davor) {
			t.Errorf("%s später: POST %d %s, healthz %d %s, Sessions %d nach %d", modus, post, postBody, health, healthBody, len(danach), len(davor))
		}
		freigeben()
		if status := <-zugelassen; status != 200 {
			t.Errorf("%s zugelassen: %d, erwartet 200", modus, status)
		}
		halten()
		if err := <-ergebnis; err != nil {
			t.Errorf("%s: Herunterfahren %v", modus, err)
		}
	}
	if zellen != 8 {
		t.Errorf("%d Zellen, erwartet 8", zellen)
	}
}

// TestAdminFehlerKanal (B12): 7 Ausgänge × 4 Modi, je 100 Wiederholungen: höchstens ein Wert, der Kanal offen, solange ein
// Serve läuft, und geschlossen, wenn beide zurück sind.
func TestAdminFehlerKanal(t *testing.T) {
	errZwei := errors.New("zweiter")
	faelle := []struct {
		name         string
		odata, adm   []error // was Serve des Servers liefern soll; leer: er bedient bis zum Anhalten
		werte        int
		istSentinell bool
	}{
		{"Admin fällt aus", nil, []error{errPhase}, 1, false},
		{"OData fällt aus", []error{errPhase}, nil, 1, false},
		{"beide zugleich", []error{errPhase}, []error{errZwei}, 1, false},
		{"Fehler und nil", []error{errPhase}, []error{nil}, 1, false},
		{"beide nil nach Anhalten", nil, nil, 0, false},
		{"Admin nil vor Herunterfahren", nil, []error{nil}, 1, true},
		{"OData nil vor Herunterfahren", []error{nil}, nil, 1, true},
	}
	zellen := 0
	for _, modus := range vierModi() {
		for _, f := range faelle {
			zellen++
			t.Run(modus+"/"+f.name, func(t *testing.T) {
				for range 100 {
					fehlerKanalLauf(t, modus, f.odata, f.adm, f.werte, f.istSentinell)
				}
			})
		}
	}
	if zellen != 28 {
		t.Errorf("%d Zellen, erwartet 28", zellen)
	}
}

// fehlerKanalLauf fährt einen Handle mit den Ausgängen odata und adm (leer: der Server bedient bis zum Anhalten) und prüft
// den Kanal: werte Werte, bei sentinell errServeBeendet ohne Code.
func fehlerKanalLauf(t *testing.T, modus string, odata, adm []error, werte int, sentinell bool) {
	t.Helper()
	z := neuZwei(modus, time.Second, nil, nil)
	for i, srv := range []*probe{z.odata, z.adm} {
		for _, e := range [][]error{odata, adm}[i] {
			srv.serve <- e
		}
	}
	var got []error
	if werte > 0 {
		select {
		case e := <-z.h.Fehler():
			got = append(got, e)
		case <-time.After(5 * time.Second):
			t.Fatal("kein Wert")
		}
		if odata == nil || adm == nil {
			select {
			case e, offen := <-z.h.Fehler():
				t.Fatalf("%v (offen %v), solange ein Serve läuft", e, offen)
			default:
			}
		}
	}
	fertig := make(chan error, 1)
	go func() { fertig <- z.h.Herunterfahren(context.Background()) }()
	select {
	case <-fertig:
	case <-time.After(5 * time.Second):
		t.Fatal("Herunterfahren hängt")
	}
	for offen := true; offen; {
		var e error
		select {
		case e, offen = <-z.h.Fehler():
			if offen {
				got = append(got, e)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("der Kanal wird nicht geschlossen")
		}
	}
	var code *inbound.Fehler
	if len(got) != werte || sentinell && (!errors.Is(got[0], bootstrap.ErrServeBeendet) || errors.As(got[0], &code)) {
		t.Fatalf("Werte %v, erwartet %d", got, werte)
	}
}
