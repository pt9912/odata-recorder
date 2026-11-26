package bootstrap_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/adapters/driving/httpodata"
	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// probe ist ein Server aus dem Test: Serve kehrt mit dem Wert aus serve zurück oder nach dem Anhalten mit nil, dann erst nach
// nachStop; die Operationen stehen mit dem Namen im Protokoll.
type probe struct {
	name     string
	p        *protokoll
	serve    chan error
	stop     chan struct{}
	einmal   sync.Once
	nachStop func()
	// schliessFehler liefert Schliessen, beiSchliessen läuft darin vorher.
	schliessFehler error
	beiSchliessen  func()

	mu      sync.Mutex
	fristen []time.Duration
}

func neuProbe(name string, p *protokoll) *probe {
	return &probe{name: name, p: p, serve: make(chan error, 1), stop: make(chan struct{})}
}

func (s *probe) Serve() error {
	s.p.add("%s:serve", s.name)
	defer s.p.add("%s:ende", s.name)
	select {
	case e := <-s.serve:
		return e
	case <-s.stop:
		if s.nachStop != nil {
			s.nachStop()
		}
		return nil
	}
}

func (s *probe) halt() { s.einmal.Do(func() { close(s.stop) }) }

func (s *probe) Anhalten() {
	s.p.add("%s:anhalten", s.name)
	s.halt()
}

func (s *probe) Schliessen(frist time.Duration) error {
	s.p.add("%s:schliessen", s.name)
	s.mu.Lock()
	s.fristen = append(s.fristen, frist)
	s.mu.Unlock()
	if s.beiSchliessen != nil {
		s.beiSchliessen()
	}
	s.halt()
	return s.schliessFehler
}

func (s *probe) frist() []time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.fristen)
}

// mitProben setzt beide Server als Fakes.
func mitProben(b *bootstrap.Bauteile, odata, adm *probe) {
	b.NeuerServer = func(httpodata.Konfig) (bootstrap.Server, error) { return odata, nil }
	b.NeuerAdmin = func(admin.Konfig) (bootstrap.Server, error) { return adm, nil }
}

// TestAdminKonfig (B06): jedes der 7 Felder von admin.Konfig × 4 Modi = 28 Zellen.
func TestAdminKonfig(t *testing.T) {
	if n := len(reflect.VisibleFields(reflect.TypeOf(admin.Konfig{}))); n != 7 {
		t.Fatalf("admin.Konfig hat %d Felder, erwartet 7", n)
	}
	zellen := 0
	for _, modus := range vierModi() {
		k := listenerKonfig(t.TempDir(), modus)
		var kfg admin.Konfig
		h, _ := adminBau(t, k, nil, func(b *bootstrap.Bauteile) {
			neu := b.NeuerAdmin
			b.NeuerAdmin = func(c admin.Konfig) (bootstrap.Server, error) {
				kfg = c
				return neu(c)
			}
		})
		if err := h.Hochfahren(context.Background()); err != nil {
			t.Fatalf("%s: %v", modus, err)
		}
		inhalt, err := os.ReadFile(k.AdminTokenFile)
		routen := bootstrap.AdminRouten(h)
		felder := map[string]bool{
			"Listener":   kfg.Listener != nil && kfg.Listener.Addr().String() == h.AdminAdresse(),
			"Token":      err == nil && string(kfg.Token) == strings.TrimSpace(string(inhalt)),
			"IDs":        kfg.IDs == bootstrap.IDsVon(h),
			"Audit":      kfg.Audit == bootstrap.AuditVon(h),
			"Telemetrie": kfg.Telemetrie == support.Telemetry(bootstrap.TelemetrieVon(h)),
			"Routen":     len(routen) > 0 && len(kfg.Routen) == len(routen) && &kfg.Routen[0] == &routen[0],
			"Vergleich":  kfg.Vergleich == nil,
		}
		for feld, ok := range felder {
			zellen++
			if !ok {
				t.Errorf("%s: Feld %s", modus, feld)
			}
		}
	}
	if zellen != 28 {
		t.Errorf("%d Zellen, erwartet 28", zellen)
	}
}

