package urlparser_test

import (
	"testing"
	"unicode/utf8"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// zeichenKlasse ist die Rolle eines Zeichens in einem Namen (SPEC-161 Nr. 4): Buchstabe (Kategorie L), Ziffer (Kategorie Nd) oder
// keines von beiden.
type zeichenKlasse int

const (
	keines zeichenKlasse = iota
	buchstabe
	ziffer
)

// unicodeZeichen ist ein Zeichen der Menge Z mit seinem Codepunkt, seinen UTF-8-Bytes (als Bytewerte, §3.18) und der Klasse von Hand.
type unicodeZeichen struct {
	cp     rune
	bytes  string
	klasse zeichenKlasse
	name   string
}

// unicodeMenge ist die Menge Z (29 Zeichen): 6 Buchstaben (Lu, Ll, Lt, Lm, Lo und ein Lu außerhalb der BMP), 4 Ziffern (Nd), 3 Nl,
// 4 No, 3 M (Mn, Mc, Me) und 9 sonstige (Pc, Zs, Cf, Sm, So, Sk, Pd, Po und ein So mit der Eigenschaft Alphabetic).
func unicodeMenge() []unicodeZeichen {
	return []unicodeZeichen{
		{0x00C4, "\xc3\x84", buchstabe, "U+00C4 Lu"}, {0x00E4, "\xc3\xa4", buchstabe, "U+00E4 Ll"}, {0x01C5, "\xc7\x85", buchstabe, "U+01C5 Lt"},
		{0x02B0, "\xca\xb0", buchstabe, "U+02B0 Lm"}, {0x3042, "\xe3\x81\x82", buchstabe, "U+3042 Lo"},
		{0x1D400, "\xf0\x9d\x90\x80", buchstabe, "U+1D400 Lu"},
		{0x0663, "\xd9\xa3", ziffer, "U+0663 Nd"}, {0x0969, "\xe0\xa5\xa9", ziffer, "U+0969 Nd"}, {0xFF13, "\xef\xbc\x93", ziffer, "U+FF13 Nd"},
		{0x1D7D9, "\xf0\x9d\x9f\x99", ziffer, "U+1D7D9 Nd"},
		{0x2167, "\xe2\x85\xa7", keines, "U+2167 Nl"}, {0x3007, "\xe3\x80\x87", keines, "U+3007 Nl"}, {0x16EE, "\xe1\x9b\xae", keines, "U+16EE Nl"},
		{0x00B2, "\xc2\xb2", keines, "U+00B2 No"}, {0x00BD, "\xc2\xbd", keines, "U+00BD No"}, {0x2460, "\xe2\x91\xa0", keines, "U+2460 No"},
		{0x2080, "\xe2\x82\x80", keines, "U+2080 No"},
		{0x0301, "\xcc\x81", keines, "U+0301 Mn"}, {0x093E, "\xe0\xa4\xbe", keines, "U+093E Mc"}, {0x20DD, "\xe2\x83\x9d", keines, "U+20DD Me"},
		{0x203F, "\xe2\x80\xbf", keines, "U+203F Pc"}, {0x00A0, "\xc2\xa0", keines, "U+00A0 Zs"}, {0x200D, "\xe2\x80\x8d", keines, "U+200D Cf"},
		{0x00D7, "\xc3\x97", keines, "U+00D7 Sm"}, {0x00A9, "\xc2\xa9", keines, "U+00A9 So"}, {0x00A8, "\xc2\xa8", keines, "U+00A8 Sk"},
		{0x2013, "\xe2\x80\x93", keines, "U+2013 Pd"}, {0x00B7, "\xc2\xb7", keines, "U+00B7 Po"}, {0x24B6, "\xe2\x92\xb6", keines, "U+24B6 So"},
	}
}

// stellung ist die Lage von z im Muster: allein `z`, erste `z+"b"`, innere `"a"+z+"b"`, letzte `"a"+z`.
type stellung int

const (
	allein stellung = iota
	erste
	innere
	letzte
)

func (s stellung) muster(z string) string {
	switch s {
	case erste:
		return z + "b"
	case innere:
		return "a" + z + "b"
	case letzte:
		return "a" + z
	}
	return z
}

func (s stellung) String() string { return [...]string{"allein", "erste", "innere", "letzte"}[s] }

// unicodeElement ist ein Element der Menge: Ort, Zeichen, Stellung und der Fall mit der Erwartung von Hand.
type unicodeElement struct {
	ort string
	z   unicodeZeichen
	st  stellung
	f   fall
}

// nameBeginntMitZiffer meldet, ob das Muster mit einer Ziffer beginnt; kein Name beginnt mit einer Ziffer (SPEC-125 Nr. 8 b).
func nameBeginntMitZiffer(z unicodeZeichen, st stellung) bool {
	return z.klasse == ziffer && (st == allein || st == erste)
}

// keinName meldet, ob das Muster kein Name ist: das Zeichen ist weder Buchstabe noch Ziffer, oder eine Ziffer steht am Anfang.
func keinName(z unicodeZeichen, st stellung) bool {
	return z.klasse == keines || nameBeginntMitZiffer(z, st)
}

// praefixElement: `+'x'` vor `'x'` im Schlüssel; Buchstaben und Ziffern, auch an erster Stelle, sind `literal_typ`, jedes andere
// Zeichen macht das Literal zu `literal` (SPEC-161 Nr. 6 g).
func praefixElement(z unicodeZeichen, st stellung) fall {
	u := model.URLLiteral
	if z.klasse != keines {
		u = model.URLLiteralTyp
	}
	return fall{pfad: "/S(Id=" + st.muster(z.bytes) + "'x')", u: u, ort: "pfad/1"}
}

// setNameElement: der Name der Menge im Pfad; ein Zeichen, das keinen Namen trägt, macht das Segment zu `segment`.
func setNameElement(z unicodeZeichen, st stellung) fall {
	muster := st.muster(z.bytes)
	if keinName(z, st) {
		return fall{pfad: "/" + muster, u: model.URLSegment, ort: "pfad/1"}
	}
	return fall{pfad: "/" + muster, want: &model.ODataURL{Ressource: model.Ressource{Art: model.RessourceMenge, EntitySet: muster}}}
}

// schluesselNameElement: der Name eines Schlüsselteils; ein Zeichen, das keinen Namen trägt, macht den Körper zu `key`.
func schluesselNameElement(z unicodeZeichen, st stellung) fall {
	muster := st.muster(z.bytes)
	pfad := "/S(" + muster + "=1)"
	if keinName(z, st) {
		return fall{pfad: pfad, u: model.URLKey, ort: "pfad/1"}
	}
	return fall{pfad: pfad, want: entity("S", teil(muster, zahlLit("1")))}
}

// aliasElement: der Name eines Parameter-Alias; ein Zeichen, das keinen Namen trägt, macht den Optionsnamen zu `option_name`
// (SPEC-162 Nr. 3 d, f); ein Name steht in `Sonstige`.
func aliasElement(z unicodeZeichen, st stellung) fall {
	muster := st.muster(z.bytes)
	query := "@" + muster + "=1"
	if keinName(z, st) {
		return fall{pfad: "/S", query: query, u: model.URLOptionName, ort: "query/1"}
	}
	return fall{pfad: "/S", query: query, want: &model.ODataURL{
		Ressource: model.Ressource{Art: model.RessourceMenge, EntitySet: "S"},
		Optionen:  model.Optionen{Sonstige: []model.SonstigeOption{{Name: "@" + muster, Wert: "1"}}},
	}}
}

// unicodeElemente bildet die Menge: Zeichen × Ort × Stellung. Typpräfix und Name der Menge in vier Stellungen (116 je), Name eines
// Schlüsselteils und Alias in erster und innerer Stellung (58 je).
func unicodeElemente() []unicodeElement {
	orte := []struct {
		name       string
		stellungen []stellung
		bau        func(unicodeZeichen, stellung) fall
	}{
		{"Typpräfix", []stellung{allein, erste, innere, letzte}, praefixElement},
		{"Name der Menge", []stellung{allein, erste, innere, letzte}, setNameElement},
		{"Name eines Schlüsselteils", []stellung{erste, innere}, schluesselNameElement},
		{"Alias", []stellung{erste, innere}, aliasElement},
	}
	var els []unicodeElement
	for _, o := range orte {
		for _, z := range unicodeMenge() {
			for _, st := range o.stellungen {
				els = append(els, unicodeElement{o.name, z, st, o.bau(z, st)})
			}
		}
	}
	return els
}

func pruefeUnicodeElement(t *testing.T, e unicodeElement) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("%s %s %s: Panik %v", e.ort, e.z.name, e.st, r)
		}
	}()
	pruefeFall(t, parserNeu(), e.f)
}

