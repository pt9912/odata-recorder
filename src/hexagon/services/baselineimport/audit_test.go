package baselineimport_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/services/baselineimport"
)

const (
	// zeichenImport sind die 37 Zeichen einer import_id (SPEC-127 Nr. 3 a).
	zeichenImport = "0123456789abcdefghijklmnopqrstuvwxyz-"
	// zeichenDienst sind die 65 Zeichen einer Kennung des Audits (SPEC-234 Nr. 3).
	zeichenDienst = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._-"
	geheimnis     = "GEHEIM-7f3a"
	zeitgrenze    = 5 * time.Second
)

// fehlgeschlagen ist der Eintrag, den Import für den Versuch imp am Service svc mit dem Code c schreibt.
func fehlgeschlagen(svc, imp string, c model.Code) support.Eintrag {
	return support.Eintrag{Art: support.ArtImportFehlgeschlagen, ServiceID: svc, ImportID: imp, ErrorCode: c}
}

// neueStellen sind die Stellen des Falls „neu“ nach der Prüfung von Imports.Get (SPEC-127 Nr. 4 d bis l):
// zwölf, ohne Imports.Get selbst.
func neueStellen(t *testing.T) []stelle {
	t.Helper()
	var out []stelle
	for _, s := range stellen() {
		if s.weg == "neu" && s.name != "Imports.Get" {
			out = append(out, s)
		}
	}
	if len(out) != 12 {
		t.Fatalf("%d neue Stellen, erwartet 12", len(out))
	}
	return out
}

// neuLauf führt den Weg „neu“ aus, lässt den Aufruf der Stelle mit fehler scheitern und ruft Import mit ctx.
func (s stelle) neuLauf(ctx context.Context, t *testing.T, fehler error, vorStart func(*welt)) (*welt, error) {
	t.Helper()
	w := neu(t)
	w.sessionR()
	if vorStart != nil {
		vorStart(w)
	}
	w.start()
	w.p.scheitere(s.op, s.n, fehler)
	_, err := w.imp.Import(ctx, anfrage())
	return w, err
}

// offeneSession lässt den Import an Schritt (d) mit ODR-E7009 scheitern.
func offeneSession(svc string) func(*welt) {
	return func(w *welt) {
		w.konfig.Dienste = []baselineimport.Dienst{{ID: svc, UpstreamURL: upstream}}
		w.sessions.m[sitzungID] = model.Session{ID: sitzungID, ServiceID: svc, CreatedAt: t0(), Status: model.SessionOpen}
	}
}

// TestAuditKennungen (E05, E01 der Stelle, M5): ein fehlgeschlagener Import schreibt genau einen Eintrag
// mit Service, import_id und Code der Anfrage und leerer KorrelationsID; die Quell-Session steht nicht
// darin. Menge: 37 Zeichen × 2 Längen für die import_id, 65 × 2 für den Service.
func TestAuditKennungen(t *testing.T) {
	n := 0
	for _, c := range zeichenImport {
		for _, l := range []int{1, 64} {
			id := strings.Repeat(string(c), l)
			w := neu(t)
			w.sessionR()
			offeneSession(dienstID)(w)
			a := anfrage()
			a.ImportID = id
			_, err := w.importiere(a)
			if err == nil {
				t.Fatalf("import_id %q: kein Fehler", id)
			}
			if got := w.audit.Eintraege(); !slices.Equal(got, []support.Eintrag{fehlgeschlagen(dienstID, id, model.CodeE7009)}) {
				t.Errorf("import_id %q: %+v", id, got)
			}
			n++
		}
	}
	if n != 74 {
		t.Errorf("import_id %d, erwartet 74", n)
	}
	n = 0
	for _, c := range zeichenDienst {
		for _, l := range []int{1, 64} {
			svc := strings.Repeat(string(c), l)
			w := neu(t)
			w.sessionR()
			offeneSession(svc)(w)
			a := anfrage()
			a.ServiceID = svc
			if _, err := w.importiere(a); err == nil {
				t.Fatalf("Service %q: kein Fehler", svc)
			}
			if got := w.audit.Eintraege(); !slices.Equal(got, []support.Eintrag{fehlgeschlagen(svc, "imp-1", model.CodeE7009)}) {
				t.Errorf("Service %q: %+v", svc, got)
			}
			n++
		}
	}
	if n != 130 {
		t.Errorf("Services %d, erwartet 130", n)
	}
}

