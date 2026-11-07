package mock_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// fremderKnoten ist ein Knoten des Filters, den der Interpreter nicht kennt: sein Aufruf endet mit ErrUngueltig.
type fremderKnoten struct{}

func (fremderKnoten) IstFilterAusdruck() {}

// lesefall ist ein Fehlerfall des Lesepfads: der Aufbau trägt den Kanarienwert in Name des Sets, Text des Schlüssels, Name einer
// Property oder Wert im Payload.
type lesefall struct {
	name    string
	code    model.Code
	ursache bool
	bau     func(u *umg)
}

// lesefaelle sind Auslöser der acht Codes des Lesepfads (`ODR-E3001`, `E6001`, `E6003`, `E6004`, `E6005`, `E6008`, `E7001`,
// `E9001`), dazu die beiden Codes des Interpreters (`E6005`, `E6001`) mit Ursache.
func lesefaelle() []lesefall {
	k := kanarienwert
	roh := func(e error) error { return fmt.Errorf("%w %s", e, k) }
	return []lesefall{
		{"ODR-E3001", model.CodeE3001, true, func(u *umg) {
			u.menge(model.Optionen{})
			u.p.setzeFehler("Blobs.Open", roh(outbound.ErrBlobFehlt))
		}},
		{"ODR-E6001 Filter an der Entity", model.CodeE6001, false, func(u *umg) { u.entity(model.Optionen{Filter: filterOpt(nameGleich(k))}, "1") }},
		{"ODR-E6001 Interpreter", model.CodeE6001, true, func(u *umg) {
			u.menge(model.Optionen{OrderBy: []model.Sortierung{{Name: "Ref"}}})
		}},
		{"ODR-E6003", model.CodeE6003, false, func(u *umg) {
			u.k.StrictCoverage = true
			u.co.setze(baselineID, setProducts, model.CoveragePartial)
			u.menge(model.Optionen{})
		}},
		{"ODR-E6004", model.CodeE6004, true, func(u *umg) {
			u.menge(model.Optionen{})
			u.p.setzeFehler("MetadatenCodec.ParseMetadata", roh(model.ErrMetadatenUngueltig))
		}},
		{"ODR-E6005 Schlüssel", model.CodeE6005, false, func(u *umg) {
			u.ressource(model.RessourceEntity, setProducts, model.Optionen{}, zahl(k, "1"))
		}},
		{"ODR-E6005 Interpreter", model.CodeE6005, true, func(u *umg) {
			u.menge(model.Optionen{OrderBy: []model.Sortierung{{Name: k}}})
		}},
		{"ODR-E6008", model.CodeE6008, false, func(u *umg) { u.ressource(model.RessourceMenge, k+"Set", model.Optionen{}) }},
		{"ODR-E7001", model.CodeE7001, true, func(u *umg) {
			u.menge(model.Optionen{})
			u.p.setzeFehler("Entities.ListSet", roh(outbound.ErrNichtGefunden))
		}},
		{"ODR-E9001", model.CodeE9001, false, func(u *umg) {
			u.en.setze(setProducts, eintrag(setProducts, "1", `{"Id":1,"`+k+`":"`+k+`"}`, alleNamen()...))
			u.menge(model.Optionen{})
		}},
	}
}

// TestFehlerWertLesen hält L37 (Teil 2 von Z19, SPEC-195 Nr. 10): jeder Code des Lesepfads trägt den Status seines Codes, keine
// Ursache (außer wo SPEC-109 Nr. 5 sie führt) und im Error() keinen Wert (Canary: Name des Sets, Text des Schlüssels, Name
// einer Property, Wert im Payload, Text des Fehlers des Ports). Schleife über die 8 Codes in 10 Auslösern.
func TestFehlerWertLesen(t *testing.T) {
	codes := map[model.Code]bool{}
	for _, f := range lesefaelle() {
		u := leseUmg(t)
		f.bau(u)
		_, err := u.leseGET()
		fe := erwarteFehler(t, f.name, err, f.code)
		if fe == nil {
			continue
		}
		codes[fe.Code] = true
		if (fe.Ursache != nil) != f.ursache {
			t.Errorf("%s: Ursache %v, erwartet vorhanden=%v", f.name, fe.Ursache, f.ursache)
		}
		if fe.Status != f.code.Stati()[0] || strings.Contains(fe.Error(), kanarienwert) {
			t.Errorf("%s: Status %d, Error() %q", f.name, fe.Status, fe.Error())
		}
	}
	soll := []model.Code{model.CodeE3001, model.CodeE6001, model.CodeE6003, model.CodeE6004, model.CodeE6005, model.CodeE6008, model.CodeE7001, model.CodeE9001}
	if len(codes) != len(soll) {
		t.Errorf("%d Codes erreicht, erwartet %d", len(codes), len(soll))
	}
	for _, c := range soll {
		if !codes[c] {
			t.Errorf("%s nicht erreicht", c)
		}
	}
}

