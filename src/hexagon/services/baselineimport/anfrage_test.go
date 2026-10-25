package baselineimport_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/baselineimport"
)

// fehlerMit prüft, dass err ein *inbound.Fehler mit Code und Status ist, und liefert ihn.
func fehlerMit(t *testing.T, name string, err error, code model.Code, status int) *inbound.Fehler {
	t.Helper()
	var f *inbound.Fehler
	if !errors.As(err, &f) {
		t.Errorf("%s: %v ist kein *Fehler", name, err)
		return nil
	}
	if f.Code != code || f.Status != status {
		t.Errorf("%s: %v/%d, erwartet %v/%d", name, f.Code, f.Status, code, status)
	}
	return f
}

// erwarteFehler ist fehlerMit ohne Rückgabe.
func erwarteFehler(t *testing.T, name string, err error, code model.Code, status int) {
	t.Helper()
	_ = fehlerMit(t, name, err, code, status)
}

func hexNullen(n int) string { return strings.Repeat("0", n) }

// TestZeileA2AnfrageUngueltig: jeder Verstoß der Anfrage (Menge 29) ist ErrUngueltig ohne Code und
// ohne Aufruf eines Ports, der Uhr oder des Generators; die Gutfälle der Grenzen sind keiner.
func TestZeileA2AnfrageUngueltig(t *testing.T) {
	gueltig := "sha256:" + hexNullen(64)
	type fall struct {
		name string
		a    func(a *inbound.ImportAnfrage)
	}
	var faelle []fall
	id := func(name, wert string) {
		faelle = append(faelle, fall{"ImportID " + name, func(a *inbound.ImportAnfrage) { a.ImportID = wert }})
	}
	for name, wert := range map[string]string{
		"leer": "", "65 Bytes": strings.Repeat("a", 65), "A": "A", "_": "_", ".": ".", "/": "/", "Leerzeichen": " ",
		"NUL": "\x00", "Byte 0x80": "\x80", "ü": "\xc3\xbc",
	} {
		id(name, wert)
	}
	faelle = append(faelle, fall{"ServiceID leer", func(a *inbound.ImportAnfrage) { a.ServiceID = "" }})
	for name, wert := range map[string][]string{
		"nil": nil, "leer": {}, "leerer Eintrag": {""}, "zwei verschiedene": {"a", "b"}, "zwei gleiche": {"a", "a"},
	} {
		faelle = append(faelle, fall{"SourceSessionIDs " + name, func(a *inbound.ImportAnfrage) { a.SourceSessionIDs = wert }})
	}
	for name, wert := range map[string]string{
		"x": "x", "Großbuchstaben": "sha256:" + strings.Repeat("A", 64), "63 Ziffern": "sha256:" + hexNullen(63),
		"65 Ziffern": "sha256:" + hexNullen(65), "führendes Leerzeichen": " " + gueltig,
	} {
		faelle = append(faelle, fall{"MetadataHash " + name, func(a *inbound.ImportAnfrage) { a.MetadataHash = wert }})
	}
	for _, wert := range []string{"Reject", "accept", "merge", "replace"} {
		faelle = append(faelle, fall{"ConflictPolicy " + wert, func(a *inbound.ImportAnfrage) { a.ConflictPolicy = wert }})
	}
	for _, wert := range []string{"New_Baseline", "existing", "baseline"} {
		faelle = append(faelle, fall{"Target " + wert, func(a *inbound.ImportAnfrage) { a.Target = wert }})
	}
	if len(faelle)+1 != 29 {
		t.Fatalf("Menge %d, erwartet 29", len(faelle)+1)
	}
	pruefe := func(name string, ctx context.Context, a inbound.ImportAnfrage) {
		w := neu(t).start()
		_, err := w.imp.Import(ctx, a)
		if !errors.Is(err, model.ErrUngueltig) || !ohneCode(err) {
			t.Errorf("%s: %v", name, err)
		}
		if n := w.gesamt(); n != 0 {
			t.Errorf("%s: %d Aufrufe von Ports, Uhr oder Generator", name, n)
		}
	}
	var keinCtx context.Context
	pruefe("ctx nil", keinCtx, anfrage())
	for _, f := range faelle {
		a := anfrage()
		f.a(&a)
		pruefe(f.name, context.Background(), a)
	}
	// Gutfälle: kein Verstoß der Anfrage; der Aufruf endet später, weil die Session kein $metadata trägt.
	var gut []inbound.ImportAnfrage
	for _, v := range []string{"a", strings.Repeat("a", 64), "a-1", "0", "-"} {
		a := anfrage()
		a.ImportID = v
		gut = append(gut, a)
	}
	for _, v := range []string{"", "reject"} {
		a := anfrage()
		a.ConflictPolicy = v
		gut = append(gut, a)
	}
	for _, v := range []string{"", "new_baseline"} {
		a := anfrage()
		a.Target = v
		gut = append(gut, a)
	}
	for _, v := range []string{"", gueltig} {
		a := anfrage()
		a.MetadataHash = v
		gut = append(gut, a)
	}
	for i, a := range gut {
		_, err := neu(t).importiere(a)
		if !errors.Is(err, baselineimport.ErrKeineMetadaten) {
			t.Errorf("Gutfall %d: %v", i, err)
		}
	}
}