// TestAuditCodeJeStelle (E14, M3): ErrorCode ist der Code des Fehlers jedes Schrittes (d) bis (l); Menge:
// zwölf Stellen × zehn Fehler der Ports = 120, Orakel erwarteterCode, Zählprüfung.
func TestAuditCodeJeStelle(t *testing.T) {
	n := 0
	for _, s := range neueStellen(t) {
		for _, f := range portFehler() {
			n++
			g := s.gegenstand
			if s.name == "Publish" && f.name == "ErrNichtGefunden" {
				g = "Session"
			}
			w, err := s.neuLauf(context.Background(), t, f.err, nil)
			code, _ := erwarteterCode(g, f.name)
			var fe *inbound.Fehler
			if !errors.As(err, &fe) || fe.Code != code {
				t.Errorf("%s mit %s: Fehler %v, erwartet %v", s.name, f.name, err, code)
				continue
			}
			if got := w.audit.Eintraege(); !slices.Equal(got, []support.Eintrag{fehlgeschlagen(dienstID, "imp-1", code)}) {
				t.Errorf("%s mit %s: Einträge %+v, erwartet Code %v", s.name, f.name, got, code)
			}
		}
	}
	if n != 120 {
		t.Errorf("Menge %d, erwartet 120", n)
	}
}

// TestAuditCodeOhneCode (E14, M3): codeVon liefert den Code des *inbound.Fehler, nackt und mit %w verpackt,
// sonst ODR-E9001; sieben Codes × zwei Verpackungen und ein Fehler ohne Code × zwei Verpackungen = 16.
func TestAuditCodeOhneCode(t *testing.T) {
	codes := []model.Code{model.CodeE5004, model.CodeE6004, model.CodeE6101, model.CodeE7003, model.CodeE7009, model.CodeE3001, model.CodeE9001}
	n := 0
	for _, c := range codes {
		fe := &inbound.Fehler{Code: c, Status: 500}
		for name, err := range map[string]error{"nackt": fe, "verpackt": fmt.Errorf("Schritt: %w", fe)} {
			n++
			if got := baselineimport.CodeVon(err); got != c {
				t.Errorf("%v %s: %v", c, name, got)
			}
		}
	}
	ohne := errors.New("ohne Code")
	for name, err := range map[string]error{"nackt": ohne, "verpackt": fmt.Errorf("Schritt: %w", ohne)} {
		n++
		if got := baselineimport.CodeVon(err); got != model.CodeE9001 {
			t.Errorf("ohne Code %s: %v", name, got)
		}
	}
	if n != 16 {
		t.Errorf("Menge %d, erwartet 16", n)
	}
}