// codeListen sind die Aufteilung von model.AlleCodes() in die 13 erreichbaren Codes (SPEC-195 Nr. 10) und die übrigen.
func codeListen() (erreichbar, uebrige []model.Code) {
	e := []string{"ODR-E3001", "ODR-E5003", "ODR-E5004", "ODR-E5005", "ODR-E6001", "ODR-E6002", "ODR-E6003", "ODR-E6004", "ODR-E6005", "ODR-E6006",
		"ODR-E6008", "ODR-E7001", "ODR-E9001"}
	u := []string{"ODR-E2001", "ODR-E2101", "ODR-E2401", "ODR-E4001", "ODR-E5001", "ODR-E5002", "ODR-E5006", "ODR-E5007", "ODR-E5008", "ODR-E5009", "ODR-E5010",
		"ODR-E6007", "ODR-E6101", "ODR-E6102", "ODR-E7002", "ODR-E7003", "ODR-E7004", "ODR-E7005", "ODR-E7006", "ODR-E7007", "ODR-E7008", "ODR-E7009",
		"ODR-E7010", "ODR-E7011", "ODR-E7012"}
	for _, t := range e {
		c, _ := model.ParseCode(t)
		erreichbar = append(erreichbar, c)
	}
	for _, t := range u {
		c, _ := model.ParseCode(t)
		uebrige = append(uebrige, c)
	}
	return erreichbar, uebrige
}

// TestRueckgabenMenge hält L37 und M5 (SPEC-196 Nr. 9): die 13 erreichbaren Codes und die übrigen sind eine Aufteilung von
// model.AlleCodes() (jeder Code des Modells steht in genau einer der beiden Listen, die Längen ergeben die Länge der Quelle);
// ein neuer Code im Modell macht den Zähler rot. Die erreichbaren Codes kommen je mindestens einmal in den Läufen vor und kein
// anderer; die Mengen der Arten und Zustände im Test haben die Größe der Modell-Mengen.
func TestRueckgabenMenge(t *testing.T) {
	erreichbar, uebrige := codeListen()
	alle := model.AlleCodes()
	if len(erreichbar)+len(uebrige) != len(alle) {
		t.Fatalf("%d + %d Codes in den Listen, %d im Modell", len(erreichbar), len(uebrige), len(alle))
	}
	for _, c := range alle {
		if slices.Contains(erreichbar, c) == slices.Contains(uebrige, c) {
			t.Errorf("%s steht in keiner oder in beiden Listen", c)
		}
	}
	gesehen := map[model.Code]bool{}
	for _, f := range fehlerFaelle() {
		_, err := laufeFall(t, f)
		if c := fehlercode(err); c != 0 {
			gesehen[c] = true
		}
	}
	for _, f := range lesefaelle() {
		u := leseUmg(t)
		f.bau(u)
		_, err := u.leseGET()
		if c := fehlercode(err); c != 0 {
			gesehen[c] = true
		}
	}
	for _, c := range erreichbar {
		if !gesehen[c] {
			t.Errorf("%s kommt in keinem Lauf vor", c)
		}
	}
	for c := range gesehen {
		if !slices.Contains(erreichbar, c) {
			t.Errorf("%s kommt vor und ist nicht erreichbar", c)
		}
	}
	arten := []model.RessourcenArt{model.RessourceMetadaten, model.RessourceMenge, model.RessourceEntity, model.RessourceAnzahl}
	zustaende := []model.CoverageStatus{"complete", "partial", "unknown"}
	if len(arten) != len(model.AlleRessourcenArten()) || len(zustaende) != len(model.AlleCoverageStatus()) {
		t.Errorf("%d Arten und %d Zustände im Test, %d und %d im Modell", len(arten), len(zustaende), len(model.AlleRessourcenArten()), len(model.AlleCoverageStatus()))
	}
	if !slices.Equal(arten, model.AlleRessourcenArten()) || !slices.Equal(zustaende, model.AlleCoverageStatus()) {
		t.Error("die Mengen des Tests weichen von den Modell-Mengen ab")
	}
}

// TestKeinFehlerobjektLesen hält L45 (Teil 2 von Z24): der Dienst ruft SchreibeFehler im Lesepfad nie; Schleife über die
// Auslöser der 8 Codes (L37) und die Formen der Kontextfehler (L43). Dazu prüft der Fake in t.Cleanup jeden Test des Pakets.
func TestKeinFehlerobjektLesen(t *testing.T) {
	for _, f := range lesefaelle() {
		u := leseUmg(t)
		f.bau(u)
		_, _ = u.leseGET()
		if n := u.p.anzahl("Schreiber.SchreibeFehler"); n != 0 {
			t.Errorf("%s: SchreibeFehler %d", f.name, n)
		}
	}
	for _, s := range stellenLesen() {
		for _, form := range kontextFormen(s.tabelle) {
			u := leseUmg(t)
			s.bau(u)
			u.p.setzeFehler(s.op, form.err)
			_, _ = u.leseGET()
			if n := u.p.anzahl("Schreiber.SchreibeFehler"); n != 0 {
				t.Errorf("%s %s: SchreibeFehler %d", s.name, form.name, n)
			}
		}
	}
}

