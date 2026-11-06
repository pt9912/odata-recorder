package jsoncodec_test

import (
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// eVerstossDeuten ist einer der elf Verstöße des Deutens (Zeilen 16, 33, 34): das Set, die Ursache, der Name des Mitglieds und der Wert
// für die Stellen a und b (zwei Mitglieder verschiedenen Namens im selben Body).
type eVerstossDeuten struct {
	ursache  model.EingabeUrsache
	set      string
	name     [2]string
	wert     string
	einmalig bool // der Name steht höchstens einmal im Body (`@odata.type`, `Id`)
}

func eVerstoesseDeuten() []eVerstossDeuten {
	nul := esc("0000")
	return []eVerstossDeuten{
		{model.EingabeUnbekannteProperty, "Sts", [2]string{"ax", "bx"}, "1", false},
		{model.EingabeTyp, "Sts", [2]string{"an", "bn"}, `"s"`, false},
		{model.EingabeWert, "Sts", [2]string{"an", "bn"}, "1.5", false},
		{model.EingabeNull, "Sts", [2]string{"ap", "bp"}, "null", false},
		{model.EingabePropertyTyp, "Sts", [2]string{"ad", "bd"}, "1", false},
		{model.EingabeNavigation, "Sts", [2]string{"av", "bv"}, "{}", false},
		{model.EingabeBind, "Sts", [2]string{"av@odata.bind", "bv@odata.bind"}, `"x"`, false},
		{model.EingabeOffen, "Sto", [2]string{"ax", "bx"}, "1", false},
		{model.EingabeODataTyp, "Sts", [2]string{"@odata.type", "@odata.type"}, "5", true},
		{model.EingabeStruktur, "Sts", [2]string{"a@b@c", "d@b@c"}, "1", false},
		{model.EingabeSchluessel, "Sts", [2]string{"Id", "Id"}, `"a` + nul + `b"`, true},
	}
}

// member liefert das Mitglied der Stelle i (0: a, 1: b) als Paar Name, Wert.
func (v eVerstossDeuten) member(i int) []string { return []string{v.name[i], v.wert} }

// TestEingabeLesenVorDeuten hält Zeile 16: jeder Verstoß des Lesens geht jedem des Deutens vor, auch wenn der des Deutens im Dokument früher
// steht. Menge 55 = 5 Ursachen des Lesens × 11 Verstöße des Deutens.
func TestEingabeLesenVorDeuten(t *testing.T) {
	n := 0
	arten := eArten()
	for _, d := range eVerstoesseDeuten() {
		for _, l := range eLesenAusloeserListe()[:5] {
			vor := eQuote(d.name[0]) + ":" + d.wert
			body, g := l.body(vor)
			a := eA(d.set, arten[n%3], body)
			a.grenze = g
			got, err := a.los()
			eFehler(t, string(d.ursache)+" vor "+string(l.ursache), got, err, l.ursache, "")
			n++
		}
	}
	if n != 55 {
		t.Errorf("Menge %d, erwartet 55", n)
	}
}

// eErwartet liefert Ursache und Ort des ersten Verstoßes im Ablauf (SPEC-181 Nr. 8): `@odata.type` zuerst, dann der erste im Dokument.
func eErwartet(erster, zweiter eVerstossDeuten) (model.EingabeUrsache, string) {
	if erster.ursache == model.EingabeODataTyp || zweiter.ursache == model.EingabeODataTyp {
		return model.EingabeODataTyp, eTypOrt
	}
	return erster.ursache, "/" + erster.name[0]
}

// TestEingabeRangfolge hält Zeile 33: der erste Verstoß im Ablauf entscheidet, kein Vorrang einer Klasse, `fehlend` zuletzt. Menge: 121 geordnete
// Paare (davon 2 unmöglich: ein Typ ist offen oder nicht; 2 teilen den Namen `@odata.type` bzw. `Id` und bilden ein Mitglied) und 22 Paare
// Verstoß und fehlende Property.
func TestEingabeRangfolge(t *testing.T) {
	vs := eVerstoesseDeuten()
	arten := eArten()
	paare, unmoeglich, n := 0, 0, 0
	for i, a := range vs {
		for j, b := range vs {
			paare++
			if (a.ursache == model.EingabeOffen && b.ursache == model.EingabeUnbekannteProperty) ||
				(b.ursache == model.EingabeOffen && a.ursache == model.EingabeUnbekannteProperty) {
				unmoeglich++
				continue
			}
			set := a.set
			if b.set == "Sto" {
				set = "Sto"
			}
			mitglieder := a.member(0)
			if !(i == j && a.einmalig) {
				mitglieder = append(mitglieder, b.member(1)...)
			}
			ursache, ort := eErwartet(a, b)
			got, err := eA(set, arten[n%3], eObj(mitglieder...)).los()
			eFehler(t, string(a.ursache)+" vor "+string(b.ursache), got, err, ursache, ort)
			n++
		}
	}
	if paare != 121 || unmoeglich != 2 || n != 119 {
		t.Errorf("Paare %d, unmöglich %d, geprüft %d", paare, unmoeglich, n)
	}
	for _, v := range vs {
		for _, art := range []model.EingabeArt{model.EingabePost, model.EingabePut} {
			got, err := eA(v.set, art, eObj(v.member(0)...)).los()
			eFehler(t, string(v.ursache)+" und fehlend "+string(art), got, err, v.ursache, "/"+v.name[0])
			n++
		}
	}
	if n != 119+22 {
		t.Errorf("Menge %d, erwartet 141", n)
	}
}

// TestEingabeStellen hält Zeile 34: derselbe Verstoß an jeder Stelle der Mitgliederliste mit dem Ort der Stelle. Menge 132 = 11 × 4 × 3.
func TestEingabeStellen(t *testing.T) {
	n := 0
	for _, v := range eVerstoesseDeuten() {
		stellen := map[string][]string{
			"einziges":  append([]string{}, v.member(0)...),
			"erstes":    append(v.member(0), "n2", "1", "n3", "2"),
			"mittleres": append(append([]string{"n1", "1"}, v.member(0)...), "n3", "2"),
			"letztes":   append([]string{"n1", "1", "n2", "2"}, v.member(0)...),
		}
		for stelle, paare := range stellen {
			for _, art := range eArten() {
				ort := "/" + v.name[0]
				got, err := eA(v.set, art, eObj(paare...)).los()
				eFehler(t, string(v.ursache)+" "+stelle+" "+string(art), got, err, v.ursache, ort)
				n++
			}
		}
	}
	if n != 132 {
		t.Errorf("Menge %d, erwartet 132", n)
	}
}

// eNullwertFall ist ein Fehlerfall, in dem der Leser schon Properties gelesen hat, bevor der Verstoß kommt.
type eNullwertFall struct {
	ursache model.EingabeUrsache
	ort     string
	set     string
	art     model.EingabeArt
	body    string
}

// eNullwertFaelle haben je Ursache (19) einen Body mit gültigen Mitgliedern vor dem Verstoß.
func eNullwertFaelle() []eNullwertFall {
	gut := `"K":1,"P0":"x","P1":true,"P2":7,"P3":8,"P4":9.5,"P5":"0f8fad5b-d9cb-469f-a165-70867728950e","P6":"2020-01-01"`
	vor := `"n1":1,`
	nul := esc("0000")
	post, patch := model.EingabePost, model.EingabePatch
	return []eNullwertFall{
		{model.EingabeSyntax, "", "Zs", patch, `{` + gut + `,"Q0":x}`},
		{model.EingabeKodierung, "", "Zs", patch, `{` + gut + `,"Q0":"` + "\xff" + `"}`},
		{model.EingabeTiefe, "", "Zs", patch, `{` + gut + `,"@a":` + strings.Repeat("[", 513) + `}`},
		{model.EingabeDoppelterName, "", "Zs", patch, `{` + gut + `,"K":2}`},
		{model.EingabeContentType, "", "Zs", patch, `{` + gut + `}`},
		{model.EingabeStruktur, "/a@b@c", "Zs", patch, `{` + gut + `,"a@b@c":1}`},
		{model.EingabeODataTyp, eTypOrt, "Zs", patch, `{` + gut + `,"@odata.type":"#Ns.X"}`},
		{model.EingabeUnbekannteProperty, "/zz", "Zs", patch, `{` + gut + `,"zz":1}`},
		{model.EingabeTyp, "/Q0", "Zs", patch, `{` + gut + `,"Q0":5}`},
		{model.EingabeWert, "/Q3", "Zs", patch, `{` + gut + `,"Q3":1.5}`},
		{model.EingabeNull, "/P7", "Zs", patch, `{` + gut + `,"P7":null}`},
		{model.EingabeZahlZuLang, "/Q4", "Zs", patch, `{` + gut + `,"Q4":1e1024}`},
		{model.EingabeSchluessel, "/Id", "Sts", patch, `{` + vor + `"Id":"a` + nul + `"}`},
		{model.EingabeFehlend, "/P7", "Zs", post, `{` + gut + `}`},
		{model.EingabePropertyTyp, "/ad", "Sts", patch, `{` + vor + `"ad":1}`},
		{model.EingabeNavigation, "/av", "Sts", patch, `{` + vor + `"av":{}}`},
		{model.EingabeBind, "/av@odata.bind", "Sts", patch, `{` + vor + `"av@odata.bind":"x"}`},
		{model.EingabeOffen, "/zz", "Sto", patch, `{` + vor + `"zz":1}`},
	}
}

// TestEingabeNullwertBeiFehler hält Zeile 37: bei jedem Fehler ist das Ergebnis der Nullwert, auch nach gelesenen Properties. Die Fälle haben
// je Ursache einen Body mit gültigen Mitgliedern vor dem Verstoß (18 Ursachen; `content_type` hat keinen Body, siehe unten).
func TestEingabeNullwertBeiFehler(t *testing.T) {
	gesehen := map[model.EingabeUrsache]bool{}
	for _, f := range eNullwertFaelle() {
		a := eA(f.set, f.art, f.body)
		if f.ursache == model.EingabeContentType {
			a.ct = []string{"text/plain"}
		}
		got, err := a.los()
		eFehler(t, string(f.ursache), got, err, f.ursache, f.ort)
		gesehen[f.ursache] = true
	}
	for _, u := range model.AlleEingabeUrsachen() {
		if !gesehen[u] && u != model.EingabeZuGross {
			t.Errorf("Ursache %s ohne Fall", u)
		}
	}
	a := eA("Zs", model.EingabePatch, `{"K":1,"P0":"x"}`+strings.Repeat(" ", 20))
	a.grenze = 18
	got, err := a.los()
	eFehler(t, "zu_gross", got, err, model.EingabeZuGross, "")
}

// TestEingabeFehlerEnthaeltKeinenWert hält die Allaussage der Zeile 35: in keinem Fehlerfall steht ein Wert des Bodys im Text; der Test setzt
// GEHEIM-7 in jeden Wert.
func TestEingabeFehlerEnthaeltKeinenWert(t *testing.T) {
	geheim := `"GEHEIM-7"`
	faelle := []struct {
		set, body string
	}{
		{"Zs", `{"K":1,"P2":` + geheim + `}`}, {"Zs", `{"K":1,"P3":` + geheim + `}`}, {"Zs", `{"K":1,"P5":` + geheim + `}`},
		{"Zs", `{"K":1,"P6":` + geheim + `}`}, {"Zs", `{"K":1,"P7":` + geheim + `}`}, {"Zs", `{"K":` + geheim + `}`},
		{"Zs", `{"@odata.type":` + geheim + `}`}, {"Sts", `{"Id":"GEHEIM-7` + esc("0000") + `"}`}, {"Sts", `{"ad":` + geheim + `}`},
		{"Sts", `{"av":` + geheim + `}`}, {"Sts", `{"av@odata.bind":` + geheim + `}`}, {"Sto", `{"ax":` + geheim + `}`},
		{"Sts", `{"ax":` + geheim + `}`}, {"Sts", `{"a@b@c":` + geheim + `}`}, {"Sts", `{"an":` + geheim + `}`},
		{"Sts", `{"n1":1,"n1":` + geheim + `}`}, {"Sts", `{"n1":x` + geheim + `}`},
	}
	for _, f := range faelle {
		for _, art := range eArten() {
			_, err := eA(f.set, art, f.body).los()
			if err == nil {
				t.Errorf("%s: kein Fehler", f.body)
				continue
			}
			if strings.Contains(err.Error(), "GEHEIM-7") {
				t.Errorf("%s: Fehlertext trägt einen Wert des Bodys: %s", f.body, err)
			}
		}
	}
}