// TestAuditKeinEintrag (E09, M2 und M4 des Imports): kein Eintrag bei einem Verstoß der Anfrage (63
// Teilmengen der sechs Verstöße), bei ODR-E5005, ODR-E6102, einem Fehler von Imports.Get, bei der
// Wiederholung mit gleichem Hash auch scheiternd (zwei Stellen × zehn Fehler), bei einem Kontextfehler
// (15 Stellen × 4 Formen) und bei Erfolg.
func TestAuditKeinEintrag(t *testing.T) {
	keiner := func(w *welt, was string) {
		t.Helper()
		if n := w.audit.Aufrufe(); n != 0 {
			t.Errorf("%s: %d Aufrufe des Audits", was, n)
		}
	}
	verstoesse := []func(*inbound.ImportAnfrage){
		func(a *inbound.ImportAnfrage) { a.ImportID = "X" },
		func(a *inbound.ImportAnfrage) { a.ServiceID = "" },
		func(a *inbound.ImportAnfrage) { a.SourceSessionIDs = nil },
		func(a *inbound.ImportAnfrage) { a.MetadataHash = "sha256:zz" },
		func(a *inbound.ImportAnfrage) { a.ConflictPolicy = "overwrite" },
		func(a *inbound.ImportAnfrage) { a.Target = "x" },
	}
	n := 0
	for maske := 1; maske < 1<<len(verstoesse); maske++ {
		w := neu(t)
		w.sessionR()
		a := anfrage()
		for i, v := range verstoesse {
			if maske>>i&1 == 1 {
				v(&a)
			}
		}
		if _, err := w.importiere(a); err == nil {
			t.Errorf("Maske %b: kein Fehler", maske)
		}
		keiner(w, fmt.Sprintf("Anfrage %b", maske))
		n++
	}
	if n != 63 {
		t.Errorf("Verstöße %d, erwartet 63", n)
	}

	w := neu(t)
	w.sessionR()
	a := anfrage()
	a.ServiceID = "unbekannt"
	if _, err := w.importiere(a); err == nil {
		t.Error("ODR-E5005: kein Fehler")
	}
	keiner(w, "ODR-E5005")

	w = neu(t)
	w.sessionR()
	mustImport(t, w, anfrage())
	b := anfrage()
	b.SourceSessionIDs = []string{"andere"}
	if _, err := w.importiere(b); err == nil {
		t.Error("ODR-E6102: kein Fehler")
	}
	keiner(w, "ODR-E6102")

	w = neu(t)
	w.sessionR()
	w.p.scheitere("Imports.Get", 1, errPlatte)
	if _, err := w.importiere(anfrage()); err == nil {
		t.Error("Imports.Get: kein Fehler")
	}
	keiner(w, "Imports.Get")

	n = 0
	for _, s := range stellen() {
		if s.weg != "wdh" {
			continue
		}
		for _, f := range portFehler() {
			w, _, err := s.lauf(t, f.err)
			if err == nil {
				t.Errorf("%s mit %s: kein Fehler", s.name, f.name)
			}
			keiner(w, "Wiederholung "+s.name+" mit "+f.name)
			n++
		}
	}
	if n != 20 {
		t.Errorf("Wiederholungen %d, erwartet 20", n)
	}

	n = 0
	for _, s := range stellen() {
		if s.weg == "get" {
			continue
		}
		for i, k := range kontextFehler() {
			w, _, err := s.lauf(t, k)
			if err == nil {
				t.Errorf("%s mit Form %d: kein Fehler", s.name, i)
			}
			keiner(w, fmt.Sprintf("Kontextfehler %s Form %d", s.name, i))
			n++
		}
	}
	if n != 60 {
		t.Errorf("Kontextfehler %d, erwartet 60", n)
	}

	w = neu(t)
	w.sessionR()
	mustImport(t, w, anfrage())
	keiner(w, "Erfolg")
}

// TestAuditWiederholung (E15): jeder gescheiterte Versuch schreibt einen Eintrag; ein gelungener Import
// danach keinen.
func TestAuditWiederholung(t *testing.T) {
	w := neu(t)
	w.sessionR()
	w.sessions.m[sitzungID] = model.Session{ID: sitzungID, ServiceID: dienstID, CreatedAt: t0(), Status: model.SessionOpen}
	for range 2 {
		if _, err := w.importiere(anfrage()); err == nil {
			t.Fatal("kein Fehler")
		}
	}
	soll := []support.Eintrag{fehlgeschlagen(dienstID, "imp-1", model.CodeE7009), fehlgeschlagen(dienstID, "imp-1", model.CodeE7009)}
	if got := w.audit.Eintraege(); !slices.Equal(got, soll) {
		t.Fatalf("Einträge %+v", got)
	}
	w.sessions.m[sitzungID] = model.Session{ID: sitzungID, ServiceID: dienstID, CreatedAt: t0(), Status: model.SessionComplete}
	mustImport(t, w, anfrage())
	if got := w.audit.Eintraege(); !slices.Equal(got, soll) {
		t.Errorf("nach dem Erfolg: Einträge %+v", got)
	}
}