// stelleLesen ist eine Aufrufstelle S5 bis S11 des Lesens (SPEC-195 Nr. 11): die Operation, ein Aufbau, der sie erreicht, die
// Operationen bis dahin und ein Fehler der Tabelle in SPEC-109 Nr. 5, den die Stelle abbildet.
type stelleLesen struct {
	name    string
	op      string
	bau     func(u *umg)
	ops     []string
	tabelle error
	gegen   outbound.Gegenstand
}

func stellenLesen() []stelleLesen {
	kopf := []string{"URL.ParseURL", "Workspaces.Get", "Metadaten.Get", "Blobs.Open", "MetadatenCodec.ParseMetadata"}
	menge := func(u *umg) { u.menge(model.Optionen{}) }
	entity := func(u *umg) { u.entity(model.Optionen{}, "1") }
	anzahl := func(u *umg) { u.anzahl(model.Optionen{}) }
	mit := func(rest ...string) []string { return append(slices.Clone(kopf), rest...) }
	return []stelleLesen{
		{"S5", "MetadatenCodec.ParseMetadata", menge, mit(), model.ErrMetadatenUngueltig, outbound.GegenstandSonstiges},
		{"S6", "Coverage.Get", menge, mit("Coverage.Get"), outbound.ErrNichtGefunden, outbound.GegenstandSonstiges},
		{"S7", "Entities.ListSet", menge, mit("Coverage.Get", "Entities.ListSet"), outbound.ErrNichtGefunden, outbound.GegenstandWorkspace},
		{"S8", "Entities.Get", entity, mit("Entities.Get"), outbound.ErrNichtGefunden, outbound.GegenstandWorkspace},
		{"S9", "Schreiber.SchreibeEntity", entity, mit("Entities.Get", "Schreiber.SchreibeEntity"), model.ErrUngueltig, outbound.GegenstandSonstiges},
		{"S10", "Schreiber.SchreibeCollection", menge, mit("Coverage.Get", "Entities.ListSet", "Schreiber.SchreibeCollection"), model.ErrUngueltig, outbound.GegenstandSonstiges},
		{"S11", "Schreiber.SchreibeAnzahl", anzahl, mit("Coverage.Get", "Entities.ListSet", "Schreiber.SchreibeAnzahl"), model.ErrUngueltig, outbound.GegenstandSonstiges},
	}
}

// kontextFormen sind die vier Formen eines Kontextfehlers (M6): nackt Canceled, nackt DeadlineExceeded, verpackt und
// errors.Join neben einem Fehler der Tabelle.
func kontextFormen(tabelle error) []struct {
	name string
	err  error
	ctx  error
} {
	return []struct {
		name string
		err  error
		ctx  error
	}{
		{"Canceled", context.Canceled, context.Canceled},
		{"DeadlineExceeded", context.DeadlineExceeded, context.DeadlineExceeded},
		{"verpackt", fmt.Errorf("port: %w", context.DeadlineExceeded), context.DeadlineExceeded},
		{"Join", errors.Join(context.Canceled, tabelle), context.Canceled},
	}
}

// TestKontextLesen hält L43 (M6, SPEC-195 Nr. 2 und 11): jeder Aufruf von ParseMetadata (S5), Coverage.Get (S6), ListSet (S7),
// Entities.Get (S8) und des Schreibers (S9, S10, S11) erhält das ctx des Aufrufs unverändert; ein Kontextfehler kehrt ohne Code
// zurück, auch in errors.Join neben einem Fehler der Tabelle, danach kein weiterer Port, der Strom ist einmal geschlossen.
// Form 4 × Stelle 7 = 28.
func TestKontextLesen(t *testing.T) {
	stellen := stellenLesen()
	if len(stellen) != 7 {
		t.Fatalf("%d Stellen, erwartet 7 (S5 bis S11; Teil 1 führt S2, S3, S4 und S12)", len(stellen))
	}
	n := 0
	for _, s := range stellen {
		u := leseUmg(t)
		s.bau(u)
		if _, err := u.leseGET(); err != nil {
			t.Fatalf("%s ohne Fehler: %v", s.name, err)
		}
		for _, e := range u.p.von(s.op) {
			if e.ctx != u.ctx {
				t.Errorf("%s: ctx %v statt dem des Aufrufs", s.name, e.ctx)
			}
		}
		for _, form := range kontextFormen(s.tabelle) {
			n++
			u := leseUmg(t)
			s.bau(u)
			u.p.setzeFehler(s.op, form.err)
			_, err := u.leseGET()
			if err == nil || fehlercode(err) != 0 || !errors.Is(err, form.ctx) {
				t.Errorf("%s %s: %v, erwartet den Kontextfehler ohne Code", s.name, form.name, err)
			}
			if ops := u.p.ops(); !slices.Equal(ops, s.ops) {
				t.Errorf("%s %s: Aufrufe %v, erwartet %v", s.name, form.name, ops, s.ops)
			}
			for _, st := range u.bl.strome() {
				if st.geschlossen() != 1 {
					t.Errorf("%s %s: Strom %d Mal geschlossen", s.name, form.name, st.geschlossen())
				}
			}
		}
	}
	if n != 28 {
		t.Errorf("%d Läufe, erwartet 28", n)
	}
}