// TestAdminOhneListener (B07): ein Literal mit leerem AdminListen, 4 Modi × {Build, Hochfahren, Herunterfahren}.
func TestAdminOhneListener(t *testing.T) {
	zellen := 0
	for _, modus := range vierModi() {
		k := listenerKonfig(t.TempDir(), modus)
		k.AdminListen = ""
		h := neuHaken()
		var hd *bootstrap.Handle
		schritte := []func() error{
			func() (err error) { hd, err = bootstrap.BuildMit(context.Background(), k, nil, h.bauteile()); return },
			func() error { return hd.Hochfahren(context.Background()) },
			func() error { return hd.Herunterfahren(context.Background()) },
		}
		for i, schritt := range schritte {
			zellen++
			if err := schritt(); err != nil {
				t.Fatalf("%s Schritt %d: %v", modus, i, err)
			}
			binds := min(i, 1)
			_, err := os.Stat(k.AdminTokenFile)
			if h.p.zaehle(pListen) != binds || hd.AdminAdresse() != "" || !errors.Is(err, os.ErrNotExist) || h.adminKfg.Listener != nil {
				t.Errorf("%s Schritt %d: %d Bind, AdminAdresse %q, Token-Datei %v", modus, i, h.p.zaehle(pListen), hd.AdminAdresse(), err)
			}
		}
		if h.p.zaehle("anhalten") != 1 || h.p.zaehle(pSchliessen) != 1 {
			t.Errorf("%s: Herunterfahren kennt nicht genau einen Server: %v", modus, h.p.liste())
		}
	}
	if zellen != 12 {
		t.Errorf("%d Zellen, erwartet 12", zellen)
	}
}

// zaehlListener zählt Close.
type zaehlListener struct {
	net.Listener
	zu *int
}

func (l zaehlListener) Close() error {
	*l.zu++
	return l.Listener.Close()
}

// bindung ist ein hochfahrender Handle mit echtem Listen, Fakes der Server und der Zahl der Close.
type bindung struct {
	h          *haken
	odata, adm *probe
	b          bootstrap.Bauteile
	zu         int
}

func neueBindung() *bindung {
	x := &bindung{h: neuHaken()}
	p := &protokoll{}
	x.odata, x.adm = neuProbe("odata", p), neuProbe("admin", p)
	x.b = x.h.bauteile()
	echt := bootstrap.StandardBauteile().Listen
	x.b.Listen = func(ctx context.Context, network, addr string) (net.Listener, error) {
		l, err := echt(ctx, network, addr)
		if err != nil {
			return nil, err
		}
		return zaehlListener{l, &x.zu}, nil
	}
	mitProben(&x.b, x.odata, x.adm)
	return x
}

func (x *bindung) serves() int {
	return x.odata.p.zaehle("odata:serve") + x.adm.p.zaehle("admin:serve")
}

// hoch baut und fährt hoch; der Test fährt am Ende herunter.
func (x *bindung) hoch(ctx context.Context, t *testing.T, k bootstrap.Konfiguration) (*bootstrap.Handle, error) {
	t.Helper()
	hd, err := bootstrap.BuildMit(context.WithoutCancel(ctx), k, nil, x.b)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	t.Cleanup(func() { _ = hd.Herunterfahren(context.WithoutCancel(ctx)) })
	return hd, hd.Hochfahren(ctx)
}

// belegt liefert eine Adresse, die während des Tests belegt bleibt, frei eine Adresse, die eben noch frei war.
func belegt(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = l.Close() })
	return l.Addr().String()
}

func frei(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = l.Close() }()
	return l.Addr().String()
}