// TestAuditFehlerAendertNichts (E12, E13, M1 des Imports): ein Fehler von Append und ein im Port nach dem
// Fehler beendeter ctx ändern Code und Status des Fehlers nicht, der Eintrag steht; Menge: zwölf Stellen ×
// Fehler (gesetzt, ungesetzt) × ctx (aktiv, im Port beendet) = 48.
func TestAuditFehlerAendertNichts(t *testing.T) {
	n := 0
	for _, s := range neueStellen(t) {
		var ref error
		for _, beendet := range []bool{false, true} {
			for _, fehler := range []error{nil, errors.New("Ablage " + marker)} {
				n++
				ctx, cancel := context.WithCancel(context.Background())
				gefeuert := false
				vor := func(w *welt) {
					if beendet {
						w.p.setzeHaken(s.op, func(k int) {
							if k == s.n {
								gefeuert = true
								cancel()
							}
						})
					}
					w.audit.Fehler = fehler
				}
				w, err := s.neuLauf(ctx, t, errPlatte, vor)
				cancel()
				if beendet && !gefeuert {
					t.Errorf("%s: der Port beendete den ctx nicht", s.name)
				}
				if ref == nil {
					ref = err
				}
				var a, b *inbound.Fehler
				if !errors.As(err, &a) || !errors.As(ref, &b) || a.Code != b.Code || a.Status != b.Status {
					t.Errorf("%s (beendet %v, Fehler %v): Fehler %v, ohne Fehler %v", s.name, beendet, fehler, err, ref)
				}
				if got := w.audit.Eintraege(); !slices.Equal(got, []support.Eintrag{fehlgeschlagen(dienstID, "imp-1", model.CodeE9001)}) {
					t.Errorf("%s (beendet %v, Fehler %v): Einträge %+v", s.name, beendet, fehler, got)
				}
			}
		}
	}
	if n != 48 {
		t.Errorf("Menge %d, erwartet 48", n)
	}
}

// TestAuditKonfig (E17): ein nil-Audit ist ErrUngueltig ohne Importer; New ruft Append nie.
func TestAuditKonfig(t *testing.T) {
	w := neu(t)
	k := w.konfig
	k.Audit = nil
	if imp, err := baselineimport.New(k); !errors.Is(err, model.ErrUngueltig) || imp != nil || !ohneCode(err) {
		t.Errorf("Audit nil: %v, Dienst %v", err, imp != nil)
	}
	if _, err := baselineimport.New(w.konfig); err != nil {
		t.Fatal(err)
	}
	if w.audit.Aufrufe() != 0 {
		t.Errorf("New rief den Audit-Port %d-mal", w.audit.Aufrufe())
	}
	if f, ok := reflect.TypeOf(baselineimport.Konfig{}).FieldByName("Audit"); !ok || f.Type != reflect.TypeOf((*support.AuditLog)(nil)).Elem() {
		t.Errorf("Feld Audit: %v", f)
	}
}

// TestAuditKeinGeheimnis (E18, M6 des Imports): kein Feld des Eintrags trägt Bytes aus dem Fehlertext eines
// Ports oder aus dem Userinfo der UpstreamURL; zwölf Stellen, je sieben Felder durchsucht.
func TestAuditKeinGeheimnis(t *testing.T) {
	n := 0
	for _, s := range neueStellen(t) {
		vor := func(w *welt) {
			w.konfig.Dienste = []baselineimport.Dienst{{ID: dienstID, UpstreamURL: "https://nutzer:" + geheimnis + "@example.test/svc"}}
		}
		w, err := s.neuLauf(context.Background(), t, errors.New("Port "+geheimnis), vor)
		if err == nil {
			t.Fatalf("%s: kein Fehler", s.name)
		}
		got := w.audit.Eintraege()
		for _, e := range got {
			v := reflect.ValueOf(e)
			if v.NumField() != 7 {
				t.Fatalf("Felder %d", v.NumField())
			}
			for i := range v.NumField() {
				if str := fmt.Sprint(v.Field(i).Interface()); strings.Contains(str, geheimnis) {
					t.Errorf("%s: Feld %s trägt das Geheimnis: %q", s.name, v.Type().Field(i).Name, str)
				}
			}
		}
		if len(got) != 1 {
			t.Errorf("%s: Einträge %+v", s.name, got)
		}
		n++
	}
	if n != 12 {
		t.Errorf("Stellen %d, erwartet 12", n)
	}
}