// TestPortFehler hält L41 (M18, SPEC-109 Nr. 5, SPEC-195 Nr. 11): Fehler der Ports im Lesepfad mit dem Gegenstand je Aufrufstelle,
// ein Fall je Stelle und nicht einer an der ersten der Kette. Stelle 7 × Fehler 6 × Form 2 = 84, dazu ErrMetadatenUngueltig an
// S5 = 85; Orakel von Hand. S6 und S8 mit ErrNichtGefunden sind kein Fehler (Coverage `unknown`, Entity fehlt).
func TestPortFehler(t *testing.T) {
	fehler := []error{outbound.ErrNichtGefunden, outbound.ErrBereitsVorhanden, outbound.ErrBeschaedigt, outbound.ErrBlobFehlt, model.ErrUngueltig, errors.New("roh " + kanarienwert)}
	e30, e90, e71, e77 := model.CodeE3001, model.CodeE9001, model.CodeE7001, model.CodeE7007
	sonstiges := [6]model.Code{e90, e90, e30, e30, e90, e90}
	workspace := [6]model.Code{e71, e77, e30, e30, e90, e90}
	soll := map[string][6]model.Code{"S5": sonstiges, "S6": sonstiges, "S7": workspace, "S8": workspace, "S9": sonstiges, "S10": sonstiges, "S11": sonstiges}
	n := 0
	for _, s := range stellenLesen() {
		for i, f := range fehler {
			for _, verpackt := range []bool{false, true} {
				n++
				u := leseUmg(t)
				s.bau(u)
				err := f
				if verpackt {
					err = fmt.Errorf("port %s: %w", s.name, f)
				}
				u.p.setzeFehler(s.op, err)
				_, got := u.leseGET()
				code := soll[s.name][i]
				name := fmt.Sprintf("%s %v verpackt=%v", s.name, f, verpackt)
				switch {
				case s.name == "S6" && i == 0:
					if got != nil {
						t.Errorf("%s: %v, erwartet die Coverage unknown", name, got)
					}
				case s.name == "S8" && i == 0:
					_ = erwarteFehler(t, name, got, model.CodeE6008) // Entity fehlt, Coverage complete
				default:
					if fe := erwarteFehler(t, name, got, code); fe != nil && !errors.Is(fe, f) {
						t.Errorf("%s: Ursache nicht erreichbar", name)
					}
				}
			}
		}
	}
	u := leseUmg(t)
	u.menge(model.Optionen{})
	u.p.setzeFehler("MetadatenCodec.ParseMetadata", model.ErrMetadatenUngueltig)
	_, got := u.leseGET()
	_ = erwarteFehler(t, "S5 ErrMetadatenUngueltig", got, model.CodeE6004)
	if n != 84 {
		t.Errorf("%d Läufe, erwartet 84 (und einer)", n)
	}
}

