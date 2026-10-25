package baselineimport_test

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/baselineimport"
)

// TestZeileF4StromGenauEinmalGeschlossen: jeder geöffnete Strom wird genau einmal geschlossen, vor
// dem nächsten Open (Menge 15: drei Stellen mal fünf Ausgänge).
func TestZeileF4StromGenauEinmalGeschlossen(t *testing.T) {
	stellenOps := []struct{ name, op string }{{"$metadata", "ParseMetadata"}, {"Collection", "ParseCollection"}, {"Entity", "ParseEntity"}}
	ausgaenge := []struct {
		name string
		err  error
	}{
		{"gelungen", nil}, {"Fehler des Lesers", nil}, {"Kontextfehler", context.Canceled},
		{"Fehler von src", outbound.ErrBlobBeschaedigt}, {"ErrUngueltig", model.ErrUngueltig},
	}
	n := 0
	for _, s := range stellenOps {
		for _, a := range ausgaenge {
			n++
			w := neu(t)
			w.sessionR()
			err := a.err
			if a.name == "Fehler des Lesers" {
				err = &model.DatenFehler{Ursache: model.DatenWert}
				if s.op == "ParseMetadata" {
					err = &model.MetadatenFehler{Ursache: model.UrsacheStruktur}
				}
			}
			if err != nil {
				w.p.scheitere(s.op, 1, err)
			}
			_, ierr := w.importiere(anfrage())
			if (ierr == nil) != (err == nil) {
				t.Errorf("%s %s: %v", s.name, a.name, ierr)
			}
			strome := w.blobs.strome()
			if len(strome) == 0 {
				t.Fatalf("%s %s: nichts geöffnet", s.name, a.name)
			}
			for i, st := range strome {
				if st.geschlossen() != 1 {
					t.Errorf("%s %s: Strom %d %d-mal geschlossen", s.name, a.name, i, st.geschlossen())
				}
			}
			offen := false
			for _, e := range w.p.folge() {
				switch e {
				case "Blobs.Open":
					if offen {
						t.Errorf("%s %s: Open vor Close", s.name, a.name)
					}
					offen = true
				case "Close":
					offen = false
				}
			}
			if offen {
				t.Errorf("%s %s: der letzte Strom blieb offen", s.name, a.name)
			}
		}
	}
	if n != 15 {
		t.Errorf("Menge %d", n)
	}
}

func datenUrsachen() []model.DatenUrsache {
	return []model.DatenUrsache{model.DatenSyntax, model.DatenKodierung, model.DatenTiefe, model.DatenZuGross, model.DatenDoppelterName,
		model.DatenStruktur, model.DatenKontext, model.DatenODataTyp, model.DatenUnbekannteProperty, model.DatenTyp, model.DatenWert,
		model.DatenNull, model.DatenZahlZuLang, model.DatenSchluessel, model.DatenEntitySet}
}

func metadatenUrsachen() []model.MetadatenUrsache {
	return []model.MetadatenUrsache{model.UrsacheSyntax, model.UrsacheKodierung, model.UrsacheDoctype, model.UrsacheWurzel, model.UrsacheStruktur,
		model.UrsacheVersion, model.UrsacheOhneSchluessel, model.UrsacheNichtAufloesbar, model.UrsacheSchluesseltyp, model.UrsacheZuGross,
		model.UrsacheTiefe, model.UrsacheKette, model.UrsacheZuVieleMembers}
}