// TestAuditSynchronOhneFrist (E19, F1, M1 des Imports): Append läuft auf der Goroutine des Dienstes, Import
// setzt erst nach seiner Rückkehr fort, auch bei im Halt beendetem ctx; im Halt ist der Fehler entstanden,
// der Eintrag fehlt, der Dienst tut nichts weiter, und Import ist nach 100 ms nicht zurückgekehrt. Nach
// der Freigabe steht der Eintrag, das Ergebnis ist das des Laufs ohne Halt. Stelle: Commit.
func TestAuditSynchronOhneFrist(t *testing.T) {
	var commit stelle
	for _, s := range neueStellen(t) {
		if s.name == "Commit" {
			commit = s
		}
	}
	type ergebnis struct{ err error }
	n := 0
	for _, beendet := range []bool{false, true} {
		n++
		refCtx, refCancel := context.WithCancel(context.Background())
		ref, rerr := commit.neuLauf(refCtx, t, errPlatte, func(w *welt) {
			w.audit.Vor = func(support.Eintrag) {
				if beendet {
					refCancel()
				}
			}
		})
		refCancel()

		ctx, cancel := context.WithCancel(context.Background())
		erreicht, frei := make(chan struct{}), make(chan struct{})
		var w *welt
		var stand []support.Eintrag
		var folge []string
		fertig := make(chan ergebnis, 1)
		go func() {
			_, err := commit.neuLauf(ctx, t, errPlatte, func(x *welt) {
				w = x
				x.audit.Vor = func(support.Eintrag) {
					if beendet {
						cancel()
					}
					stand, folge = x.audit.Eintraege(), x.p.folge()
					close(erreicht)
					<-frei
				}
			})
			fertig <- ergebnis{err}
		}()
		select {
		case <-erreicht:
		case <-time.After(zeitgrenze):
			t.Fatalf("beendet %v: Append nicht erreicht", beendet)
		}
		select {
		case <-fertig:
			t.Errorf("beendet %v: Import kehrte im Halt von Append zurück", beendet)
		case <-time.After(100 * time.Millisecond):
		}
		if !slices.Equal(folge, ref.p.folge()) || len(stand) != 0 || !slices.Equal(w.p.folge(), folge) {
			t.Errorf("beendet %v: im Halt Folge %v, danach %v, ohne Halt %v, Einträge %+v", beendet, folge, w.p.folge(), ref.p.folge(), stand)
		}
		if len(folge) == 0 || folge[len(folge)-1] != "Abort" {
			t.Errorf("beendet %v: der Fehlerschritt fehlt in %v", beendet, folge)
		}
		close(frei)
		select {
		case r := <-fertig:
			var a, b *inbound.Fehler
			if !errors.As(r.err, &a) || !errors.As(rerr, &b) || a.Code != b.Code {
				t.Errorf("beendet %v: Fehler %v; ohne Halt %v", beendet, r.err, rerr)
			}
		case <-time.After(zeitgrenze):
			t.Fatalf("beendet %v: Import kehrte nach der Freigabe nicht zurück", beendet)
		}
		cancel()
		if got, soll := w.audit.Eintraege(), ref.audit.Eintraege(); !slices.Equal(got, soll) || len(got) != 1 {
			t.Errorf("beendet %v: Einträge %+v, ohne Halt %+v", beendet, got, soll)
		}
	}
	if n != 2 {
		t.Errorf("Menge %d, erwartet 2", n)
	}
}