// TestRangfolgeLesen hält L44 (M4, SPEC-195 Nr. 3): „Set unbekannt“, „Coverage im strikten Modus“ und „Fehler des Interpreters“
// gegen jede Stufe von Teil 1 und gegeneinander; der erste Verstoß entscheidet, die Ports dahinter bleiben ungerufen. 27 Paare
// mit mindestens einer der drei Stufen (Teil 1 führt die 27 der ersten acht), aufgefächert nach den Formen der Stufen.
func TestRangfolgeLesen(t *testing.T) {
	kopf := []string{"URL.ParseURL", "Workspaces.Get", "Metadaten.Get", "Blobs.Open", "MetadatenCodec.ParseMetadata"}
	neue := []stufe{
		{name: "Set unbekannt", code: model.CodeE6008, stufe: 8, ops: kopf, bau: func(u *umg, _ *inbound.Anfrage) {
			u.ressource(model.RessourceMenge, "Unbekannt", model.Optionen{})
		}},
		{name: "Coverage im strikten Modus", code: model.CodeE6003, stufe: 9, ops: append(slices.Clone(kopf), "Coverage.Get"), bau: func(u *umg, _ *inbound.Anfrage) {
			u.k.StrictCoverage = true
			u.co.setze(baselineID, setProducts, model.CoveragePartial)
			u.co.setze(baselineID, "Unbekannt", model.CoveragePartial) // das unbekannte Set trägt ebenfalls partial: nur die Reihenfolge der Stufen unterscheidet ODR-E6008 von ODR-E6003
			u.menge(model.Optionen{})
		}},
		{name: "Fehler des Interpreters", code: model.CodeE6005, stufe: 10, ops: append(slices.Clone(kopf), "Coverage.Get", "Entities.ListSet"), bau: func(u *umg, _ *inbound.Anfrage) {
			u.menge(model.Optionen{OrderBy: []model.Sortierung{{Name: "Unbekannt"}}})
		}},
	}
	alle := append(stufenM4(), neue...)
	pruefe := func(name string, verletzt ...stufe) {
		u := leseUmg(t)
		a := metadatenAnfrage()
		for k := len(verletzt) - 1; k >= 0; k-- {
			verletzt[k].bau(u, &a)
		}
		ctx := u.ctx
		if slices.ContainsFunc(verletzt, func(s stufe) bool { return s.keinCtx }) {
			ctx = keinKontext()
		}
		_, err := u.mocker().Mock(ctx, a)
		erste := verletzt[0]
		if erste.code == 0 {
			var f *inbound.Fehler
			if !errors.Is(err, model.ErrUngueltig) || errors.As(err, &f) {
				t.Errorf("%s: Fehler %v, erwartet ErrUngueltig ohne Code", name, err)
			}
		} else {
			_ = erwarteFehler(t, name, err, erste.code)
		}
		if ops := u.p.ops(); !slices.Equal(ops, erste.ops) {
			t.Errorf("%s: Aufrufe %v, erwartet %v", name, ops, erste.ops)
		}
	}
	paare := map[[2]int]bool{}
	faecher := 0
	for i, s := range alle {
		if s.stufe >= 8 {
			pruefe(s.name, s)
		}
		for j := i + 1; j < len(alle); j++ {
			if s.stufe == alle[j].stufe || alle[j].stufe < 8 {
				continue
			}
			paare[[2]int{s.stufe, alle[j].stufe}] = true
			faecher++
			pruefe(s.name+" vor "+alle[j].name, s, alle[j])
		}
	}
	if len(paare) != 27 || faecher != 12*3+3 {
		t.Errorf("%d Paare der Stufen und %d aufgefächert, erwartet 27 und 39", len(paare), faecher)
	}
}

// TestAufrufzahlen hält L40 (SPEC-195 Nr. 2, SPEC-196, SPEC-197): Aufrufzahl je Port und Anfrage. Art 4 × Ausgang 3 (Erfolg,
// Fehler der Coverage, Set unbekannt) mit einer Soll-Tabelle; der Test vergleicht den Zähler jeder Methode jedes Ports mit dem Soll
// (Schleife über alle Methoden der Schnittstellen, Zählprüfung gegen NumMethod), auch der, die der Dienst nie rufen darf.
func TestAufrufzahlen(t *testing.T) {
	portTypen := []struct {
		praefix string
		typ     reflect.Type
	}{
		{"Workspaces", reflect.TypeFor[outbound.WorkspaceRepository]()}, {"Metadaten", reflect.TypeFor[outbound.MetadataRepository]()},
		{"Blobs", reflect.TypeFor[outbound.BlobRepository]()}, {"Entities", reflect.TypeFor[outbound.EntityRepository]()},
		{"Coverage", reflect.TypeFor[outbound.CoverageRepository]()}, {"MetadatenCodec", reflect.TypeFor[outbound.MetadataCodec]()},
		{"URL", reflect.TypeFor[outbound.URLCodec]()}, {"Schreiber", reflect.TypeFor[outbound.ODataSchreiber]()},
	}
	var ops []string
	for _, p := range portTypen {
		for i := 0; i < p.typ.NumMethod(); i++ {
			ops = append(ops, p.praefix+"."+p.typ.Method(i).Name)
		}
	}
	if len(ops) != 5+4+5+5+2+1+1+5 {
		t.Fatalf("%d Methoden der acht Ports, erwartet 28", len(ops))
	}
	kern := map[string]int{"URL.ParseURL": 1, "Workspaces.Get": 1, "Metadaten.Get": 1}
	mit := func(zusatz map[string]int) map[string]int {
		r := map[string]int{}
		for k, v := range kern {
			r[k] = v
		}
		for k, v := range zusatz {
			r[k] = v
		}
		return r
	}
	modell := map[string]int{"Blobs.Open": 1, "MetadatenCodec.ParseMetadata": 1}
	mitModell := func(zusatz map[string]int) map[string]int {
		r := mit(modell)
		for k, v := range zusatz {
			r[k] = v
		}
		return r
	}
	faelle := []struct {
		name  string
		bau   func(u *umg)
		fehl  string
		soll  map[string]int
		fehlt bool
	}{
		{"Metadaten Erfolg", func(u *umg) { u.standard(model.RessourceMetadaten, "1") }, "", mit(map[string]int{"Blobs.Open": 1, "Schreiber.SchreibeAnzahl": 1}), false},
		{"Metadaten Coverage-Fehler", func(u *umg) { u.standard(model.RessourceMetadaten, "1") }, "Coverage.Get", mit(map[string]int{"Blobs.Open": 1, "Schreiber.SchreibeAnzahl": 1}), false},
		{"Metadaten Set unbekannt", func(u *umg) { u.standard(model.RessourceMetadaten, "1") }, "", mit(map[string]int{"Blobs.Open": 1, "Schreiber.SchreibeAnzahl": 1}), false},
		{"Menge Erfolg", func(u *umg) { u.standard(model.RessourceMenge, "1") }, "", mitModell(map[string]int{"Coverage.Get": 1, "Entities.ListSet": 1, "Schreiber.SchreibeCollection": 1}), false},
		{"Menge Coverage-Fehler", func(u *umg) { u.standard(model.RessourceMenge, "1") }, "Coverage.Get", mitModell(map[string]int{"Coverage.Get": 1}), true},
		{"Menge Set unbekannt", func(u *umg) { u.ressource(model.RessourceMenge, "Nirgends", model.Optionen{}) }, "", mitModell(nil), true},
		{"Entity Erfolg", func(u *umg) { u.standard(model.RessourceEntity, "1") }, "", mitModell(map[string]int{"Entities.Get": 1, "Schreiber.SchreibeEntity": 1}), false},
		{"Entity Coverage-Fehler", func(u *umg) { u.standard(model.RessourceEntity, "99") }, "Coverage.Get", mitModell(map[string]int{"Entities.Get": 1, "Coverage.Get": 1}), true},
		{"Entity Set unbekannt", func(u *umg) { u.ressource(model.RessourceEntity, "Nirgends", model.Optionen{}, zahl("", "1")) }, "", mitModell(nil), true},
		{"Anzahl Erfolg", func(u *umg) { u.standard(model.RessourceAnzahl, "1") }, "", mitModell(map[string]int{"Coverage.Get": 1, "Entities.ListSet": 1, "Schreiber.SchreibeAnzahl": 1}), false},
		{"Anzahl Coverage-Fehler", func(u *umg) { u.standard(model.RessourceAnzahl, "1") }, "Coverage.Get", mitModell(map[string]int{"Coverage.Get": 1}), true},
		{"Anzahl Set unbekannt", func(u *umg) { u.ressource(model.RessourceAnzahl, "Nirgends", model.Optionen{}) }, "", mitModell(nil), true},
	}
	for _, f := range faelle {
		u := leseUmg(t)
		f.bau(u)
		if f.fehl != "" {
			u.p.setzeFehler(f.fehl, errors.New("roh "+kanarienwert))
		}
		_, err := u.leseGET()
		if (err != nil) != f.fehlt {
			t.Errorf("%s: Fehler %v", f.name, err)
		}
		for _, op := range ops {
			if got := u.p.anzahl(op); got != f.soll[op] {
				t.Errorf("%s: %s %d Mal, erwartet %d", f.name, op, got, f.soll[op])
			}
		}
	}
	if len(faelle) != 4*3 {
		t.Errorf("%d Fälle, erwartet 12", len(faelle))
	}
}