// TestZeileA3DienstUnbekannt: ein Service außerhalb von Dienste ist ODR-E5005 ohne Port; der
// Verstoß der Anfrage geht vor.
func TestZeileA3DienstUnbekannt(t *testing.T) {
	w := neu(t)
	a := anfrage()
	a.ServiceID = "fremd"
	_, err := w.importiere(a)
	f := fehlerMit(t, "Dienst unbekannt", err, model.CodeE5005, 404)
	if f != nil && !errors.Is(f, outbound.ErrDienstUnbekannt) {
		t.Errorf("Ursache %v", f.Ursache)
	}
	if n := w.gesamt(); n != 0 {
		t.Errorf("%d Aufrufe", n)
	}
	a.ImportID = ""
	if _, err := w.importiere(a); !errors.Is(err, model.ErrUngueltig) || !ohneCode(err) {
		t.Errorf("Anfrage und Dienst ungültig: %v", err)
	}
}

// TestZeileB1AnfrageHash: der Hash bildet die Kanonform der sechs Mitglieder nach den Standardwerten.
func TestZeileB1AnfrageHash(t *testing.T) {
	a := inbound.ImportAnfrage{ImportID: "imp-1", ServiceID: "catalog", SourceSessionIDs: []string{"ses-00000000000000000000000000000001"}}
	const gold1 = "sha256:e13358b428ac05c8a6027520d61b5d5341551c708cfa0ac5f7dac617f42ee89c"
	const gold2 = "sha256:3c7cc48bfbce5a30cd3a5d951ad005edb9abd00e7d214efd59a244d862f5cd0c"
	if got := baselineimport.AnfrageHash(a); got != gold1 {
		t.Errorf("Beispiel 1: %s", got)
	}
	b := a
	b.MetadataHash = "sha256:" + hexNullen(63) + "1"
	if got := baselineimport.AnfrageHash(b); got != gold2 {
		t.Errorf("Beispiel 2: %s", got)
	}
	aenderungen := map[string]func(*inbound.ImportAnfrage){
		"ImportID":     func(x *inbound.ImportAnfrage) { x.ImportID = "imp-2" },
		"ServiceID":    func(x *inbound.ImportAnfrage) { x.ServiceID = "other" },
		"Session":      func(x *inbound.ImportAnfrage) { x.SourceSessionIDs = []string{"ses-2"} },
		"MetadataHash": func(x *inbound.ImportAnfrage) { x.MetadataHash = "sha256:" + hexNullen(64) },
	}
	for name, f := range aenderungen {
		x := a
		f(&x)
		if baselineimport.AnfrageHash(x) == gold1 {
			t.Errorf("%s ändert den Hash nicht", name)
		}
	}
	x := a
	x.ConflictPolicy, x.Target = "reject", "new_baseline"
	if baselineimport.AnfrageHash(x) != gold1 {
		t.Error("Standardwerte nicht normalisiert")
	}
	zweit := a
	if baselineimport.AnfrageHash(a) != baselineimport.AnfrageHash(zweit) {
		t.Error("derselbe Aufruf, anderer Hash")
	}
}