// TestZeileF5UrsacheUndOrt: Ursache und Ort bleiben erreichbar (Menge 43).
func TestZeileF5UrsacheUndOrt(t *testing.T) {
	if len(datenUrsachen()) != 15 || len(metadatenUrsachen()) != 13 {
		t.Fatalf("Ursachen %d und %d", len(datenUrsachen()), len(metadatenUrsachen()))
	}
	n := 0
	for _, op := range []string{"ParseCollection", "ParseEntity"} {
		for _, u := range datenUrsachen() {
			n++
			w := neu(t)
			w.sessionR()
			df := &model.DatenFehler{Ursache: u, Ort: "/value/0/" + string(u)}
			w.p.scheitere(op, 1, df)
			_, err := w.importiere(anfrage())
			erwarteFehler(t, op+" "+string(u), err, model.CodeE6101, 422)
			var got *model.DatenFehler
			if !errors.As(err, &got) || got != df || got.Ursache != u || got.Ort != df.Ort {
				t.Errorf("%s %s: %v", op, u, err)
			}
		}
	}
	for _, u := range metadatenUrsachen() {
		n++
		w := neu(t)
		w.sessionR()
		mf := &model.MetadatenFehler{Ursache: u, Ort: "Shop/" + string(u)}
		w.p.scheitere("ParseMetadata", 1, mf)
		_, err := w.importiere(anfrage())
		erwarteFehler(t, "ParseMetadata "+string(u), err, model.CodeE6004, 422)
		var got *model.MetadatenFehler
		if !errors.As(err, &got) || got != mf {
			t.Errorf("ParseMetadata %s: %v", u, err)
		}
	}
	if n != 43 {
		t.Errorf("Menge %d, erwartet 43", n)
	}
}

// verstoss ist einer der Verstöße des Ablaufs (SPEC-127 Nr. 4) in der Reihenfolge der Spezifikation.
type verstoss struct {
	nr     int
	name   string
	setup  func(w *welt, a *inbound.ImportAnfrage)
	trifft func(err error) bool
}

func istCode(code model.Code) func(error) bool {
	return func(err error) bool {
		var f *inbound.Fehler
		return errors.As(err, &f) && f.Code == code
	}
}

// verstoesse liefert die 16 Verstöße; jeder Aufruf bildet eigene Fehlerwerte.
func verstoesse() []verstoss {
	tag := func(n int) error { return fmt.Errorf("Verstoß %d: %w", n, errPlatte) }
	t7, t11, t14, t15, t16 := tag(7), tag(11), tag(14), tag(15), tag(16)
	mf := &model.MetadatenFehler{Ursache: model.UrsacheStruktur, Ort: "Shop"}
	df := &model.DatenFehler{Ursache: model.DatenWert, Ort: "/value/0"}
	istTag := func(e error) func(error) bool { return func(err error) bool { return errors.Is(err, e) } }
	return []verstoss{
		{1, "Anfrage ungültig", func(w *welt, a *inbound.ImportAnfrage) { a.ImportID = "" },
			func(err error) bool { return errors.Is(err, model.ErrUngueltig) && ohneCode(err) }},
		{2, "Dienst unbekannt", func(w *welt, a *inbound.ImportAnfrage) { a.ServiceID = "fremd" }, istCode(model.CodeE5005)},
		{3, "import_id bekannt, anderer Hash", func(w *welt, a *inbound.ImportAnfrage) {
			w.imports.m["imp-1"] = model.Import{ID: "imp-1", RequestHash: "sha256:" + hexNullen(64), BaselineSnapshotID: "snp-x", ServiceID: dienstID, CreatedAt: t0(), ReportJSON: "{}"}
		}, istCode(model.CodeE6102)},
		{4, "Session unbekannt", func(w *welt, a *inbound.ImportAnfrage) { w.sessions.loesche(sitzungID) }, istCode(model.CodeE7003)},
		{5, "Session eines anderen Service", func(w *welt, a *inbound.ImportAnfrage) {
			w.sessions.m[sitzungID] = model.Session{ID: sitzungID, ServiceID: "other", CreatedAt: t0(), Status: w.sessions.m[sitzungID].Status}
		}, istCode(model.CodeE5004)},
		{6, "Session offen", func(w *welt, a *inbound.ImportAnfrage) {
			s := w.sessions.m[sitzungID]
			s.Status = model.SessionOpen
			w.sessions.m[sitzungID] = s
		}, istCode(model.CodeE7009)},
		{7, "ListBySession scheitert", func(w *welt, a *inbound.ImportAnfrage) { w.p.scheitere("Exchanges.ListBySession", 0, t7) }, istTag(t7)},
		{8, "kein $metadata", func(w *welt, a *inbound.ImportAnfrage) { w.exchanges.m[sitzungID] = w.exchanges.m[sitzungID][1:] },
			istTag(baselineimport.ErrKeineMetadaten)},
		{9, "verschiedene $metadata", func(w *welt, a *inbound.ImportAnfrage) {
			(&bauer{w: w, ses: sitzungID, seq: 10}).metadaten("<edmx 2/>")
		}, istTag(baselineimport.ErrMetadatenAbweichend)},
		{10, "Hash der Anfrage falsch", func(w *welt, a *inbound.ImportAnfrage) { a.MetadataHash = refVon([]byte("anderes")) },
			istTag(baselineimport.ErrMetadatenHash)},
		{11, "Open ($metadata) scheitert", func(w *welt, a *inbound.ImportAnfrage) { w.p.scheitere("Blobs.Open", 1, t11) }, istTag(t11)},
		{12, "Metadatenleser scheitert", func(w *welt, a *inbound.ImportAnfrage) { w.meta.vorgabe.err = mf }, func(err error) bool {
			var g *model.MetadatenFehler
			return errors.As(err, &g) && g == mf
		}},
		{13, "Fehler einer Antwort", func(w *welt, a *inbound.ImportAnfrage) {
			w.daten.nachRef[w.exchanges.m[sitzungID][1].ResponseBlob] = datenVorgabe{err: df}
		}, func(err error) bool {
			var g *model.DatenFehler
			return errors.As(err, &g) && g == df
		}},
		{14, "Schreiben des Manifests scheitert", func(w *welt, a *inbound.ImportAnfrage) { w.p.scheitere("Blobs.Begin", 0, t14) }, istTag(t14)},
		{15, "Publish scheitert", func(w *welt, a *inbound.ImportAnfrage) { w.p.scheitere("Imports.Publish", 0, t15) }, istTag(t15)},
		{16, "SetMetadata scheitert", func(w *welt, a *inbound.ImportAnfrage) { w.p.scheitere("Metadaten.SetMetadata", 0, t16) }, istTag(t16)},
	}
}

