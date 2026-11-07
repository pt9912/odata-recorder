package mock_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// standard setzt die Ressource der Art: Menge, Anzahl und Metadaten ohne Optionen, die Entity mit dem Schlüssel 1 (key).
func (u *umg) standard(art model.RessourcenArt, key string) {
	switch art {
	case model.RessourceMetadaten:
		u.ressource(art, "", model.Optionen{})
	case model.RessourceEntity:
		u.entity(model.Optionen{}, key)
	default:
		u.ressource(art, setProducts, model.Optionen{})
	}
}

// artAnzahl und artOp ordnen jeder Art die Operation des Schreibers zu, die ihre Ausgabe schreibt.
func artOp(art model.RessourcenArt) string {
	switch art {
	case model.RessourceMenge:
		return "Schreiber.SchreibeCollection"
	case model.RessourceEntity:
		return "Schreiber.SchreibeEntity"
	}
	return "Schreiber.SchreibeAnzahl"
}

// coverageHeader liefert den Wert des Headers der Coverage (leer, wenn er fehlt) und sagt, ob er steht.
func coverageHeader(a inbound.Antwort) (string, bool) {
	h, da := a.Header["X-Recorder-Coverage"]
	if !da || len(h) != 1 {
		return "", da
	}
	return h[0], true
}

// TestCoverageLesen hält L18 (SPEC-197 Nr. 1): Coverage.Get genau einmal mit (ctx, Baseline des Workspace, Set) bei Collection,
// Anzahl und Lookup ohne Treffer, sonst nie; ErrNichtGefunden ist `unknown`, ein anderer Fehler nach SPEC-109, ein Status außerhalb
// der drei (auch in anderer Schreibweise, leer oder mit Leerraum) ODR-E9001. Art 5 × Fehler 4 und Status 4 × Art 3.
func TestCoverageLesen(t *testing.T) {
	arten := []struct {
		name string
		art  model.RessourcenArt
		key  string
		zahl int
	}{
		{"Collection", model.RessourceMenge, "1", 1}, {"Anzahl", model.RessourceAnzahl, "1", 1}, {"Lookup mit Treffer", model.RessourceEntity, "1", 0},
		{"Lookup ohne Treffer", model.RessourceEntity, "99", 1}, {"Metadaten", model.RessourceMetadaten, "1", 0},
	}
	fehler := []struct {
		name string
		err  error
		code model.Code
	}{
		{"nicht gefunden", outbound.ErrNichtGefunden, model.CodeE6003}, {"beschädigt", outbound.ErrBeschaedigt, model.CodeE3001},
		{"roh", errors.New("roh " + kanarienwert), model.CodeE9001}, {"kein Fehler", nil, 0},
	}
	n := 0
	for _, a := range arten {
		for _, f := range fehler {
			n++
			u := leseUmg(t)
			u.standard(a.art, a.key)
			u.p.setzeFehler("Coverage.Get", f.err)
			ant, err := u.leseGET()
			ev := u.p.von("Coverage.Get")
			if len(ev) != a.zahl {
				t.Errorf("%s/%s: %d Aufrufe von Coverage.Get, erwartet %d", a.name, f.name, len(ev), a.zahl)
			}
			if a.zahl == 1 && (ev[0].ctx != u.ctx || !reflect.DeepEqual(ev[0].args, []any{baselineID, setProducts})) {
				t.Errorf("%s/%s: Coverage.Get mit %v", a.name, f.name, ev[0].args)
			}
			var soll model.Code
			switch {
			case a.zahl == 0 || a.art != model.RessourceEntity && f.err == nil:
				soll = 0
			case a.art == model.RessourceEntity && f.err == nil:
				soll = model.CodeE6008 // Zustand complete, kein Treffer
			case a.art != model.RessourceEntity && f.code == model.CodeE6003:
				soll = 0 // unknown ist bei Collection und Anzahl im Modus ohne STRICT eine Antwort
			default:
				soll = f.code
			}
			if fehlercode(err) != soll {
				t.Errorf("%s/%s: %v, erwartet %s", a.name, f.name, err, soll)
			}
			if err == nil {
				lies(t, ant)
			}
		}
	}
	if n != 20 {
		t.Errorf("%d Zellen, erwartet 20", n)
	}
	for _, status := range []model.CoverageStatus{"", "COMPLETE", "complete ", "unbekannt"} {
		for _, art := range []struct {
			art model.RessourcenArt
			key string
		}{{model.RessourceMenge, "1"}, {model.RessourceAnzahl, "1"}, {model.RessourceEntity, "99"}} {
			u := leseUmg(t)
			u.co.setze(baselineID, setProducts, status)
			u.standard(art.art, art.key)
			_, err := u.leseGET()
			erwarteFehler(t, fmt.Sprintf("Status %q Art %d", status, art.art), err, model.CodeE9001)
		}
	}
}