// sitzungenC liefert die Sessions des Falls C1: Kennung, Service, Status.
func TestZeileC1Session(t *testing.T) {
	type fall struct {
		name    string
		service string
		status  model.SessionStatus
		fehlt   bool
		code    model.Code
		status2 int
		weiter  bool
	}
	faelle := []fall{
		{"unbekannt", dienstID, model.SessionComplete, true, model.CodeE7003, 404, false},
		{"fremd", "other", model.SessionComplete, false, model.CodeE5004, 409, false},
		{"open", dienstID, model.SessionOpen, false, model.CodeE7009, 409, false},
		{"fremd und open", "other", model.SessionOpen, false, model.CodeE5004, 409, false},
		{"complete", dienstID, model.SessionComplete, false, 0, 0, true},
		{"incomplete", dienstID, model.SessionIncomplete, false, 0, 0, true},
		{"weird", dienstID, "weird", false, 0, 0, true},
	}
	for _, f := range faelle {
		w := neu(t)
		if f.fehlt {
			w.sessions.loesche(sitzungID)
		} else {
			w.sessions.m[sitzungID] = model.Session{ID: sitzungID, ServiceID: f.service, CreatedAt: t0(), Status: f.status}
		}
		_, err := w.importiere(anfrage())
		if f.weiter {
			if !errors.Is(err, baselineimport.ErrKeineMetadaten) || w.p.anzahl("Exchanges.ListBySession") != 1 {
				t.Errorf("%s: %v, ListBySession %d", f.name, err, w.p.anzahl("Exchanges.ListBySession"))
			}
		} else {
			fe := fehlerMit(t, f.name, err, f.code, f.status2)
			if f.code == model.CodeE7009 && fe != nil && fe.Ursache != nil {
				t.Errorf("%s: Ursache %v", f.name, fe.Ursache)
			}
			if n := w.p.anzahl("Exchanges.ListBySession"); n != 0 {
				t.Errorf("%s: ListBySession %d nach dem Verstoß", f.name, n)
			}
		}
		if a := w.p.argsVon("Sessions.Get"); len(a) != 1 || a[0][0] != sitzungID {
			t.Errorf("%s: Sessions.Get %v", f.name, a)
		}
	}
}

// schreibweisen sind zehn Schreibweisen einer Kennung, die ein Dienst weder kürzen noch umschreiben darf
// (SPEC-127 Nr. 3): Klein- und Großschreibung, Leerraum am Anfang und am Ende, ein Tabulator, ein Zeichen
// mit Akzent, das große I mit Punkt (U+0130), das Kelvin-Zeichen (U+212A) und das lange s (U+017F). Die
// Zeichen stehen als Bytes (AGENTS.md §3.18).
func schreibweisen() []string {
	return []string{"a", "A", "aB", " a", "a ", "\ta", "\xc3\xa9", "\xc4\xb0", "\xe2\x84\xaa", "\xc5\xbf"}
}