// nichtZugleich sind die neun Paare von Verstößen, die kein Aufbau zugleich herstellt (Plan, P2).
func nichtZugleich(i, j int) bool {
	for _, p := range [][2]int{{4, 5}, {4, 6}, {8, 9}, {8, 10}, {8, 11}, {8, 12}, {9, 11}, {9, 12}, {11, 12}} {
		if p == [2]int{i, j} {
			return true
		}
	}
	return false
}

// phase nennt, ab welchem Verstoß der Ablauf die Operation ruft.
func phase(op string) int {
	switch op {
	case "Imports.Get":
		return 3
	case "Sessions.Get":
		return 4
	case "Exchanges.ListBySession":
		return 7
	case "ParseMetadata":
		return 12
	case "ParseCollection", "ParseEntity":
		return 13
	case "IDs.SnapshotID", "Uhr.Now", "Blobs.Begin":
		return 14
	case "Write", "Commit":
		return 15
	case "Imports.Publish":
		return 15
	case "Metadaten.SetMetadata":
		return 16
	case "Abort", "Snapshots.Get":
		return 99
	}
	return 0
}

// closeLauf ist ein Lauf von F4: der Import auf der Session R, dessen n-te Operation op mit fehler scheitert, und alle
// Streams scheitern beim Schließen mit closeFehler.
type closeLauf struct {
	r       inbound.ImportErgebnis
	err     error
	zaehler map[string]int
	w       *welt
}

func laufeMitClose(t *testing.T, op string, fehler, closeFehler error) closeLauf {
	t.Helper()
	w := neu(t)
	w.sessionR()
	if fehler != nil {
		w.p.scheitere(op, 1, fehler)
	}
	if closeFehler != nil {
		w.p.scheitere("Close", 0, closeFehler)
	}
	r, err := w.importiere(anfrage())
	w.p.mu.Lock()
	z := maps.Clone(w.p.zaehler)
	w.p.mu.Unlock()
	return closeLauf{r, err, z, w}
}