// TestUnicodeKlassen (Zeile 28,SPEC-161 Nr. 4 und 6 g, SPEC-125 Nr. 8 b): ein Buchstabe ist ein Zeichen der Kategorie L, eine
// Ziffer eines der Kategorie Nd; Nl, No, M und die sonstigen Zeichen sind keines von beiden. Schleife über 29 Zeichen (10 positive,
// 19 negative) × Ort × Stellung: Typpräfix und Name der Menge in vier Stellungen (je 116), Name eines Schlüsselteils und Alias in
// erster und innerer Stellung (je 58), zusammen 348 Elemente. Die Bytes der Zeichen werden gegen ihren Codepunkt gehalten.
func TestUnicodeKlassen(t *testing.T) {
	zahl := map[zeichenKlasse]int{}
	for _, z := range unicodeMenge() {
		if r, n := utf8.DecodeRuneInString(z.bytes); r != z.cp || n != len(z.bytes) {
			t.Errorf("%s: Bytes %q sind U+%04X, erwartet U+%04X", z.name, z.bytes, r, z.cp)
		}
		zahl[z.klasse]++
	}
	if zahl[buchstabe] != 6 || zahl[ziffer] != 4 || zahl[keines] != 19 {
		t.Fatalf("Zeichen: %d Buchstaben, %d Ziffern, %d andere, erwartet 6, 4, 19", zahl[buchstabe], zahl[ziffer], zahl[keines])
	}
	els := unicodeElemente()
	for _, e := range els {
		pruefeUnicodeElement(t, e)
	}
	if len(els) != 348 {
		t.Fatalf("Zählprobe: %d Elemente, erwartet 348", len(els))
	}
}