// TestAdminBindung (B08): M-Bind × 4 Modi = 20, ctx vor dem Bind und im zweiten Bind = 8, ctx endet nach der Bindung = 8,
// Konstruktorfehler = 8.
func TestAdminBindung(t *testing.T) {
	const listen, adm = "ODR-E2001 LISTEN: Adresse nicht bindbar", "ODR-E2001 ADMIN_LISTEN: Adresse nicht bindbar"
	faelle := []struct {
		name   string
		setze  func(t *testing.T, k *bootstrap.Konfiguration)
		zeilen []string
		zu     int
	}{
		{"beide frei", func(*testing.T, *bootstrap.Konfiguration) {}, nil, 0},
		{"LISTEN belegt", func(t *testing.T, k *bootstrap.Konfiguration) { k.Listen = belegt(t) }, []string{listen}, 1},
		{"ADMIN_LISTEN belegt", func(t *testing.T, k *bootstrap.Konfiguration) { k.AdminListen = belegt(t) }, []string{adm}, 1},
		{"beide belegt", func(t *testing.T, k *bootstrap.Konfiguration) { k.Listen, k.AdminListen = belegt(t), belegt(t) }, []string{adm, listen}, 0},
		{"dieselbe Adresse", func(t *testing.T, k *bootstrap.Konfiguration) { a := frei(t); k.Listen, k.AdminListen = a, a }, []string{adm}, 1},
	}
	zellen := 0
	leer := func(t *testing.T, hd *bootstrap.Handle, x *bindung) {
		t.Helper()
		if x.serves() != 0 || hd.Adresse() != "" || hd.AdminAdresse() != "" {
			t.Errorf("%d Serve, Adressen %q %q, erwartet keine", x.serves(), hd.Adresse(), hd.AdminAdresse())
		}
	}
	beide := func(t *testing.T, hd *bootstrap.Handle, x *bindung) {
		t.Helper()
		wartetBis(func() bool { return x.serves() >= 2 })
		_, a, errA := net.SplitHostPort(hd.Adresse())
		_, b, errB := net.SplitHostPort(hd.AdminAdresse())
		if x.serves() != 2 || errA != nil || errB != nil || a == b {
			t.Errorf("%d Serve, Adressen %q %q, erwartet je einen Server mit eigenem Port", x.serves(), hd.Adresse(), hd.AdminAdresse())
		}
	}
	for _, modus := range vierModi() {
		for _, f := range faelle {
			zellen++
			t.Run(modus+"/"+f.name, func(t *testing.T) {
				x, k := neueBindung(), listenerKonfig(t.TempDir(), modus)
				f.setze(t, &k)
				hd, err := x.hoch(context.Background(), t, k)
				var kf *bootstrap.KonfigFehler
				if f.zeilen == nil {
					if err != nil {
						t.Fatal(err)
					}
					beide(t, hd, x)
					return
				}
				if !errors.As(err, &kf) || !slices.Equal(kf.Zeilen, f.zeilen) || x.zu != f.zu {
					t.Errorf("%v mit %d Close, erwartet %q und %d", err, x.zu, f.zeilen, f.zu)
				}
				leer(t, hd, x)
			})
		}
		for _, ort := range []string{"vor dem Bind", "im zweiten Bind"} {
			zellen++
			t.Run(modus+"/ctx "+ort, func(t *testing.T) {
				x := neueBindung()
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				binds, echt := 0, x.b.Listen
				x.b.Listen = func(c context.Context, network, addr string) (net.Listener, error) {
					binds++
					if binds == 2 {
						cancel()
						return nil, fmt.Errorf("bind: %w", c.Err())
					}
					return echt(c, network, addr)
				}
				soll, zu := 2, 1
				if ort == "vor dem Bind" {
					cancel()
					soll, zu = 0, 0
				}
				hd, err := x.hoch(ctx, t, listenerKonfig(t.TempDir(), modus))
				var kf *bootstrap.KonfigFehler
				if !errors.Is(err, context.Canceled) || errors.As(err, &kf) || binds != soll || x.zu != zu {
					t.Errorf("%v, %d Bind, %d Close; erwartet den Kontextfehler, %d und %d", err, binds, x.zu, soll, zu)
				}
				leer(t, hd, x)
			})
		}
		for _, ort := range []string{"beim zweiten Bind", "beim Bau des Admin-Servers"} {
			zellen++
			t.Run(modus+"/ctx endet "+ort, func(t *testing.T) {
				x := neueBindung()
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				binds, echt, neu := 0, x.b.Listen, x.b.NeuerAdmin
				x.b.Listen = func(c context.Context, network, addr string) (net.Listener, error) {
					l, err := echt(c, network, addr)
					if binds++; binds == 2 && ort == "beim zweiten Bind" {
						cancel()
					}
					return l, err
				}
				x.b.NeuerAdmin = func(k admin.Konfig) (bootstrap.Server, error) {
					if ort == "beim Bau des Admin-Servers" {
						cancel()
					}
					return neu(k)
				}
				hd, err := x.hoch(ctx, t, listenerKonfig(t.TempDir(), modus))
				if err != nil {
					t.Fatalf("Hochfahren: %v, erwartet nil", err)
				}
				beide(t, hd, x)
			})
		}
		for _, welcher := range []string{"OData", "Admin"} {
			zellen++
			t.Run(modus+"/Konstruktorfehler "+welcher, func(t *testing.T) {
				x := neueBindung()
				if welcher == "OData" {
					x.b.NeuerServer = func(httpodata.Konfig) (bootstrap.Server, error) { return nil, errPhase }
				} else {
					x.b.NeuerAdmin = func(admin.Konfig) (bootstrap.Server, error) { return nil, errPhase }
				}
				hd, err := x.hoch(context.Background(), t, listenerKonfig(t.TempDir(), modus))
				var kf *bootstrap.KonfigFehler
				if !errors.Is(err, errPhase) || errors.As(err, &kf) || x.zu != 2 {
					t.Errorf("%v mit %d Close, erwartet den Konstruktorfehler und 2", err, x.zu)
				}
				leer(t, hd, x)
				if err := hd.Herunterfahren(context.Background()); err != nil || x.odata.p.zaehle("odata:anhalten")+x.adm.p.zaehle("admin:anhalten") != 0 {
					t.Errorf("Herunterfahren: %v, Protokoll %v", err, x.odata.p.liste())
				}
			})
		}
	}
	if zellen != 44 {
		t.Errorf("%d Zellen, erwartet 44", zellen)
	}
}