// TestZeileF4CloseFehler: ein Fehler von Close eines gelesenen Stroms wird nicht ausgewertet (SPEC-127 Nr. 7 f, 9):
// drei Stellen {$metadata, Collection, Entity} mal fünf Ausgänge {gelungen, Fehler des Lesers, Kontextfehler, Fehler von
// src, ErrUngueltig} mal vier Formen von Close {gelingt, scheitert mit einem fremden Fehler, mit einem Kontextfehler,
// mit einem Fehler der Tabelle (ErrBlobFehlt)} (60): Close gleich Open gleich 1 je Strom und vor dem nächsten Open,
// dieselbe Rückgabe (der Fehler per Wert, das Ergebnis byte-gleich) und dieselben Zähler aller Ports wie mit gelingendem Close.
func TestZeileF4CloseFehler(t *testing.T) {
	stellenOps := []struct{ name, op string }{{"$metadata", "ParseMetadata"}, {"Collection", "ParseCollection"}, {"Entity", "ParseEntity"}}
	ausgaenge := []struct {
		name string
		err  error
	}{
		{"gelungen", nil}, {"Fehler des Lesers", nil}, {"Kontextfehler", context.Canceled},
		{"Fehler von src", outbound.ErrBlobBeschaedigt}, {"ErrUngueltig", model.ErrUngueltig},
	}
	closeFormen := []struct {
		name string
		err  error
	}{
		{"gelingt", nil}, {"fremder Fehler", errPlatte}, {"Kontextfehler", context.DeadlineExceeded}, {"Fehler der Tabelle", outbound.ErrBlobFehlt},
	}
	n := 0
	for _, s := range stellenOps {
		for _, a := range ausgaenge {
			err := a.err
			if a.name == "Fehler des Lesers" {
				err = &model.DatenFehler{Ursache: model.DatenWert}
				if s.op == "ParseMetadata" {
					err = &model.MetadatenFehler{Ursache: model.UrsacheStruktur}
				}
			}
			ref := laufeMitClose(t, s.op, err, nil)
			for _, c := range closeFormen {
				n++
				name := s.name + " / " + a.name + " / Close " + c.name
				got := laufeMitClose(t, s.op, err, c.err)
				// Close gleich Open gleich 1 je Strom, vor dem nächsten Open.
				strome := got.w.blobs.strome()
				if len(strome) == 0 || got.zaehler["Close"] != got.zaehler["Blobs.Open"] {
					t.Errorf("%s: Close %d, Open %d", name, got.zaehler["Close"], got.zaehler["Blobs.Open"])
				}
				for i, st := range strome {
					if st.geschlossen() != 1 {
						t.Errorf("%s: Strom %d %d-mal geschlossen", name, i, st.geschlossen())
					}
				}
				offen := false
				for _, e := range got.w.p.folge() {
					switch e {
					case "Blobs.Open":
						if offen {
							t.Errorf("%s: Open vor Close", name)
						}
						offen = true
					case "Close":
						offen = false
					}
				}
				// Dieselbe Rückgabe wie mit gelingendem Close.
				switch {
				case err == nil:
					if got.err != nil || got.r.Import.ReportJSON != ref.r.Import.ReportJSON || got.r.Neu != ref.r.Neu {
						t.Errorf("%s: Fehler %v oder anderes Ergebnis", name, got.err)
					}
				case a.name == "Kontextfehler":
					if got.err != err {
						t.Errorf("%s: Fehler %v, erwartet derselbe Kontextfehler", name, got.err)
					}
				default:
					var f, rf *inbound.Fehler
					if !errors.As(got.err, &f) || !errors.As(ref.err, &rf) || f.Code != rf.Code || f.Status != rf.Status || !errors.Is(got.err, err) {
						t.Errorf("%s: Fehler %v, erwartet wie ohne Fehler von Close: %v", name, got.err, ref.err)
					}
				}
				// Dieselben Zähler aller Ports.
				if !maps.Equal(got.zaehler, ref.zaehler) {
					t.Errorf("%s: Zähler %v, erwartet %v", name, got.zaehler, ref.zaehler)
				}
			}
		}
	}
	if n != 60 {
		t.Errorf("Menge %d, erwartet 60", n)
	}
}