// TestCoverageMatrix hält L19 und L25 (M11, SPEC-197 Nr. 2, 3): Art 3 (Collection, Anzahl, Lookup ohne Treffer) × Zustand 3 ×
// Modus von STRICT_COVERAGE 2 × Methode 2 = 36 Läufe, Orakel die Tabelle; im strikten Modus ohne Lesen der Entities, der
// Header trägt den Zustand, der Lookup ohne Treffer ist in beiden Modi gleich.
func TestCoverageMatrix(t *testing.T) {
	n := 0
	for _, art := range []model.RessourcenArt{model.RessourceMenge, model.RessourceAnzahl, model.RessourceEntity} {
		for _, status := range model.AlleCoverageStatus() {
			for _, strikt := range []bool{false, true} {
				for _, methode := range []string{"GET", "HEAD"} {
					n++
					u := leseUmg(t)
					u.k.StrictCoverage = strikt
					u.co.setze(baselineID, setProducts, status)
					u.standard(art, "99")
					ant, err := u.rufe(anfrage(methode, dienstRoot+"/x"))
					name := fmt.Sprintf("Art %d, %s, strikt=%v, %s", art, status, strikt, methode)
					switch {
					case art == model.RessourceEntity && status == model.CoverageComplete:
						erwarteFehler(t, name, err, model.CodeE6008)
					case art == model.RessourceEntity:
						erwarteFehler(t, name, err, model.CodeE6003)
					case status != model.CoverageComplete && strikt:
						erwarteFehler(t, name, err, model.CodeE6003)
						if u.p.anzahl("Entities.ListSet") != 0 || u.p.anzahl(artOp(art)) != 0 {
							t.Errorf("%s: Entities gelesen oder geschrieben", name)
						}
					default:
						if err != nil {
							t.Errorf("%s: %v", name, err)
							continue
						}
						if h, _ := coverageHeader(ant); h != string(status) {
							t.Errorf("%s: Header %q", name, h)
						}
						lies(t, ant)
					}
				}
			}
		}
	}
	if n != 36 {
		t.Errorf("%d Läufe, erwartet 36", n)
	}
}

// TestCoverageHeader hält L20 (SPEC-197 Nr. 4): X-Recorder-Coverage steht genau auf der 200-Antwort von Collection und Anzahl
// mit dem Zustand, nie auf Entity, $metadata oder einem Fehler. Art 4 × Zustand 3 × Ergebnis 2.
func TestCoverageHeader(t *testing.T) {
	n := 0
	for _, art := range model.AlleRessourcenArten() {
		for _, status := range model.AlleCoverageStatus() {
			for _, fehler := range []bool{false, true} {
				n++
				u := leseUmg(t)
				u.co.setze(baselineID, setProducts, status)
				u.standard(art, "1")
				if fehler {
					u.p.setzeFehler(artOp(art), model.ErrUngueltig)
				}
				ant, err := u.leseGET()
				h, da := coverageHeader(ant)
				soll := !fehler && (art == model.RessourceMenge || art == model.RessourceAnzahl)
				if (err != nil) != fehler || da != soll || (soll && h != string(status)) {
					t.Errorf("Art %d, %s, Fehler=%v: Header %q (%v), Fehler %v", art, status, fehler, h, da, err)
				}
			}
		}
	}
	if n != 24 {
		t.Errorf("%d Läufe, erwartet 24", n)
	}
}