// TestAdminAusfall (B13): {Admin, OData} × {Fehler, nil} × 4 Modi; das Schliessen des OData-Servers meldet eine eigene Zeile,
// die vor der Zeile des Ausfalls steht.
func TestAdminAusfall(t *testing.T) {
	zeilen := []string{model.CodeE5009.String() + ": " + model.CodeE5009.Message(), "ODR-E9001: Internal error"}
	zellen := 0
	for _, modus := range vierModi() {
		for _, welcher := range []string{"admin", "odata"} {
			for _, art := range []error{errPhase, nil} {
				zellen++
				t.Run(fmt.Sprintf("%s/%s/%v", modus, welcher, art), func(t *testing.T) {
					p := &protokoll{}
					odata, adm := neuProbe("odata", p), neuProbe("admin", p)
					odata.schliessFehler = &inbound.Fehler{Code: model.CodeE5009, Status: 500, Ursache: errPhase}
					h := neuHaken()
					b := h.bauteile()
					mitProben(&b, odata, adm)
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					puffer := &syncPuffer{}
					ergebnis := make(chan int, 1)
					go func() { ergebnis <- bootstrap.StartMit(ctx, umgebungFuer(t, modus, 1), puffer, b) }()
					wartetBis(func() bool { return strings.Contains(puffer.String(), "start: bereit") })
					map[string]*probe{"admin": adm, "odata": odata}[welcher].serve <- art
					code := <-ergebnis
					if ctx.Err() != nil || code != 1 || !slices.Equal(meldungsZeilen(puffer.String()), zeilen) {
						t.Errorf("Rückgabe %d (%v), Meldungen %q, erwartet 1 und %q", code, ctx.Err(), meldungsZeilen(puffer.String()), zeilen)
					}
					if p.zaehle("admin:anhalten") != 1 || p.zaehle("odata:anhalten") != 1 || p.zaehle("odata:schliessen") != 1 {
						t.Errorf("Protokoll %v", p.liste())
					}
				})
			}
		}
	}
	if zellen != 16 {
		t.Errorf("%d Zellen, erwartet 16", zellen)
	}
}
