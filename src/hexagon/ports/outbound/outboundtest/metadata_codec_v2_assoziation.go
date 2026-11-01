package outboundtest

import (
	"fmt"
	"slices"
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

func faelleV2N3() []fall {
	var out []fall
	ends := func(m0, m1 string) string {
		return `<Association Name="R"><End Type="Ns.A" Role="F" Multiplicity="` + m0 + `"/><End Type="Ns.B" Role="G" Multiplicity="` + m1 + `"/></Association>`
	}
	for _, verweis := range []bool{false, true} {
		nav := ""
		if verweis {
			nav = navRV2
		}
		for _, v := range []string{"", "0..*", "1..1", "*..1", "* ", "1.", "0..1 ", " 1", "2", "0", "many", "0..1..*"} {
			out = append(out, schlecht(fmt.Sprintf("erstes End %q, Verweis %v", v, verweis), baseV2(nav, "", ends(v, "1"), ""), model.UrsacheStruktur, "Ns.R@Multiplicity"),
				schlecht(fmt.Sprintf("zweites End %q, Verweis %v", v, verweis), baseV2(nav, "", ends("1", v), ""), model.UrsacheStruktur, "Ns.R@Multiplicity"))
		}
		for _, v := range []string{"0..1", "1", "*"} {
			out = append(out, gut(fmt.Sprintf("Multiplicity %s, Verweis %v", v, verweis), baseV2(nav, "", ends(v, v), "")))
		}
	}
	return out
}

func zeileV2N3(t *testing.T, neu CodecNeu) { pruefeFaelle(t, neu, faelleV2N3()) }

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

// faelleV2N fasst die Fälle der Assoziationen zusammen, die A4, A5 und die Allaussagen mitlaufen.
func faelleV2N() []fall {
	out := slices.Concat(faelleV2N1(), faelleV2N2(), faelleV2N3(), faelleV2N5(), faelleV2N6(), faelleV2N7())
	return append(out, gut("Bindings A vor B", bindungenV2(false)), gut("Bindings B vor A", bindungenV2(true)))
}