// TestLeereMenge hält L21 (SPEC-197 Nr. 5, SPEC-196 Nr. 4 a): null beobachtete Entities sind kein Leerheitsbeweis. Zustand 3 ×
// Modus 2 × Option 2 (ohne, `$filter` auf eine Property): `complete` ist `value: []` und Anzahl 0 mit Maske nil, unvollständig
// ist im strikten Modus ODR-E6003, sonst die leere Collection mit Header und leerer Maske (ein Filter nennt dann eine nicht
// beobachtete Property).
func TestLeereMenge(t *testing.T) {
	n := 0
	for _, status := range model.AlleCoverageStatus() {
		for _, strikt := range []bool{false, true} {
			for _, filter := range []bool{false, true} {
				n++
				u := leseUmg(t)
				u.k.StrictCoverage = strikt
				u.co.setze(baselineID, setProducts, status)
				u.en.setze(setProducts)
				opt := model.Optionen{Count: model.WahrOption{Gesetzt: true, Wert: true}}
				if filter {
					opt.Filter = filterOpt(nameGleich("x"))
				}
				u.menge(opt)
				ant, err := u.leseGET()
				name := fmt.Sprintf("%s, strikt=%v, Filter=%v", status, strikt, filter)
				unvollstaendig := status != model.CoverageComplete
				switch {
				case unvollstaendig && (strikt || filter):
					soll := model.CodeE6003
					erwarteFehler(t, name, err, soll)
				case err != nil:
					t.Errorf("%s: %v", name, err)
				default:
					lies(t, ant)
					c := u.letzteCollection()
					if len(c.Entities) != 0 || !c.Anzahl.Vorhanden || c.Anzahl.Wert != 0 {
						t.Errorf("%s: Antwort %+v", name, c)
					}
					if h, _ := coverageHeader(ant); h != string(status) {
						t.Errorf("%s: Header %q", name, h)
					}
				}
			}
		}
	}
	if n != 12 {
		t.Errorf("%d Läufe, erwartet 12", n)
	}
}

// TestCoverageBaseline hält L22 (SPEC-197 Nr. 2, 6): die Coverage hängt an der Baseline des Workspace; zwei Workspaces mit
// verschiedener Baseline desselben Service haben ihre Zustände, zusätzliche Einträge und Tombstones ändern sie nie. Zwei
// Baselines × Zustand 3; dazu: strikt, `partial` und `$orderby=Unbekannt` ergibt ODR-E6003 (die Coverage geht dem Interpreter voraus).
func TestCoverageBaseline(t *testing.T) {
	zustaende := model.AlleCoverageStatus()
	for i, s := range zustaende {
		anders := zustaende[(i+1)%len(zustaende)]
		u := leseUmg(t)
		u.ws.setze(model.Workspace{ID: "ws2", ServiceID: dienstID, BaselineSnapshotID: "b2"})
		u.co.setze("b1", setProducts, s)
		u.co.setze("b2", setProducts, anders)
		tot := produkt(9, "t")
		tot.Deleted = true
		u.en.setze(setProducts, produkt(1, "a"), tot)
		u.menge(model.Optionen{})
		for _, w := range []struct {
			id   string
			soll model.CoverageStatus
		}{{"ws1", s}, {"ws2", anders}} {
			a := anfrage("GET", dienstRoot+"/x")
			a.Header[headerName] = []string{w.id}
			ant, err := u.rufe(a)
			if err != nil {
				t.Fatal(err)
			}
			lies(t, ant)
			if h, _ := coverageHeader(ant); h != string(w.soll) {
				t.Errorf("Workspace %s: Header %q, erwartet %q", w.id, h, w.soll)
			}
		}
	}
	for _, s := range zustaende[1:] {
		u := leseUmg(t)
		u.k.StrictCoverage = true
		u.co.setze(baselineID, setProducts, s)
		u.menge(model.Optionen{OrderBy: []model.Sortierung{{Name: "Unbekannt"}}})
		_, err := u.leseGET()
		erwarteFehler(t, "strikt "+string(s)+" und unbekannte Property", err, model.CodeE6003)
	}
	// Ein Set ohne Eintrag der Coverage ist `unknown`, auch wenn alle Einträge des Workspace vollständig sind.
	u := leseUmg(t)
	u.co = &coverage{p: u.p}
	u.k.Coverage = u.co
	u.menge(model.Optionen{})
	ant, err := u.leseGET()
	if err != nil {
		t.Fatal(err)
	}
	lies(t, ant)
	if h, _ := coverageHeader(ant); h != "unknown" {
		t.Errorf("ohne Eintrag: Header %q", h)
	}
}

