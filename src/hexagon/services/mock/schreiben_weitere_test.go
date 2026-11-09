package mock_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// mutationsCodes sind die Codes der Läufe der Mutationen, die TestRueckgabenMenge zu den Läufen des Lesens zählt (W63): ein
// Fehler des Eingabe-Lesers (ODR-E6007, ODR-E6002), ein Duplicate Key (ODR-E6009), ein abgesichert fehlender Eintrag (ODR-E6008),
// eine nicht erfüllte Vorbedingung (ODR-E6010).
func mutationsCodes(t *testing.T) map[model.Code]bool {
	t.Helper()
	r := map[model.Code]bool{}
	u := schreibUmg(t)
	u.p.setzeFehler("Eingabe.LeseEntity", &model.EingabeFehler{Ursache: model.EingabeSyntax})
	_, err := u.mutiere("POST", "7", eingabeIDName("7", "x"))
	r[fehlercode(err)] = true
	u = schreibUmg(t)
	_, err = u.mutiere("POST", "2", eingabeIDName("2", "x"))
	r[fehlercode(err)] = true
	u = schreibUmg(t)
	_, err = u.mutiere("DELETE", "99", eingabeIDName("99", "x"))
	r[fehlercode(err)] = true
	u = etagUmg(t, true, 1, 1, 1)
	_, err = u.mutiereMit("DELETE", "2", map[string][]string{"If-Match": {`W/"0"`}})
	r[fehlercode(err)] = true
	return r
}

// TestPostKey (W16): bei POST ist der Key der des Bodys (EntityKey gleich Eingabe.Schluessel), der Mock vergibt keinen. Schlüsseltyp
// {Edm.Int32, Edm.String, zwei Teile, Int32 und String} × Form 2 (der Schlüssel ist ein Wert, der keinem Teil des Bodys gleicht).
func TestPostKey(t *testing.T) {
	faelle := []struct {
		set, typ, key string
	}{
		{setProducts, typProduct, "7"}, {"Sks", "N.Sk", "'abc'"}, {"Pairs", "N.Pair", "A='x',B='y'"}, {"Bunt", "N.Bunt", "I=3,S='z'"},
		{"Viele", "N.Viele", "5"}, // Properties nicht unterstützten Typs fehlen im Payload (SPEC-193 Nr. 3 c)
	}
	n := 0
	for _, f := range faelle {
		for _, schluessel := range []string{f.key, f.key + "9"} {
			n++
			u := schreibUmg(t)
			u.ressource(model.RessourceMenge, f.set, model.Optionen{})
			e := model.EntityEingabe{Typ: f.typ, Schluessel: schluessel}
			switch f.set {
			case "Viele":
				e.Properties = []model.PropertyWert{pw("Id", model.ArtInt32, "5")}
			case setProducts:
				e.Properties = []model.PropertyWert{pw("Id", model.ArtInt32, "7")}
			case "Sks":
				e.Properties = []model.PropertyWert{pw("K", model.ArtString, "abc")}
			case "Pairs":
				e.Properties = []model.PropertyWert{pw("A", model.ArtString, "x"), pw("B", model.ArtString, "y")}
			default:
				e.Properties = []model.PropertyWert{pw("I", model.ArtInt32, "3"), pw("S", model.ArtString, "z")}
			}
			u.ei.ergebnis = e
			a, err := u.rufe(anfrage("POST", dienstRoot+"/x"))
			erwarteStatus(t, f.set, a, err, 201)
			put := u.p.von("Entities.Put")
			if len(put) != 1 || put[0].args[0].(model.EntityState).EntityKey != schluessel {
				t.Errorf("%s: Put %v, erwartet EntityKey %q", f.set, put, schluessel)
			}
		}
	}
	if n != 10 {
		t.Errorf("Menge %d, erwartet 10", n)
	}
}