// TestInterpreterFehler hält L13 (SPEC-195 Nr. 11 S14 bis S16, SPEC-196 Nr. 8): Fehler des Interpreters an jeder Aufrufstelle:
// die Klasse einer Ursache bestimmt den Code (`ungueltig` ODR-E6005, `nicht_unterstuetzt` ODR-E6001, `nicht_beobachtet`
// ODR-E6003, ErrUngueltig ODR-E9001); ein Verstoß der Entities geht dem Fehler des Interpreters voraus. Schleife über Stelle 3 ×
// (AlleAbfrageUrsachen und AlleFilterTypUrsachen); die Paare, die der Parser oder der Dienst an einer Stelle nicht zulässt, stehen
// mit Grund in einer Liste; die Zählprüfung gilt gegen die beiden Quellen.
func TestInterpreterFehler(t *testing.T) {
	num := func(op model.VergleichOp, name, text string) model.FilterAusdruck {
		return vergleich(op, name, model.Literal{Art: model.LiteralZahl, Text: text})
	}
	filter := map[model.FilterTypUrsache]*model.FilterOption{
		model.FilterEigenschaftUnbekannt:      filterOpt(num(model.VergleichEq, "Unbekannt", "1")),
		model.FilterNichtBoolesch:             filterOpt(model.FilterEigenschaft{Name: "Name"}),
		model.FilterTypUnvertraeglich:         filterOpt(num(model.VergleichEq, "Name", "1")),
		model.FilterLiteralBereich:            filterOpt(num(model.VergleichEq, "Count", "99999999999")),
		model.FilterTypNichtUnterstuetzt:      filterOpt(vergleich(model.VergleichEq, "Blob", model.Literal{Art: model.LiteralZeichenkette, Text: "'x'"})),
		model.FilterOperatorNichtUnterstuetzt: filterOpt(vergleich(model.VergleichGt, "Ref", model.Literal{Art: model.LiteralGUID, Text: "00000000-0000-0000-0000-000000000000"})),
		model.FilterNichtBeobachtet:           filterOpt(nameGleich("a")),
	}
	filterCode := map[model.FilterTypUrsache]model.Code{
		model.FilterEigenschaftUnbekannt: model.CodeE6005, model.FilterNichtBoolesch: model.CodeE6005, model.FilterTypUnvertraeglich: model.CodeE6005,
		model.FilterLiteralBereich: model.CodeE6005, model.FilterTypNichtUnterstuetzt: model.CodeE6001, model.FilterOperatorNichtUnterstuetzt: model.CodeE6001,
		model.FilterNichtBeobachtet: model.CodeE6003,
	}
	// Eingaben: Aufbau (Set, Einträge) und Optionen; die Einträge kennen Name nur dort, wo die Ursache es nicht verlangt.
	seltsam := []model.EntityState{eintrag("Seltsam", "1", `{"Id":1}`, "Id", "Blob")}
	bekannt := []model.EntityState{produkt(1, "a"), produkt(2, "b")}
	nurID := []model.EntityState{produkt(1, "a", "Id"), produkt(2, "b", "Id")}
	type eingabe struct {
		set       string
		eintraege []model.EntityState
		opt       model.Optionen
		code      model.Code
	}
	abfrage := map[string]eingabe{
		"Menge/" + string(model.AbfrageEigenschaftUnbekannt):     {setProducts, bekannt, model.Optionen{OrderBy: []model.Sortierung{{Name: "Unbekannt"}}}, model.CodeE6005},
		"Menge/" + string(model.AbfrageTypNichtUnterstuetzt):     {"Seltsam", seltsam, model.Optionen{Select: &model.SelectOption{Namen: []string{"Blob"}}}, model.CodeE6001},
		"Menge/" + string(model.AbfrageOrdnungNichtUnterstuetzt): {setProducts, bekannt, model.Optionen{OrderBy: []model.Sortierung{{Name: "Ref"}}}, model.CodeE6001},
		"Menge/" + string(model.AbfrageNichtBeobachtet):          {setProducts, nurID, model.Optionen{OrderBy: []model.Sortierung{{Name: "Name"}}}, model.CodeE6003},
		"Entity/" + string(model.AbfrageEigenschaftUnbekannt):    {setProducts, bekannt, model.Optionen{Select: &model.SelectOption{Namen: []string{"Unbekannt"}}}, model.CodeE6005},
		"Entity/" + string(model.AbfrageTypNichtUnterstuetzt):    {"Seltsam", seltsam, model.Optionen{Select: &model.SelectOption{Namen: []string{"Blob"}}}, model.CodeE6001},
	}
	erreichbar := map[string]bool{}
	lauf := func(name string, art model.RessourcenArt, e eingabe) {
		erreichbar[name] = true
		u := leseUmg(t)
		u.co.setze(baselineID, e.set, model.CoverageComplete)
		u.en.setze(e.set, e.eintraege...)
		u.ressource(art, e.set, e.opt, zahl("", "1"))
		_, err := u.leseGET()
		_ = erwarteFehler(t, name, err, e.code)
	}
	for name, e := range abfrage {
		art := model.RessourceMenge
		if strings.HasPrefix(name, "Entity/") {
			art = model.RessourceEntity
		}
		lauf(name, art, e)
	}
	for u, f := range filter {
		set, es := setProducts, bekannt
		switch u {
		case model.FilterTypNichtUnterstuetzt:
			set, es = "Seltsam", seltsam
		case model.FilterNichtBeobachtet:
			es = nurID
		case model.FilterOperatorNichtUnterstuetzt:
			es = []model.EntityState{eintrag(setProducts, "1", `{"Id":1,"Ref":"00000000-0000-0000-0000-000000000000"}`, alleNamen()...)}
		}
		lauf("Menge/filter:"+string(u), model.RessourceMenge, eingabe{set, es, model.Optionen{Filter: f}, filterCode[u]})
		lauf("Anzahl/filter:"+string(u), model.RessourceAnzahl, eingabe{set, es, model.Optionen{Filter: f}, filterCode[u]})
	}
	nicht := map[string]string{}
	for _, u := range model.AlleAbfrageUrsachen() {
		nicht["Anzahl/"+string(u)] = "an der Anzahl übergibt der Dienst allein `$filter`"
	}
	nicht["Entity/"+string(model.AbfrageOrdnungNichtUnterstuetzt)] = "`$orderby` wirkt an der Entity nicht"
	nicht["Entity/"+string(model.AbfrageNichtBeobachtet)] = "bei der Entity ist die Maske nil, die Vollständigkeit prüft der Dienst"
	for _, u := range model.AlleFilterTypUrsachen() {
		nicht["Entity/filter:"+string(u)] = "`$filter` an der Entity ist ODR-E6001, bevor der Interpreter läuft"
	}
	zaehl := 0
	for _, stelle := range []string{"Menge", "Entity", "Anzahl"} {
		var namen []string
		for _, u := range model.AlleAbfrageUrsachen() {
			namen = append(namen, stelle+"/"+string(u))
		}
		for _, u := range model.AlleFilterTypUrsachen() {
			namen = append(namen, stelle+"/filter:"+string(u))
		}
		for _, n := range namen {
			zaehl++
			// Die Abfrage-Ursachen heißen wie die des Filters; die Namen mit „filter:“ gehören den Filter-Ursachen.
			if erreichbar[n] == (nicht[n] != "") {
				t.Errorf("%s: erreichbar=%v, nicht erreichbar mit Grund %q", n, erreichbar[n], nicht[n])
			}
		}
	}
	if zaehl != 3*(len(model.AlleAbfrageUrsachen())+len(model.AlleFilterTypUrsachen())) {
		t.Errorf("%d Paare", zaehl)
	}
	if len(erreichbar)+len(nicht) != zaehl {
		t.Errorf("%d erreichbar und %d nicht erreichbar, erwartet zusammen %d", len(erreichbar), len(nicht), zaehl)
	}
	// ErrUngueltig des Interpreters ist ODR-E9001: zwei Fälle je Stelle.
	doppelt := &model.SelectOption{Namen: []string{"Name", "Name"}}
	fremd := filterOpt(fremderKnoten{})
	fremdIn := filterOpt(model.FilterVergleich{Op: model.VergleichEq, Links: fremderKnoten{}, Rechts: model.FilterLiteral{Literal: model.Literal{Art: model.LiteralZahl, Text: "1"}}})
	ungueltig := []struct {
		name string
		art  model.RessourcenArt
		opt  model.Optionen
	}{
		{"Menge Select doppelt", model.RessourceMenge, model.Optionen{Select: doppelt}}, {"Menge fremder Knoten", model.RessourceMenge, model.Optionen{Filter: fremd}},
		{"Entity Select doppelt", model.RessourceEntity, model.Optionen{Select: doppelt}}, {"Entity Select leer", model.RessourceEntity, model.Optionen{Select: &model.SelectOption{}}},
		{"Anzahl fremder Knoten", model.RessourceAnzahl, model.Optionen{Filter: fremd}}, {"Anzahl fremder Knoten im Vergleich", model.RessourceAnzahl, model.Optionen{Filter: fremdIn}},
	}
	for _, g := range ungueltig {
		u := leseUmg(t)
		u.ressource(g.art, setProducts, g.opt, zahl("", "1"))
		_, err := u.leseGET()
		_ = erwarteFehler(t, g.name, err, model.CodeE9001)
	}
	// Ein Verstoß der Entities geht dem Fehler des Interpreters voraus (SPEC-196 Nr. 4 vor Nr. 5 und 8): 4 Paare je Stelle, an der
	// Entity die beiden, die ein einzelner Eintrag zulässt.
	verstoesse := map[string]func(es []model.EntityState) []model.EntityState{
		"Payload": func(es []model.EntityState) []model.EntityState { es[1].Payload = `{"Id":`; return es },
		"Schlüssel-Property": func(es []model.EntityState) []model.EntityState {
			es[1].Payload = `{"Name":"b"}`
			return es
		},
		"Eindeutigkeit": func(es []model.EntityState) []model.EntityState { es[1].EntityKey = es[0].EntityKey; return es },
		"redigiert in den Optionen": func(es []model.EntityState) []model.EntityState {
			es[1].Payload = `{"Id":2,"Name":"` + model.Marker + `"}`
			return es
		},
	}
	faelle := 0
	for _, stelle := range []model.RessourcenArt{model.RessourceMenge, model.RessourceAnzahl, model.RessourceEntity} {
		for name, bau := range verstoesse {
			if stelle == model.RessourceEntity && (name == "Eindeutigkeit" || name == "redigiert in den Optionen") {
				continue // ein Lookup liest einen Eintrag: kein Paar, keine Dublette, keine Option, die eine Property nennt
			}
			faelle++
			u := leseUmg(t)
			es := bau([]model.EntityState{produkt(1, "a"), produkt(2, "b")})
			opt := model.Optionen{}
			switch stelle {
			case model.RessourceMenge:
				opt.OrderBy = []model.Sortierung{{Name: "Name"}, {Name: "Unbekannt"}}
			case model.RessourceAnzahl:
				opt.Filter = filterOpt(model.FilterVerknuepfung{Op: model.VerknuepfungUnd, Links: nameGleich("a"), Rechts: num(model.VergleichEq, "Unbekannt", "1")})
			default:
				opt.Select = &model.SelectOption{Namen: []string{"Unbekannt"}}
				es = es[1:] // der Eintrag mit dem Verstoß ist der einzige; Schlüssel 2
			}
			u.en.setze(setProducts, es...)
			if stelle == model.RessourceEntity {
				u.ressource(stelle, setProducts, opt, zahl("", "2"))
			} else {
				u.ressource(stelle, setProducts, opt)
			}
			_, err := u.leseGET()
			_ = erwarteFehler(t, fmt.Sprintf("Art %d: %s vor Interpreter", stelle, name), err, model.CodeE9001)
		}
	}
	if faelle != 4+4+2 {
		t.Errorf("%d Paare, erwartet 10", faelle)
	}
}