// TestHead hält L29 (SPEC-198 Nr. 5): HEAD auf Collection, Entity und Anzahl ist GET mit leerem Body, Status, Header und Fehler
// gleich. Art 3 × Ergebnis 2 (Erfolg, ODR-E6003).
func TestHead(t *testing.T) {
	for _, art := range []model.RessourcenArt{model.RessourceMenge, model.RessourceEntity, model.RessourceAnzahl} {
		for _, fehler := range []bool{false, true} {
			var antworten [2]inbound.Antwort
			var fehlerA [2]error
			var bodies [2][]byte
			for i, m := range []string{"GET", "HEAD"} {
				u := leseUmg(t)
				u.k.StrictCoverage = true
				status := model.CoverageComplete
				key := "1"
				if fehler {
					status, key = model.CoveragePartial, "99"
				}
				u.co.setze(baselineID, setProducts, status)
				u.standard(art, key)
				antworten[i], fehlerA[i] = u.rufe(anfrage(m, dienstRoot+"/x"))
				if fehlerA[i] == nil {
					bodies[i] = lies(t, antworten[i])
				}
			}
			get, head := antworten[0], antworten[1]
			name := fmt.Sprintf("Art %d, Fehler=%v", art, fehler)
			if !reflect.DeepEqual(fehlerA[0], fehlerA[1]) || (fehler && fehlercode(fehlerA[0]) != model.CodeE6003) {
				t.Errorf("%s: Fehler GET %v, HEAD %v", name, fehlerA[0], fehlerA[1])
			}
			if fehler {
				continue
			}
			if get.Status != head.Status || !reflect.DeepEqual(get.Header, head.Header) || len(bodies[0]) == 0 || len(bodies[1]) != 0 {
				t.Errorf("%s: GET %d %v %d Bytes, HEAD %d %v %d Bytes", name, get.Status, get.Header, len(bodies[0]), head.Status, head.Header, len(bodies[1]))
			}
		}
	}
}

// TestDarstellungLeer hält L46 (M20, SPEC-198 Nr. 8): ein Kennzeichen mit leerem Namen oder leerem Wert und ein leerer
// ContentType sind bei Collection, Entity und Anzahl ODR-E9001 ohne Ursache, bei GET und HEAD, ohne Header und ohne Body; der
// Schreiber ist genau einmal gerufen. Ausgabe 3 × Verstoß 4 × Methode 2 = 24 Läufe; dazu geht die Coverage im strikten Modus
// voraus (ODR-E6003, der Schreiber bleibt ungerufen).
func TestDarstellungLeer(t *testing.T) {
	verstoesse := []model.Darstellung{
		{ContentType: "text/x", Kennzeichen: model.Kopfzeile{Name: "", Wert: "v"}},
		{ContentType: "text/x", Kennzeichen: model.Kopfzeile{Name: "X-K", Wert: ""}},
		{ContentType: "text/x", Kennzeichen: model.Kopfzeile{}},
		{ContentType: "", Kennzeichen: model.Kopfzeile{Name: "X-K", Wert: "v"}},
	}
	n := 0
	for _, art := range []model.RessourcenArt{model.RessourceMenge, model.RessourceEntity, model.RessourceAnzahl} {
		for _, d := range verstoesse {
			for _, m := range []string{"GET", "HEAD"} {
				n++
				u := leseUmg(t)
				u.c.sch.kopfe = map[string]model.Darstellung{artOp(art): d}
				u.standard(art, "1")
				ant, err := u.rufe(anfrage(m, dienstRoot+"/x"))
				f := erwarteFehler(t, fmt.Sprintf("Art %d %+v %s", art, d, m), err, model.CodeE9001)
				if f != nil && f.Ursache != nil || ant.Body != nil || ant.Header != nil || u.p.anzahl(artOp(art)) != 1 {
					t.Errorf("Art %d %+v %s: Antwort %+v, %d Aufrufe des Schreibers", art, d, m, ant, u.p.anzahl(artOp(art)))
				}
			}
		}
	}
	if n != 24 {
		t.Errorf("%d Läufe, erwartet 24", n)
	}
	for _, art := range []model.RessourcenArt{model.RessourceMenge, model.RessourceAnzahl} {
		u := leseUmg(t)
		u.k.StrictCoverage = true
		u.co.setze(baselineID, setProducts, model.CoveragePartial)
		u.c.sch.kopfe = map[string]model.Darstellung{artOp(art): {}}
		u.standard(art, "1")
		_, err := u.leseGET()
		erwarteFehler(t, "strikt partial", err, model.CodeE6003)
		if u.p.anzahl(artOp(art)) != 0 {
			t.Errorf("Art %d: Schreiber trotz ODR-E6003 gerufen", art)
		}
	}
}