// TestKeineVergabe (W16): der Quelltext des Pakets importiert weder time noch math/rand, crypto noch eine Kennung.
func TestKeineVergabe(t *testing.T) {
	dateien, err := filepath.Glob("*.go")
	if err != nil || len(dateien) == 0 {
		t.Fatalf("keine Quelle: %v", err)
	}
	n := 0
	for _, d := range dateien {
		if strings.HasSuffix(d, "_test.go") {
			continue
		}
		n++
		f, err := parser.ParseFile(token.NewFileSet(), d, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range f.Imports {
			p := strings.Trim(imp.Path.Value, `"`)
			if p == "time" || strings.HasPrefix(p, "math/rand") || strings.HasPrefix(p, "crypto/") || strings.HasPrefix(p, "hash/") || strings.Contains(p, "uuid") {
				t.Errorf("%s importiert %s", d, p)
			}
		}
	}
	if n < 8 {
		t.Errorf("%d Quelldateien gelesen", n)
	}
	if _, err := os.Stat("schreiben.go"); err != nil {
		t.Errorf("schreiben.go fehlt: %v", err)
	}
}

// TestTypAbgleich (W27, M11): bei PATCH und PUT an einem vorhandenen Eintrag ist Eingabe.Typ gleich dem Typ des Eintrags, sonst
// ODR-E6007 ohne Ursache. Typ des Eintrags 2 (Set, abgeleiteter) × Typ der Eingabe 3 (Set, derselbe abgeleitete, ein anderer
// abgeleiteter) × Methode 2 = 12.
func TestTypAbgleich(t *testing.T) {
	n := 0
	for _, eintragTyp := range []string{typProduct, typSpecial} {
		for _, eingabeTyp := range []string{typProduct, typSpecial, "N.Tief"} {
			for _, m := range []string{"PATCH", "PUT"} {
				n++
				u := schreibUmg(t)
				alt := voll(2, "b")
				if eintragTyp == typSpecial {
					alt.Payload = `{"@odata.type":"#N.Special","Id":2,"Name":"b"}`
					alt.KnownProperties = []string{"Id", "Name"}
				}
				u.en.setze(setProducts, alt)
				e := eingabeIDName("2", "neu")
				e.Typ = eingabeTyp
				a, err := u.mutiere(m, "2", e)
				kontext := fmt.Sprintf("%s Eintrag %s Eingabe %s", m, eintragTyp, eingabeTyp)
				if eintragTyp == eingabeTyp {
					erwarteStatus(t, kontext, a, err, 204)
					continue
				}
				if f := erwarteFehler(t, kontext, err, e6007); f != nil && f.Ursache != nil {
					t.Errorf("%s: Ursache %v", kontext, f.Ursache)
				}
				u.ohneSchreiben(kontext)
			}
		}
	}
	if n != 12 {
		t.Errorf("Menge %d, erwartet 12", n)
	}
}

// TestBefundVorTyp (W28): der Befund geht dem Abgleich des Typs vor: PATCH auf partial mit falschem Typ ist ODR-E6003, PUT auf einen
// Tombstone mit falschem Typ ODR-E6008.
func TestBefundVorTyp(t *testing.T) {
	u := schreibUmg(t)
	u.en.setze(setProducts, teilEintrag("2", 2))
	e := eingabeIDName("2", "neu")
	e.Typ = typSpecial
	_, err := u.mutiere("PATCH", "2", e)
	pruefeFehler(t, "PATCH partial falscher Typ", err, e6003)
	u = schreibUmg(t)
	tomb := voll(2, "b")
	tomb.Deleted = true
	u.en.setze(setProducts, tomb)
	_, err = u.mutiere("PUT", "2", e)
	pruefeFehler(t, "PUT Tombstone falscher Typ", err, e6008)
}

// TestNachDelete (W42, M25): nach DELETE erscheint der Eintrag weder in der Collection noch in der Anzahl, der Lookup ist
// ODR-E6008 (auch bei Coverage partial: der Tombstone belegt das Nichtvorhandensein). Art 3 × Coverage 2 × Position 3 = 18.
func TestNachDelete(t *testing.T) {
	n := 0
	for _, art := range []string{"Collection", "Anzahl", "Lookup"} {
		for _, cov := range []model.CoverageStatus{model.CoverageComplete, model.CoveragePartial} {
			for _, p := range []string{"1", "3", "5"} {
				n++
				u := schreibUmg(t)
				u.co.setze(baselineID, setProducts, cov)
				u.en.setze(setProducts, fuenfEintraege()...)
				a, err := u.mutiere("DELETE", p, eingabeIDName(p, "x"))
				erwarteStatus(t, "DELETE", a, err, 204)
				kontext := fmt.Sprintf("%s %s Position %s", art, cov, p)
				switch art {
				case "Collection":
					u.menge(model.Optionen{})
					u.c.sch.zeigeIDs = true
					if _, err := u.leseGET(); err != nil {
						t.Errorf("%s: %v", kontext, err)
						continue
					}
					var want []string
					for _, k := range []string{"1", "2", "3", "4", "5"} {
						if k != p {
							want = append(want, k)
						}
					}
					if got := ids(u.letzteCollection().Entities); got != strings.Join(want, ",") {
						t.Errorf("%s: Schlüssel %q, erwartet %q", kontext, got, strings.Join(want, ","))
					}
				case "Anzahl":
					u.anzahl(model.Optionen{})
					if _, err := u.leseGET(); err != nil {
						t.Errorf("%s: %v", kontext, err)
						continue
					}
					ev := u.p.von("Schreiber.SchreibeAnzahl")
					if got := ev[len(ev)-1].args[1]; got != int64(4) {
						t.Errorf("%s: Anzahl %v, erwartet 4", kontext, got)
					}
				default:
					u.entity(model.Optionen{}, p)
					_, err := u.leseGET()
					pruefeFehler(t, kontext, err, e6008)
				}
			}
		}
	}
	if n != 18 {
		t.Errorf("Menge %d, erwartet 18", n)
	}
}

// TestCoverageUnveraendert (W48, M18): eine Mutation ändert die Coverage nie und leitet complete nie ab: vom Coverage-Port läuft
// allein Get; ein strenges GET auf einem Set mit Coverage partial oder unknown bleibt ODR-E6003 nach POST, PATCH, PUT und DELETE.
// Methode 4 × Zustand 3 = 12.
func TestCoverageUnveraendert(t *testing.T) {
	n := 0
	for _, m := range []string{"POST", "PATCH", "PUT", "DELETE"} {
		for _, cov := range model.AlleCoverageStatus() {
			n++
			u := schreibUmg(t)
			u.k.StrictCoverage = true
			u.co.m = map[string]model.Coverage{}
			if cov != model.CoverageUnknown {
				u.co.setze(baselineID, setProducts, cov)
			}
			id := schluesselFuer(m)
			a, err := u.mutiere(m, id, eingabeIDName(id, "neu"))
			erwarteStatus(t, string(cov)+" "+m, a, err, sollStatus(m))
			zeilen := 1
			if cov == model.CoverageUnknown {
				zeilen = 0
			}
			if u.p.anzahl("Coverage.List") != 0 || len(u.co.m) != zeilen {
				t.Errorf("%s %s: Coverage.List %d, Zeilen %d", m, cov, u.p.anzahl("Coverage.List"), len(u.co.m))
			}
			u.menge(model.Optionen{})
			_, err = u.leseGET()
			if cov == model.CoverageComplete {
				if err != nil {
					t.Errorf("%s complete: GET %v", m, err)
				}
				continue
			}
			pruefeFehler(t, fmt.Sprintf("%s %s GET streng", m, cov), err, e6003)
		}
	}
	if n != 12 || !slices.Contains(model.AlleCoverageStatus(), model.CoverageUnknown) {
		t.Errorf("Menge %d, erwartet 12", n)
	}
}

// TestVertragsverletzungen (Z-Liste, Zweige des Programmfehlers): was der Leser, das Modell oder der Speicher gegen ihren Vertrag
// liefern, ist ODR-E9001 ohne Ursache, und der Zustand bleibt unverändert (die Transaktion rollt zurück). Jeder Fall erreicht einen
// Zweig, den kein Request eines Clients erreicht.
func TestVertragsverletzungen(t *testing.T) {
	faelle := []struct {
		name string
		meth string
		bau  func(u *umg) model.EntityEingabe
	}{
		{"Set ohne auflösbaren Typ", "POST", func(u *umg) model.EntityEingabe {
			u.ressource(model.RessourceMenge, "Waise", model.Optionen{})
			return eingabeIDName("7", "neu")
		}},
		{"Typ der Eingabe unbekannt", "POST", func(*umg) model.EntityEingabe {
			e := eingabeIDName("7", "neu")
			e.Typ = "N.Nirgends"
			return e
		}},
		{"Typ der Eingabe nicht aus dem Set abgeleitet", "POST", func(*umg) model.EntityEingabe {
			e := eingabeIDName("7", "neu")
			e.Typ = "N.Fremd"
			return e
		}},
		{"Wert der Art redigiert bei POST", "POST", func(*umg) model.EntityEingabe {
			e := eingabeIDName("7", "neu")
			e.Properties[1].Wert = model.Wert{Art: model.ArtRedigiert}
			return e
		}},
		{"Wert der Art redigiert bei PATCH", "PATCH", func(*umg) model.EntityEingabe {
			e := eingabeIDName("2", "neu")
			e.Properties[1].Wert = model.Wert{Art: model.ArtRedigiert}
			return e
		}},
		{"erforderliche Property nicht genannt", "POST", func(*umg) model.EntityEingabe {
			e := eingabeIDName("7", "neu")
			e.Properties = e.Properties[1:]
			return e
		}},
		{"Property außerhalb des Typs genannt", "POST", func(*umg) model.EntityEingabe {
			e := eingabeIDName("7", "neu")
			e.Properties = append(e.Properties, pw("Nirgends", model.ArtString, "x"))
			return e
		}},
		{"erforderliche Property nicht genannt bei PUT", "PUT", func(*umg) model.EntityEingabe {
			e := eingabeIDName("2", "neu")
			e.Properties = e.Properties[1:]
			return e
		}},
		{"Payload des Eintrags kein JSON", "PATCH", func(u *umg) model.EntityEingabe {
			alt := voll(2, "b")
			alt.Payload = "kein JSON"
			u.en.setze(setProducts, alt)
			return eingabeIDName("2", "neu")
		}},
		{"@odata.type des Eintrags ohne #", "PATCH", func(u *umg) model.EntityEingabe {
			alt := voll(2, "b")
			alt.Payload = `{"@odata.type":"N.Special","Id":2}`
			u.en.setze(setProducts, alt)
			return eingabeIDName("2", "neu")
		}},
	}
	for _, f := range faelle {
		u := schreibUmg(t)
		id := schluesselFuer(f.meth)
		u.setzeMutation(f.meth, id, model.EntityEingabe{})
		u.ei.ergebnis = f.bau(u)
		vorher, rev := u.en.inhalt(), u.ws.revision(wsID)
		_, err := u.rufe(anfrage(f.meth, dienstRoot+"/x"))
		if g := erwarteFehler(t, f.name, err, e9001); g != nil && g.Ursache != nil {
			t.Errorf("%s: Ursache %v", f.name, g.Ursache)
		}
		u.unveraendert(vorher, rev, f.name)
	}
}

// TestSchluesselZeichen (F-3, SPEC-193 Nr. 1, SPEC-126 Nr. 11): der Schlüssel eines Edm.String steht in der URL als Literal mit
// verdoppelten einfachen Anführungszeichen; der Abgleich mit dem Body vergleicht den Text byte-gleich. Text 9 × Methode {PATCH,
// PUT} = 18 Läufe mit gleichem Schlüssel (204) und 9 mit abweichendem Body (ODR-E6007). Das Orakel steht je Zeile von Hand.
func TestSchluesselZeichen(t *testing.T) {
	ue := string([]byte{0xC3, 0xBC})
	faelle := []struct{ text, literal string }{
		{"O'Neil", "'O''Neil'"}, {"''", "''''''"}, {"'", "''''"}, {"x''y", "'x''''y'"}, {"a b", "'a b'"}, {"50%", "'50%'"},
		{"a\"b", "'a\"b'"}, {"a/b\\c", "'a/b\\c'"}, {"m" + ue + "ller", "'m" + ue + "ller'"},
	}
	n := 0
	for _, f := range faelle {
		for _, m := range []string{"PATCH", "PUT"} {
			for _, passt := range []bool{true, false} {
				n++
				u := schreibUmg(t)
				alt := model.EntityState{
					WorkspaceID: wsID, EntitySet: "Sks", EntityKey: f.literal, Completeness: model.EntityComplete, Version: 1,
					Payload: "{}", KnownProperties: []string{"K", "Name"},
				}
				u.en.setze("Sks", alt)
				u.ressource(model.RessourceEntity, "Sks", model.Optionen{},
					model.SchluesselTeil{Literal: model.Literal{Art: model.LiteralZeichenkette, Text: f.literal}})
				body := f.text
				if !passt {
					body += "!"
				}
				u.ei.ergebnis = model.EntityEingabe{Typ: "N.Sk", Properties: []model.PropertyWert{pw("K", model.ArtString, body), pw("Name", model.ArtString, "neu")}}
				a, err := u.rufe(anfrage(m, dienstRoot+"/x"))
				kontext := fmt.Sprintf("%s %q passt=%v", m, f.text, passt)
				if passt {
					erwarteStatus(t, kontext, a, err, 204)
				} else {
					pruefeFehler(t, kontext, err, e6007)
				}
			}
		}
	}
	if n != 36 {
		t.Errorf("Menge %d, erwartet 36", n)
	}
}

// TestStellenImQuelltext (F-5, SPEC-195 Nr. 11): die Aufrufe der Ports im Quelltext des Pakets sind 22 Aufrufstellen im Text, 20 davon
// mit ctx; eine weitere Stelle macht den Test rot, bis die Listen der Fänger (TestKontextMutation, TestPortFehlerMutation, TestKontextLesen)
// und diese Zahl nachgezogen sind. 22 Stellen im Text gegen 26 Stellen S1 bis S26 der Spezifikation: S4/S5 (Modell) und S2, S3, S6, S8
// teilen eine Stelle des Textes mit der Mutation oder dem Lesen.
func TestStellenImQuelltext(t *testing.T) {
	ops := map[string]bool{"WithWorkspaceTransaction": true, "Get": true, "Put": true, "Advance": true, "SchreibeEntity": true, "SchreibeAnzahl": true,
		"SchreibeCollection": true, "Adresse": true, "LeseEntity": true, "Open": true, "ParseMetadata": true, "ParseURL": true, "ListSet": true}
	ports := map[string]bool{"URL": true, "Metadaten": true, "Workspaces": true, "Blobs": true, "Coverage": true, "Entities": true, "Schreiber": true,
		"Eingabe": true, "Transaktion": true, "sch": true}
	dateien, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	gesamt, mitCtx := 0, 0
	for _, d := range dateien {
		if strings.HasSuffix(d, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(token.NewFileSet(), d, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(f, func(nd ast.Node) bool {
			ce, ok := nd.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := ce.Fun.(*ast.SelectorExpr)
			if !ok || !ops[sel.Sel.Name] {
				return true
			}
			name := ""
			switch x := sel.X.(type) {
			case *ast.SelectorExpr:
				name = x.Sel.Name
			case *ast.Ident:
				name = x.Name
			}
			if !ports[name] {
				return true
			}
			gesamt++
			if id, ok := ce.Args[0].(*ast.Ident); ok && id.Name == "ctx" {
				mitCtx++
			}
			return true
		})
	}
	if gesamt != 22 || mitCtx != 20 {
		t.Errorf("%d Aufrufe der Ports im Quelltext, davon %d mit ctx; erwartet 22 und 20 (SPEC-195 Nr. 11)", gesamt, mitCtx)
	}
}