// TestZeileC1Schreibweisen: die Kennung der Session geht unverändert an Sessions.Get und ListBySession, die
// ServiceID unverändert an Dienste und SetMetadata (SPEC-127 Nr. 3 c, 6): zehn Schreibweisen mal die zwei Orte (20).
func TestZeileC1Schreibweisen(t *testing.T) {
	n := 0
	for _, s := range schreibweisen() {
		for _, ort := range []string{"Session", "Service"} {
			n++
			w := neu(t)
			ses, svc := sitzungID, dienstID
			if ort == "Session" {
				ses = s
			} else {
				svc = s
				w.konfig.Dienste = []baselineimport.Dienst{{ID: s, UpstreamURL: upstream}}
			}
			w.sessions.loesche(sitzungID)
			w.sessions.m[ses] = model.Session{ID: ses, ServiceID: svc, CreatedAt: t0(), Status: model.SessionComplete}
			(&bauer{w: w, ses: ses}).metadaten("<edmx 1/>")
			if _, err := w.importiere(inbound.ImportAnfrage{ImportID: "imp-1", ServiceID: svc, SourceSessionIDs: []string{ses}}); err != nil {
				t.Errorf("%s %q: %v", ort, s, err)
				continue
			}
			if a := w.p.argsVon("Sessions.Get"); len(a) != 1 || a[0][0] != ses {
				t.Errorf("%s %q: Sessions.Get %v", ort, s, a)
			}
			if a := w.p.argsVon("Exchanges.ListBySession"); len(a) != 1 || a[0][0] != ses {
				t.Errorf("%s %q: ListBySession %v", ort, s, a)
			}
			if a := w.p.argsVon("Metadaten.SetMetadata"); len(a) != 1 || a[0][0] != svc {
				t.Errorf("%s %q: SetMetadata %v", ort, s, a)
			}
		}
	}
	if n != 20 {
		t.Errorf("Menge %d, erwartet 20", n)
	}
	// Ergänzung außerhalb der Menge: zwei verschiedene Schreibweisen sind zwei Kennungen (90 geordnete Paare je Ort);
	// eine Kennung, die der Dienst kürzt, in der Schreibweise ändert oder vergleicht, ohne sie zu unterscheiden, fällt auf.
	for _, s1 := range schreibweisen() {
		for _, s2 := range schreibweisen() {
			if s1 == s2 {
				continue
			}
			w := neu(t)
			w.konfig.Dienste = []baselineimport.Dienst{{ID: s1, UpstreamURL: upstream}}
			w.sessions.m[sitzungID] = model.Session{ID: sitzungID, ServiceID: s1, CreatedAt: t0(), Status: model.SessionComplete}
			_, err := w.importiere(inbound.ImportAnfrage{ImportID: "imp-1", ServiceID: s2, SourceSessionIDs: []string{sitzungID}})
			erwarteFehler(t, fmt.Sprintf("Service %q gegen %q", s2, s1), err, model.CodeE5005, 404)
			w = neu(t)
			w.sessions.m[s1] = model.Session{ID: s1, ServiceID: dienstID, CreatedAt: t0(), Status: model.SessionComplete}
			_, err = w.importiere(inbound.ImportAnfrage{ImportID: "imp-1", ServiceID: dienstID, SourceSessionIDs: []string{s2}})
			erwarteFehler(t, fmt.Sprintf("Session %q gegen %q", s2, s1), err, model.CodeE7003, 404)
		}
	}
}

// TestZeileI9GetSchreibweisen: die Kennung von Get geht unverändert an Imports.Get (zehn Schreibweisen).
func TestZeileI9GetSchreibweisen(t *testing.T) {
	for _, id := range schreibweisen() {
		w := neu(t).start()
		_, err := w.imp.Get(context.Background(), id)
		erwarteFehler(t, "Get "+id, err, model.CodeE7005, 404)
		if a := w.p.argsVon("Imports.Get"); len(a) != 1 || a[0][0] != id {
			t.Errorf("Get %q: Argumente %v", id, a)
		}
	}
	if len(schreibweisen()) != 10 {
		t.Errorf("Menge %d, erwartet 10", len(schreibweisen()))
	}
}
