package outboundtest

import (
	"errors"
	"fmt"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// assocRV2 ist die Association `R` von `Ns.A` (Rolle F, Vielfachheit *) nach `Ns.B` (Rolle G, Vielfachheit 0..1).
const assocRV2 = `<Association Name="R"><End Type="Ns.A" Role="F" Multiplicity="*"/><End Type="Ns.B" Role="G" Multiplicity="0..1"/></Association>`

// navRV2 ist eine Navigation-Property N von F nach G auf `Ns.R`.
const navRV2 = `<NavigationProperty Name="N" Relationship="Ns.R" FromRole="F" ToRole="G"/>`

// asetV2 ist eine AssociationSet mit den EntitySets ef (Rolle F) und eg (Rolle G).
func asetV2(name, assoc, ef, eg string) string {
	return `<AssociationSet Name="` + name + `" Association="` + assoc + `"><End Role="F" EntitySet="` + ef + `"/><End Role="G" EntitySet="` + eg + `"/></AssociationSet>`
}

// setV2 ist ein EntitySet.
func setV2(name, typ string) string {
	return `<EntitySet Name="` + name + `" EntityType="` + typ + `"/>`
}

// containerV2 ist der Container `C` mit den Kindern inhalt.
func containerV2(inhalt string) string {
	return `<EntityContainer Name="C">` + inhalt + `</EntityContainer>`
}

// baseV2 ist ein Dokument mit den Typen `A` und `B` (Alias `Self`, ComplexType `Cx`), deren Navigation-Properties navA und navB.
func baseV2(navA, navB, assoc, container string) []byte {
	return []byte(edmx2(schemaAlias2("Ns", "Self", `<ComplexType Name="Cx"/>`+typ("A", navA)+typ("B", navB)+assoc+container)))
}

// navFormV2 ist eine Form der Navigation-Property und der Association von N1.
type navFormV2 struct {
	name, rel, von, nach, typF, typG string
	ok                               bool
}

func navFormenV2() []navFormV2 {
	return []navFormV2{
		{"gültig", "Ns.R", "F", "G", "Ns.A", "Ns.B", true}, {"gültig über Alias", "Self.R", "F", "G", "Self.A", "Self.B", true},
		{"Typ des F-End unauflösbar", "Ns.R", "F", "G", "Ns.Nope", "Ns.B", true}, {"Typ des F-End ein ComplexType", "Ns.R", "F", "G", "Ns.Cx", "Ns.B", true},
		{"Association fehlt", "Ns.Nope", "F", "G", "Ns.A", "Ns.B", false}, {"Alias unbekannt", "Zzz.R", "F", "G", "Ns.A", "Ns.B", false},
		{"ohne Punkt", "R", "F", "G", "Ns.A", "Ns.B", false}, {"Leerraum vorn", " Ns.R", "F", "G", "Ns.A", "Ns.B", false},
		{"Leerraum hinten", "Ns.R ", "F", "G", "Ns.A", "Ns.B", false}, {"Schreibweise des Namensraums", "ns.R", "F", "G", "Ns.A", "Ns.B", false},
		{"Schreibweise des Namens", "Ns.r", "F", "G", "Ns.A", "Ns.B", false}, {"ohne Rolle F", "Ns.R", "X", "G", "Ns.A", "Ns.B", false},
		{"ohne Rolle G", "Ns.R", "F", "X", "Ns.A", "Ns.B", false}, {"F gleich G", "Ns.R", "F", "F", "Ns.A", "Ns.B", false},
		{"Schreibweise der Rolle", "Ns.R", "f", "G", "Ns.A", "Ns.B", false}, {"Typ G nicht auflösbar", "Ns.R", "F", "G", "Ns.A", "Ns.Nope", false},
		{"Typ G ein ComplexType", "Ns.R", "F", "G", "Ns.A", "Ns.Cx", false}, {"Typ G mit Leerraum", "Ns.R", "F", "G", "Ns.A", " Ns.B", false},
	}
}

// navDokumentV2 trägt die Navigation-Property N an A (erster Typ) oder an D, der Ableitung von A (abgeleiteter Typ).
func navDokumentV2(f navFormV2, abgeleitet bool) []byte {
	n := `<NavigationProperty Name="N" Relationship="` + f.rel + `" FromRole="` + f.von + `" ToRole="` + f.nach + `"/>`
	assoc := `<Association Name="R"><End Type="` + f.typF + `" Role="F" Multiplicity="*"/><End Type="` + f.typG + `" Role="G" Multiplicity="0..1"/></Association>`
	if abgeleitet {
		return []byte(edmx2(schemaAlias2("Ns", "Self", `<ComplexType Name="Cx"/>`+typ("A", "")+typ("B", "")+ableitung("D", "Ns.A", n)+assoc)))
	}
	return baseV2(n, "", assoc, "")
}

func faelleV2N1() []fall {
	var out []fall
	for _, f := range navFormenV2() {
		for _, abgeleitet := range []bool{false, true} {
			stelle, tn := "erster Typ", "Ns.A"
			if abgeleitet {
				stelle, tn = "abgeleiteter Typ", "Ns.D"
			}
			name := f.name + ", " + stelle
			if f.ok {
				out = append(out, gut(name, navDokumentV2(f, abgeleitet)))
			} else {
				out = append(out, schlecht(name, navDokumentV2(f, abgeleitet), model.UrsacheNichtAufloesbar, tn+"/N@Relationship"))
			}
		}
	}
	return out
}

// zeileV2N1: eine Navigation-Property löst über Relationship, FromRole, ToRole und den Typ des End der Rolle ToRole auf
// (SPEC-171 Nr. 6): Schleife über die Formen mal erster und abgeleiteter Typ.
func zeileV2N1(t *testing.T, neu CodecNeu) { pruefeFaelle(t, neu, faelleV2N1()) }

func faelleV2N2() []fall {
	var out []fall
	for _, mult := range []string{"0..1", "1", "*"} {
		for _, ref := range []string{"Self.B", "Ns.B"} {
			for _, mf := range []string{"*", "1"} {
				doc := baseV2(navRV2, "", `<Association Name="R"><End Type="Ns.A" Role="F" Multiplicity="`+mf+`"/><End Type="`+ref+`" Role="G" Multiplicity="`+mult+`"/></Association>`, "")
				out = append(out, gut(fmt.Sprintf("G-End %s über %s, F-End %s", mult, ref, mf), doc))
			}
		}
	}
	return out
}

// zeileV2N2: Target ist der qualifizierte Name des Typs am End der Rolle ToRole, Collection wahr genau bei Multiplicity * dieses End.
func zeileV2N2(t *testing.T, neu CodecNeu) {
	faelle := faelleV2N2()
	pruefeFaelle(t, neu, faelle)
	for i, mult := range []string{"0..1", "1", "*"} {
		for j := range 4 {
			m, _ := pruefeFall(t, neu, faelle[i*4+j])
			a, _ := m.EntityType("Ns.A")
			soll := []model.NavigationProperty{{Name: "N", Target: "Ns.B", Collection: mult == "*"}}
			if !slices.Equal(a.NavigationProperties, soll) {
				t.Errorf("%s: %+v, erwartet %+v", faelle[i*4+j].name, a.NavigationProperties, soll)
			}
		}
	}
}

// multiplicitaetenV2 sind die drei gültigen Werte von Multiplicity (SPEC-171 Nr. 10 d).
func multiplicitaetenV2() []string { return []string{"0..1", "1", "*"} }

// n3Zeichenketten sind alle Zeichenketten bis zur Länge 4 aus dem Alphabet `0 1 2 * .` und Leerzeichen, die leere
// eingeschlossen: 1 + 6 + 36 + 216 + 1296 = 1555.
func n3Zeichenketten() []string {
	alphabet := []string{"0", "1", "2", "*", ".", " "}
	out, stufe := []string{""}, []string{""}
	for range 4 {
		var naechste []string
		for _, s := range stufe {
			for _, a := range alphabet {
				naechste = append(naechste, s+a)
			}
		}
		out, stufe = append(out, naechste...), naechste
	}
	return out
}

// n3Werte sind die ungültigen Werte von Multiplicity der Schleife von N3: alle Zeichenketten von n3Zeichenketten ohne die
// drei gültigen (1552), dazu `many`, `0..1..*` und jeder gültige Wert mit Tab oder NBSP (Bytefolge C2 A0) davor oder
// dahinter (12).
func n3Werte() []string {
	out := slices.DeleteFunc(n3Zeichenketten(), func(s string) bool { return slices.Contains(multiplicitaetenV2(), s) })
	zusatz := []string{"many", "0..1..*"}
	for _, v := range multiplicitaetenV2() {
		for _, leer := range []string{"\t", "\xc2\xa0"} {
			zusatz = append(zusatz, leer+v, v+leer)
		}
	}
	return append(out, zusatz...)
}

// n3Dokument ist das Dokument mit der Association `R`, deren End die Vielfachheiten m0 und m1 tragen, mit dem Verweis
// einer Navigation oder ohne.
func n3Dokument(m0, m1 string, verweis bool) []byte {
	nav := ""
	if verweis {
		nav = navRV2
	}
	return baseV2(nav, "", `<Association Name="R"><End Type="Ns.A" Role="F" Multiplicity="`+m0+`"/><End Type="Ns.B" Role="G" Multiplicity="`+m1+`"/></Association>`, "")
}

// faelleV2N3Klein sind die benannten Werte von N3 und die gültigen; sie laufen in allen Zeilen, die über alle Dokumente
// der Suite gehen. Die Schleife über n3Werte läuft allein in N3.
func faelleV2N3Klein() []fall {
	var out []fall
	for _, verweis := range []bool{false, true} {
		for _, v := range []string{"", "0..*", "1..1", "*..1", "* ", "1.", "0..1 ", " 1", "2", "0", "many", "0..1..*"} {
			out = append(out, schlecht(fmt.Sprintf("erstes End %q, Verweis %v", v, verweis), n3Dokument(v, "1", verweis), model.UrsacheStruktur, "Ns.R@Multiplicity"),
				schlecht(fmt.Sprintf("zweites End %q, Verweis %v", v, verweis), n3Dokument("1", v, verweis), model.UrsacheStruktur, "Ns.R@Multiplicity"))
		}
		for _, v := range multiplicitaetenV2() {
			out = append(out, gut(fmt.Sprintf("Multiplicity %s, Verweis %v", v, verweis), n3Dokument(v, v, verweis)))
		}
	}
	return out
}

// faelleV2N3 sind die Dokumente der Schleife über n3Werte: jeder Wert mal Association (referenziert oder nicht) mal
// erstes und zweites End.
func faelleV2N3() []fall {
	var out []fall
	for _, v := range n3Werte() {
		for _, verweis := range []bool{false, true} {
			out = append(out, schlecht(fmt.Sprintf("erstes End %q, Verweis %v", v, verweis), n3Dokument(v, "1", verweis), model.UrsacheStruktur, "Ns.R@Multiplicity"),
				schlecht(fmt.Sprintf("zweites End %q, Verweis %v", v, verweis), n3Dokument("1", v, verweis), model.UrsacheStruktur, "Ns.R@Multiplicity"))
		}
	}
	return out
}

// pruefeKompaktV2 fährt Fälle wie pruefeFall, ohne je Fall einen Unterlauf anzulegen; nach 20 Abweichungen bricht es ab.
func pruefeKompaktV2(t *testing.T, neu CodecNeu, faelle []fall) {
	t.Helper()
	abweichungen := 0
	for _, f := range faelle {
		m, b, err := lies(neu, f)
		var mf *model.MetadatenFehler
		switch {
		case f.ursache == "" && err != nil:
			t.Errorf("%s: Fehler %v bei gültigem Dokument", f.name, err)
		case f.ursache == "":
			continue
		case !errors.As(err, &mf) || mf.Ursache != f.ursache || mf.Ort != f.ort || mf.Code() != model.CodeE6004 || !errors.Is(err, model.ErrMetadatenUngueltig):
			t.Errorf("%s: Fehler %v, erwartet %s %q", f.name, err, f.ursache, f.ort)
		case m.Version() != "" || len(m.EntityTypes()) != 0 || b.Funde != nil:
			t.Errorf("%s: Modell und Bericht sind keine Nullwerte", f.name)
		default:
			continue
		}
		if abweichungen++; abweichungen >= 20 {
			t.Fatalf("%d Abweichungen, Abbruch", abweichungen)
		}
	}
}

// zeileV2N3: Multiplicity ist `0..1`, `1` oder `*`, jede andere Zeichenkette ist struktur am Ort der Association (SPEC-171
// Nr. 10 d): Schleife über alle Zeichenketten bis Länge 4 aus dem Alphabet `0 1 2 * .` und Leerzeichen (1555, ohne die
// drei gültigen 1552), über die benannten Formen, mal Association referenziert oder nicht, mal erstes und zweites End.
// Die Regel ist eine Liste mit drei Werten; Zeichenketten ab Länge 5 und andere Zeichen prüft die Schleife nicht
// (Negativ im Plan).
func zeileV2N3(t *testing.T, neu CodecNeu) {
	if n := len(n3Zeichenketten()); n != 1555 {
		t.Fatalf("%d Zeichenketten bis Länge 4, erwartet 1555", n)
	}
	pruefeFaelle(t, neu, faelleV2N3Klein())
	pruefeKompaktV2(t, neu, faelleV2N3())
}

// bindungenV2 ist das Dokument von N4: zwei Navigation-Properties an A, eine an B, zwei AssociationSet je Association in
// gemischter Reihenfolge; erst ist der Typ, der zuerst im Dokument steht.
func bindungenV2(bZuerst bool) []byte {
	container := containerV2(asetV2("AS2", "Ns.R", "SA2", "SB2") + setV2("SA1", "Ns.A") + asetV2("AS1", "Self.R", "SA1", "SB1") + setV2("SA2", "Ns.A") +
		setV2("SB1", "Ns.B") + setV2("SB2", "Ns.B"))
	a := `<NavigationProperty Name="n1" Relationship="Ns.R" FromRole="F" ToRole="G"/><NavigationProperty Name="n2" Relationship="Ns.R" FromRole="F" ToRole="G"/>`
	b := `<NavigationProperty Name="m1" Relationship="Ns.R" FromRole="G" ToRole="F"/>`
	if bZuerst {
		return []byte(edmx2(schemaAlias2("Ns", "Self", typ("B", b)+typ("A", a)+assocRV2+container)))
	}
	return baseV2(a, b, assocRV2, container)
}

// zeileV2N4: zu jeder AssociationSet und jeder Navigation-Property mit gleicher Relationship ein Binding, in der
// Reihenfolge der AssociationSet im Dokument, darin die der Navigation-Properties (SPEC-171 Nr. 6).
func zeileV2N4(t *testing.T, neu CodecNeu) {
	bind := func(set, pfad, ziel string) model.NavigationBinding {
		return model.NavigationBinding{Set: set, Path: pfad, Target: "Ns.C/" + ziel}
	}
	sets := []model.EntitySet{{Name: "SA1", EntityType: "Ns.A"}, {Name: "SA2", EntityType: "Ns.A"}, {Name: "SB1", EntityType: "Ns.B"}, {Name: "SB2", EntityType: "Ns.B"}}
	for _, f := range []struct {
		name     string
		bZuerst  bool
		bindings []model.NavigationBinding
	}{
		{"A vor B", false, []model.NavigationBinding{
			bind("SA2", "n1", "SB2"), bind("SA2", "n2", "SB2"), bind("SB2", "m1", "SA2"), bind("SA1", "n1", "SB1"), bind("SA1", "n2", "SB1"), bind("SB1", "m1", "SA1")}},
		{"B vor A", true, []model.NavigationBinding{
			bind("SB2", "m1", "SA2"), bind("SA2", "n1", "SB2"), bind("SA2", "n2", "SB2"), bind("SB1", "m1", "SA1"), bind("SA1", "n1", "SB1"), bind("SA1", "n2", "SB1")}},
	} {
		m, _ := pruefeFall(t, neu, gut(f.name, bindungenV2(f.bZuerst)))
		gleich(t, f.name+" Bindings", m.NavigationBindings(), f.bindings)
		gleich(t, f.name+" Sets", m.EntitySets(), sets)
	}
}

func faelleV2N5() []fall {
	return []fall{
		gut("Navigation ohne AssociationSet", baseV2(navRV2, "", assocRV2, containerV2(setV2("SA", "Ns.A")+setV2("SB", "Ns.B")))),
		gut("Navigation ohne Container", baseV2(navRV2, "", assocRV2, "")),
		gut("AssociationSet ohne Navigation", baseV2("", "", assocRV2, containerV2(setV2("SA", "Ns.A")+setV2("SB", "Ns.B")+asetV2("AS", "Ns.R", "SA", "SB")))),
		gut("AssociationSet ohne Navigation, Rollen der Association nicht getragen", baseV2("", "", assocRV2, containerV2(setV2("SA", "Ns.A")+
			`<AssociationSet Name="AS" Association="Ns.R"><End Role="X" EntitySet="SA"/><End Role="Y" EntitySet="SA"/></AssociationSet>`))),
		gut("Association ohne Navigation und ohne AssociationSet", baseV2("", "", assocRV2, "")),
		gut("Container ohne Kinder", baseV2(navRV2, "", assocRV2, containerV2(""))),
	}
}

// zeileV2N5: eine Navigation ohne AssociationSet bleibt im Modell ohne Binding, eine AssociationSet ohne Navigation
// erzeugt keines (SPEC-171 Nr. 6).
func zeileV2N5(t *testing.T, neu CodecNeu) {
	faelle := faelleV2N5()
	pruefeFaelle(t, neu, faelle)
	for i, f := range faelle {
		m, _ := pruefeFall(t, neu, f)
		a, _ := m.EntityType("Ns.A")
		navs := 1
		if i == 2 || i == 3 || i == 4 {
			navs = 0
		}
		if len(m.NavigationBindings()) != 0 || len(a.NavigationProperties) != navs {
			t.Errorf("%s: Bindings %v, Navigationen %v", f.name, m.NavigationBindings(), a.NavigationProperties)
		}
	}
}

func faelleV2N6() []fall {
	ort := "Ns.C/AS"
	nA := model.UrsacheNichtAufloesbar
	mit := func(aset string) []byte {
		return baseV2(navRV2, "", assocRV2, containerV2(setV2("SA", "Ns.A")+setV2("SB", "Ns.B")+aset))
	}
	var out []fall
	for _, a := range []string{"Ns.Nope", "Zzz.R", "R", " Ns.R", "Ns.R ", "ns.R"} {
		out = append(out, schlecht("Association "+a, mit(asetV2("AS", a, "SA", "SB")), nA, ort))
	}
	for _, rolle := range []string{"F", "G"} {
		falsch := func(r string) string {
			if rolle == "F" {
				return `<AssociationSet Name="AS" Association="Ns.R"><End Role="` + r + `" EntitySet="SA"/><End Role="G" EntitySet="SB"/></AssociationSet>`
			}
			return `<AssociationSet Name="AS" Association="Ns.R"><End Role="F" EntitySet="SA"/><End Role="` + r + `" EntitySet="SB"/></AssociationSet>`
		}
		for _, r := range []string{"X", "f", "g", " F", "F "} {
			out = append(out, schlecht(fmt.Sprintf("End %s mit Rolle %q", rolle, r), mit(falsch(r)), nA, ort))
		}
		for _, s := range []string{"Nope", "SA ", "sa", "Ns.C/SA", ""} {
			ef, eg := "SA", "SB"
			if rolle == "F" {
				ef = s
			} else {
				eg = s
			}
			if s == "" {
				out = append(out, schlecht(fmt.Sprintf("End %s mit EntitySet leer", rolle), mit(asetV2("AS", "Ns.R", ef, eg)), model.UrsacheStruktur, ort+"@EntitySet"))
				continue
			}
			out = append(out, schlecht(fmt.Sprintf("End %s mit EntitySet %q", rolle, s), mit(asetV2("AS", "Ns.R", ef, eg)), nA, ort))
		}
	}
	setOhneZiel, asetOhneZiel := setV2("SX", "Ns.Nope"), asetV2("AS", "Ns.Nope", "SA", "SB")
	return append(out,
		schlecht("Set ohne Ziel vor AssociationSet ohne Ziel", mit(setOhneZiel+asetOhneZiel), nA, "Ns.C/SX@EntityType"),
		schlecht("AssociationSet ohne Ziel vor Set ohne Ziel", mit(asetOhneZiel+setOhneZiel), nA, ort),
		gut("zweite AssociationSet mit anderen Sets", mit(asetV2("AS", "Ns.R", "SA", "SB")+asetV2("AS2", "Ns.R", "SB", "SA"))))
}

// zeileV2N6: eine AssociationSet ohne Association, mit Rollen, die die Navigation nicht trägt, oder mit einem EntitySet,
// das der Container nicht hat, ist nicht_aufloesbar am Ort der AssociationSet (SPEC-171 Nr. 7 e, 10 f).
func zeileV2N6(t *testing.T, neu CodecNeu) {
	faelle := faelleV2N6()
	pruefeFaelle(t, neu, faelle)
	m, _ := pruefeFall(t, neu, faelle[len(faelle)-1])
	if len(m.NavigationBindings()) != 2 {
		t.Errorf("zwei AssociationSet einer Association: %v, erwartet je ein Binding", m.NavigationBindings())
	}
}

func faelleV2N7() []fall {
	struk := model.UrsacheStruktur
	var out []fall
	endA := func(r string) string {
		return `<End Type="Ns.A" Role="` + r + `" Multiplicity="*"/>`
	}
	endS := func(r string) string { return `<End Role="` + r + `" EntitySet="SA"/>` }
	for _, rollen := range [][]string{{}, {"F"}, {"F", "G", "H"}, {"F", "G", "H", "I"}} {
		var a, s string
		for _, r := range rollen {
			a += endA(r)
			s += endS(r)
		}
		assoc := `<Association Name="R">` + a + `</Association>`
		aset := `<AssociationSet Name="AS" Association="Ns.R">` + s + `</AssociationSet>`
		name := fmt.Sprintf("%d End", len(rollen))
		out = append(out, schlecht(name+" in der Association", baseV2("", "", assoc, ""), struk, "Ns.R"),
			schlecht(name+" in der AssociationSet", baseV2("", "", assocRV2, containerV2(setV2("SA", "Ns.A")+aset)), struk, "Ns.C/AS"))
	}
	out = append(out,
		schlecht("gleiche Rolle in der Association", baseV2("", "", `<Association Name="R">`+endA("F")+endA("F")+`</Association>`, ""), struk, "Ns.R@Role"),
		schlecht("gleiche Rolle in der AssociationSet", baseV2("", "", assocRV2, containerV2(setV2("SA", "Ns.A")+`<AssociationSet Name="AS" Association="Ns.R">`+endS("F")+endS("F")+`</AssociationSet>`)), struk, "Ns.C/AS@Role"),
		gut("zwei End mit verschiedenen Rollen", baseV2("", "", assocRV2, "")))
	return out
}

// zeileV2N7: eine Association und eine AssociationSet haben genau zwei End mit verschiedener Rolle (SPEC-171 Nr. 10 b, c).
func zeileV2N7(t *testing.T, neu CodecNeu) { pruefeFaelle(t, neu, faelleV2N7()) }

// n8Gruppe sind die Navigation-Properties und AssociationSet einer Association `Ns.<name>` des Dokuments von N8: navs
// Navigationen am Typ `T` und sets AssociationSet; davor und danach stehen AssociationSet vor und hinter den erzeugten.
// ohneSchluessel macht `T` abstrakt und ohne Schlüssel: alle 65 536 Members, die das Dokument tragen darf (Nr. 9 b), sind
// dann Navigationen.
type n8Gruppe struct {
	name           string
	sets, navs     int
	davor, nach    string
	ohneSchluessel bool
}

// n8Dokument ist ein Dokument mit dem Typ `T`, dem EntitySet `S` und je Gruppe einer Association (Rollen F und G, beide
// End vom Typ `T`), den Navigationen `n<Gruppe>_<i>` am Typ und den AssociationSet `A<Gruppe>_<i>` im Container, in der
// Reihenfolge der Gruppen. Die Gruppe `R<k>` trägt sets mal navs Bindings.
func n8Dokument(gruppen []n8Gruppe) []byte {
	var navs, assocs, asets strings.Builder
	t := typ
	for _, g := range gruppen {
		if g.ohneSchluessel {
			t = func(name, inhalt string) string {
				return `<EntityType Name="` + name + `" Abstract="true">` + inhalt + `</EntityType>`
			}
		}
		assocs.WriteString(`<Association Name="` + g.name + `"><End Type="Ns.T" Role="F" Multiplicity="*"/><End Type="Ns.T" Role="G" Multiplicity="0..1"/></Association>`)
		for i := range g.navs {
			navs.WriteString(`<NavigationProperty Name="n` + g.name + fmt.Sprintf("_%d", i) + `" Relationship="Ns.` + g.name + `" FromRole="F" ToRole="G"/>`)
		}
		asets.WriteString(g.davor)
		for i := range g.sets {
			asets.WriteString(asetV2(fmt.Sprintf("A%s_%d", g.name, i), "Ns."+g.name, "S", "S"))
		}
		asets.WriteString(g.nach)
	}
	return []byte(edmx2(schema2("Ns", t("T", navs.String())+assocs.String()+containerV2(setV2("S", "Ns.T")+asets.String()))))
}

// n8Fall ist ein Dokument von N8; bindings ist die erwartete Zahl bei einem gültigen Dokument, speicher verlangt, dass
// der Aufruf höchstens n8Speicher Bytes zuweist.
type n8Fall struct {
	fall
	bindings int
	speicher bool
}

// n8Speicher ist die Obergrenze der Zuweisungen eines Aufrufs mit einem Dokument von rund 450 KiB, das ohne Grenze
// vier Millionen Bindings (rund 400 MiB) ergäbe.
const n8Speicher = 64 << 20

// faelleV2N8 sind die Fälle der Grenze der Bindings (SPEC-171 Nr. 6, 7). `unaufloesbar` ist eine AssociationSet ohne Association.
func faelleV2N8() []n8Fall {
	unaufloesbar := asetV2("BAD", "Ns.Nope", "S", "S")
	zuViele := func(name, ort string, gruppen ...n8Gruppe) n8Fall {
		return n8Fall{fall: schlecht(name, n8Dokument(gruppen), model.UrsacheZuVieleMembers, ort)}
	}
	nichtAufl := func(name, ort string, gruppen ...n8Gruppe) n8Fall {
		return n8Fall{fall: schlecht(name, n8Dokument(gruppen), model.UrsacheNichtAufloesbar, ort)}
	}
	gueltig := func(name string, n int, gruppen ...n8Gruppe) n8Fall {
		return n8Fall{fall: gut(name, n8Dokument(gruppen)), bindings: n}
	}
	gross := zuViele("2000 AssociationSet mal 2000 Navigationen", "Ns.C/AR0_32", n8Gruppe{name: "R0", sets: 2000, navs: 2000})
	gross.speicher = true
	return []n8Fall{
		gueltig("genau 65536: 256 AssociationSet mal 256 Navigationen", 65536, n8Gruppe{name: "R0", sets: 256, navs: 256}),
		zuViele("65537: eine zweite Association zuletzt", "Ns.C/AR1_0", n8Gruppe{name: "R0", sets: 256, navs: 256}, n8Gruppe{name: "R1", sets: 1, navs: 1}),
		zuViele("65537: die zweite Association zuerst, Ort die überschreitende AssociationSet", "Ns.C/AR1_255", n8Gruppe{name: "R0", sets: 1, navs: 1}, n8Gruppe{name: "R1", sets: 256, navs: 256}),
		gross,
		nichtAufl("unaufloesbare AssociationSet vor der überschreitenden", "Ns.C/BAD", n8Gruppe{name: "R0", sets: 256, navs: 256}, n8Gruppe{name: "R1", sets: 1, navs: 1, davor: unaufloesbar}),
		zuViele("unaufloesbare AssociationSet hinter der überschreitenden", "Ns.C/AR1_0", n8Gruppe{name: "R0", sets: 256, navs: 256}, n8Gruppe{name: "R1", sets: 1, navs: 1, nach: unaufloesbar}),
		nichtAufl("überschreitende AssociationSet mit Rolle ohne Navigation", "Ns.C/AR1_0", n8Gruppe{name: "R0", sets: 256, navs: 256},
			n8Gruppe{name: "R1", navs: 1, davor: `<AssociationSet Name="AR1_0" Association="Ns.R1"><End Role="X" EntitySet="S"/><End Role="G" EntitySet="S"/></AssociationSet>`}),
		gueltig("Summe über zwei Associations genau 65536", 65536, n8Gruppe{name: "R0", sets: 128, navs: 256}, n8Gruppe{name: "R1", sets: 128, navs: 256}),
		zuViele("Summe über drei Associations 65537", "Ns.C/AR2_0", n8Gruppe{name: "R0", sets: 128, navs: 256}, n8Gruppe{name: "R1", sets: 128, navs: 256}, n8Gruppe{name: "R2", sets: 1, navs: 1}),
		gueltig("eine AssociationSet allein mit genau 65536 Navigationen", 65536, n8Gruppe{name: "R0", sets: 1, navs: 65536, ohneSchluessel: true}),
		zuViele("dieselbe Association in einer zweiten Set, Ort die zweite Set", "Ns.C/AR0_1", n8Gruppe{name: "R0", sets: 2, navs: 65536, ohneSchluessel: true}),
	}
}

// zeileV2N8: höchstens 65 536 Bindings, die 65 537. ist zu_viele_properties am Ort der überschreitenden AssociationSet,
// gezählt über alle Associations und vor dem Bauen; je AssociationSet erst ihre Auflösung, dann die Grenze (SPEC-171 Nr. 6,
// 7): sechs Fälle, die nach der Zahl der Bindings, der Reihenfolge und der Auflösung schneiden, dazu eine einzige
// AssociationSet mit allen 65 536 Navigationen, die ein Dokument tragen darf (gültig), und ihre zweite Set (elf Dokumente).
func zeileV2N8(t *testing.T, neu CodecNeu) {
	faelle := faelleV2N8()
	for _, f := range faelle {
		var vorher, nachher runtime.MemStats
		runtime.ReadMemStats(&vorher)
		m, _ := pruefeFall(t, neu, f.fall)
		runtime.ReadMemStats(&nachher)
		if f.ursache == "" && len(m.NavigationBindings()) != f.bindings {
			t.Errorf("%s: %d Bindings, erwartet %d", f.name, len(m.NavigationBindings()), f.bindings)
		}
		if zugewiesen := nachher.TotalAlloc - vorher.TotalAlloc; f.speicher && zugewiesen > n8Speicher {
			t.Errorf("%s: %d MiB zugewiesen, erwartet höchstens %d", f.name, zugewiesen>>20, n8Speicher>>20)
		}
	}
}

// faelleV2N fasst die Fälle der Assoziationen zusammen, die A4, A5 und die Allaussagen mitlaufen.
func faelleV2N() []fall {
	out := slices.Concat(faelleV2N1(), faelleV2N2(), faelleV2N3Klein(), faelleV2N5(), faelleV2N6(), faelleV2N7())
	return append(out, gut("Bindings A vor B", bindungenV2(false)), gut("Bindings B vor A", bindungenV2(true)))
}
