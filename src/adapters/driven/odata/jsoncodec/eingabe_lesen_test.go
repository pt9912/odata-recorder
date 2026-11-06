package jsoncodec_test

import (
	"errors"
	"io"
	"math"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

func errEOF() error { return io.EOF }

// eMitteLeer ist das Ergebnis von `{}` bei PATCH und der Fehler `fehlend` bei POST und PUT am Typ Mitte (Schlüssel Id).
func eMitteLeer(t *testing.T, was string, art model.EingabeArt, got model.EntityEingabe, err error) {
	t.Helper()
	if art == model.EingabePatch {
		eOK(t, was, got, err, model.EntityEingabe{Typ: "Ns.Mitte"})
		return
	}
	eFehler(t, was, got, err, model.EingabeFehlend, "/Id")
}

// TestEingabeLeerBody hält Zeile 9: fehlender und leerer Body sind `syntax`, Ort leer. Menge 18 = 6 Bodys × 3 Arten.
func TestEingabeLeerBody(t *testing.T) {
	bodys := []string{"", " ", "\t", "\n", "\r", strings.Repeat(" ", 4096)}
	n := 0
	for _, b := range bodys {
		for _, art := range eArten() {
			got, err := eA("Mitten", art, b).los()
			eFehler(t, "Body "+eQuote(b[:min(len(b), 3)]), got, err, model.EingabeSyntax, "")
			n++
		}
	}
	if n != 18 {
		t.Errorf("Menge %d, erwartet 18", n)
	}
}

// TestEingabeLeerraum hält SPEC-126 Nr. 4: Leerzeichen, Tabulator, LF und CR sind Leerraum vor, in und hinter dem Dokument.
func TestEingabeLeerraum(t *testing.T) {
	lr := " \t\n\r"
	for _, art := range eArten() {
		body := lr + "{" + lr + `"Id"` + lr + ":" + lr + "1" + lr + "}" + lr
		got, err := eA("Mitten", art, body).los()
		eOK(t, "Leerraum", got, err, model.EntityEingabe{Typ: "Ns.Mitte", Schluessel: "1", Properties: []model.PropertyWert{ePW("Id", model.ArtInt32, "1")}})
	}
}

// TestEingabeWurzel hält Zeile 10: die Wurzel ist ein Objekt, sonst `struktur` mit leerem Ort. Menge 21 = 6 × 3 + 3.
func TestEingabeWurzel(t *testing.T) {
	n := 0
	for _, w := range []string{"[]", `""`, "0", "true", "false", "null"} {
		for _, art := range eArten() {
			got, err := eA("Mitten", art, w).los()
			eFehler(t, "Wurzel "+w, got, err, model.EingabeStruktur, "")
			n++
		}
	}
	for _, art := range eArten() {
		got, err := eA("Mitten", art, "{}").los()
		eMitteLeer(t, "{}", art, got, err)
		n++
	}
	if n != 21 {
		t.Errorf("Menge %d, erwartet 21", n)
	}
}

// eReferenzBody ist der Referenz-Body von 60 Bytes der Zeile 11: ein Escape, ein Zeichen mit 2 und eines mit 4 Bytes, ein
// Wahrheitswert, eine Dezimalzahl, `null`.
func eReferenzBody() []byte {
	return []byte(`{"Id":123,"S":"a` + esc("00e9") + "\xc3\xa9\xf0\x9f\x98\x80" + `","B":true,"Dec":-1.50,"T":null}`)
}

func eReferenzErgebnis() model.EntityEingabe {
	return model.EntityEingabe{Typ: "Ns.W", Schluessel: "123", Properties: []model.PropertyWert{
		ePW("Id", model.ArtInt32, "123"), ePW("S", model.ArtString, "a\xc3\xa9\xc3\xa9\xf0\x9f\x98\x80"), ePW("B", model.ArtBoolean, "true"),
		ePW("Dec", model.ArtDecimal, "-1.5"), ePW("T", model.ArtNull, ""),
	}}
}

// TestEingabeZerlegung hält Zeile 11: das Ergebnis hängt nicht von der Zerlegung des Stroms ab, und `src` bekommt allein `Read`.
// Menge 183 = 61 Zerlegungen (ganz, einzelne Bytes, 59 Zweiteilungen) × 3 Arten, dazu leere Reads zwischen den Bytes × 3.
func TestEingabeZerlegung(t *testing.T) {
	body := eReferenzBody()
	if len(body) != 60 {
		t.Fatalf("Referenz-Body hat %d Bytes, erwartet 60", len(body))
	}
	var zerlegungen [][]int
	zerlegungen = append(zerlegungen, nil)
	einzeln := make([]int, len(body))
	for i := range einzeln {
		einzeln[i] = 1
	}
	zerlegungen = append(zerlegungen, einzeln)
	for i := 1; i < len(body); i++ {
		zerlegungen = append(zerlegungen, []int{i})
	}
	n := 0
	for _, art := range eArten() {
		for z, laengen := range zerlegungen {
			s := eStromGeteilt(body, laengen...)
			a := eA("Ws", art, "")
			a.src = s
			got, err := a.los()
			eOK(t, "Zerlegung "+eP(z), got, err, eReferenzErgebnis())
			s.verboten(t, "Zerlegung "+eP(z))
			n++
		}
		var schritte []eSchritt
		for i := range body {
			schritte = append(schritte, eSchritt{}, eSchritt{daten: body[i : i+1]})
		}
		s := eNeuerStrom(schritte...)
		a := eA("Ws", art, "")
		a.src = s
		got, err := a.los()
		eOK(t, "leere Reads", got, err, eReferenzErgebnis())
	}
	if n != 183 {
		t.Errorf("Menge %d, erwartet 183", n)
	}
}

// TestEingabeSrcFehler hält Zeile 11: ein Fehler von `src` bleibt unverändert (`==`), gleich wann er kommt, und der Verstoß des
// Lesens in den Bytes desselben Reads geht ihm vor. Menge 15 + 3.
func TestEingabeSrcFehler(t *testing.T) {
	quelle := errors.New("src kaputt")
	faelle := []struct {
		name     string
		schritte []eSchritt
		fehler   error
	}{
		{"vor dem ersten Byte", []eSchritt{{err: quelle}}, quelle},
		{"zwischen zwei Reads", []eSchritt{{daten: []byte(`{"Id"`)}, {err: quelle}}, quelle},
		{"mit Bytes im selben Read", []eSchritt{{daten: []byte(`{"Id":1`), err: quelle}}, quelle},
		{"nach dem vollständigen Dokument", []eSchritt{{daten: []byte(`{"Id":1}`)}, {err: quelle}}, quelle},
		{"ErrUnexpectedEOF", []eSchritt{{daten: []byte(`{"Id":1`)}, {err: io.ErrUnexpectedEOF}}, io.ErrUnexpectedEOF},
		{"DatenFehler von src", []eSchritt{{daten: []byte(`{"Id":1`)}, {err: &model.DatenFehler{Ursache: model.DatenSyntax}}}, nil},
	}
	n := 0
	for _, f := range faelle {
		for _, art := range eArten() {
			s := eNeuerStrom(f.schritte...)
			a := eA("Mitten", art, "")
			a.src = s
			got, err := a.los()
			if f.fehler != nil && err != f.fehler {
				t.Errorf("%s %s: Fehler %v, erwartet den Fehler von src unverändert", f.name, art, err)
			}
			var df *model.DatenFehler
			if f.fehler == nil && (!errors.As(err, &df) || errors.Is(err, model.ErrEingabeUngueltig)) {
				t.Errorf("%s %s: Fehler %v, erwartet der DatenFehler von src unverändert", f.name, art, err)
			}
			if got.Properties != nil {
				t.Errorf("%s %s: Ergebnis %+v", f.name, art, got)
			}
			s.verboten(t, f.name)
			n++
		}
	}
	if n != 18 {
		t.Errorf("Menge %d, erwartet 18", n)
	}
	for _, art := range eArten() {
		got, err := func() (model.EntityEingabe, error) {
			a := eA("Mitten", art, "")
			a.src = eNeuerStrom(eSchritt{daten: []byte(`{x`), err: quelle})
			return a.los()
		}()
		eFehler(t, "Verstoß und Fehler im selben Read", got, err, model.EingabeSyntax, "")
	}
}

// TestEingabeNurRead hält Zeile 11: `Close`, `WriteTo`, `ReadByte`, `Seek` und `ReadAt` werden nie gerufen, kein Read folgt auf das
// Ende; Gutfall, Verstoß des Lesens und Verstoß des Deutens.
func TestEingabeNurRead(t *testing.T) {
	for _, body := range []string{`{"Id":1}`, `{"Id":`, `{"x":1}`, ""} {
		for _, art := range eArten() {
			s := eStromText(body)
			a := eA("Mitten", art, "")
			a.src = s
			_, _ = a.los()
			s.verboten(t, "Body "+body)
			if s.readN == 0 {
				t.Errorf("Body %q: kein Read", body)
			}
		}
	}
}

// eRumpf ist ein Body der Länge n (n >= 2): `{`, Leerraum, `}`.
func eRumpf(n int) string { return "{" + strings.Repeat(" ", n-2) + "}" }

// TestEingabeGroesse hält Zeile 12: höchstens `grenze` Bytes, das Byte `grenze + 1` ist `zu_gross`, nichts dahinter wird gelesen. Menge
// 66 = (7 Grenzen × 3 Bodys + MaxInt64 mit 3 Bytes) × 3 Arten.
func TestEingabeGroesse(t *testing.T) {
	n := 0
	for _, art := range eArten() {
		for _, g := range []int64{1, 2, 3, 100, 32767, 32768, 32769, math.MaxInt64} {
			laengen := []int64{g - 1, g, g + 1}
			if g == math.MaxInt64 {
				laengen = []int64{3}
			}
			for _, l := range laengen {
				var body string
				switch {
				case l <= 0:
					body = ""
				case l == 1:
					body = "5"
				default:
					body = eRumpf(int(l))
				}
				ueber := l > g
				if ueber {
					body = strings.Repeat(" ", int(l)) + "}}}garbage" // der Müll hinter dem Byte grenze + 1 wird nie gesehen
				}
				s := eStromText(body)
				a := eA("Mitten", art, "")
				a.src, a.grenze = s, g
				got, err := a.los()
				was := "grenze " + eP(int(min(g, 99999999))) + " Länge " + eP(int(l)) + " " + string(art)
				switch {
				case ueber:
					eFehler(t, was, got, err, model.EingabeZuGross, "")
				case l <= 0:
					eFehler(t, was, got, err, model.EingabeSyntax, "")
				case l == 1:
					eFehler(t, was, got, err, model.EingabeStruktur, "")
				default:
					eMitteLeer(t, was, art, got, err)
				}
				// Kein Read fordert mehr an, als `grenze + 1` abzüglich der schon gelieferten Bytes zulassen (SPEC-126 Nr. 5); die Summe aller
				// Anforderungen ist bei einem Strom, der weniger liefert als gefordert, nicht durch `grenze + 1` begrenzt (Rückgabe an den Architect).
				var geliefert int64
				for i, x := range s.geliefert {
					if a := int64(s.angefordert[i]); a < 1 || a > 32768 || (g != math.MaxInt64 && a > g+1-geliefert) {
						t.Errorf("%s: Read %d fordert %d Bytes an, nach %d gelieferten Bytes", was, i, a, geliefert)
					}
					geliefert += int64(x)
				}
				if g != math.MaxInt64 && (geliefert > g+1 || (ueber && geliefert != g+1)) {
					t.Errorf("%s: geliefert %d, höchstens %d erlaubt und bei zu_gross genau so viele", was, geliefert, g+1)
				}
				if ueber && s.nachEnde != 0 {
					t.Errorf("%s: Read nach dem Verstoß", was)
				}
				n++
			}
		}
	}
	if n != 66 {
		t.Errorf("Menge %d, erwartet 66", n)
	}
}

// eTiefeBody ist ein Body, dessen Annotation `@a` Arrays oder Objekte so verschachtelt, dass das Dokument die Tiefe d hat.
func eTiefeBody(d int, objekte bool) string {
	if d == 1 {
		return `{"Id":1,"@a":0}`
	}
	if objekte {
		return `{"Id":1,"@a":` + strings.Repeat(`{"a":`, d-1) + "0" + strings.Repeat("}", d-1) + "}"
	}
	return `{"Id":1,"@a":` + strings.Repeat("[", d-1) + strings.Repeat("]", d-1) + "}"
}

// TestEingabeTiefe hält Zeile 13: 512 Ebenen sind erlaubt, die 513. offene Klammer ist `tiefe`. Menge 36 = 6 × 2 × 3.
func TestEingabeTiefe(t *testing.T) {
	n := 0
	for _, d := range []int{1, 511, 512, 513, 514, 1000} {
		for _, objekte := range []bool{false, true} {
			for _, art := range eArten() {
				got, err := eA("Mitten", art, eTiefeBody(d, objekte)).los()
				was := "Tiefe " + eP(d) + " " + string(art)
				if d <= 512 {
					eOK(t, was, got, err, model.EntityEingabe{Typ: "Ns.Mitte", Schluessel: "1", Properties: []model.PropertyWert{ePW("Id", model.ArtInt32, "1")}})
				} else {
					eFehler(t, was, got, err, model.EingabeTiefe, "")
				}
				n++
			}
		}
	}
	if n != 36 {
		t.Errorf("Menge %d, erwartet 36", n)
	}
}

// eDoppelt liefert den Inhalt eines Objekts mit doppeltem Namen N in der Form f.
func eDoppelt(f int) string {
	return [...]string{`"N":1,"N":1`, `"N":1,"N":2`, `"N":1,"` + esc("004E") + `":1`, `"N":1,"M":2,"N":3`}[f]
}

// TestEingabeDoppelterName hält Zeile 14: `doppelter_name` in jedem Objekt des Dokuments. Menge 26 = 5 Stellen × 4 Formen + 2 Gegenfälle × 3.
func TestEingabeDoppelterName(t *testing.T) {
	stellen := []func(inner string) string{
		func(i string) string { return "{" + i + "}" },
		func(i string) string { return `{"Id":1,"@a":{` + i + `}}` },
		func(i string) string { return `{"Id":1,"Dbl":{` + i + `}}` },
		func(i string) string { return `{"Id":1,"@a":{"x":{` + i + `}}}` },
		func(i string) string { return `{"Id":1,"Einzel":{` + i + `}}` },
	}
	n := 0
	for s, stelle := range stellen {
		for f := range 4 {
			got, err := eA("Ms", model.EingabePatch, stelle(eDoppelt(f))).los()
			eFehler(t, "Stelle "+eP(s)+" Form "+eP(f), got, err, model.EingabeDoppelterName, "")
			n++
		}
	}
	gegen := []string{`{"Id":1,"@a":{"x":1,"X":2}}`, `{"Id":1,"@a":{"x":1},"@b":{"x":1}}`}
	for _, b := range gegen {
		for _, art := range eArten() {
			got, err := eA("Ms", art, b).los()
			eOK(t, b, got, err, model.EntityEingabe{Typ: "Ns.M", Schluessel: "1", Properties: []model.PropertyWert{ePW("Id", model.ArtInt32, "1")}})
			n++
		}
	}
	if n != 26 {
		t.Errorf("Menge %d, erwartet 26", n)
	}
}

// eLesenAusloeser sind die Auslöser der sechs Ursachen des Lesens: ein Body und die Grenze; hinter `vor` steht der Rest, der
// einen Verstoß des Lesens trägt, wenn ein Verstoß des Deutens davor steht (Zeile 16).
type eLesenAusloeser struct {
	ursache model.EingabeUrsache
	ort     string
	body    func(vor string) (string, int64)
}

func eLesenAusloeserListe() []eLesenAusloeser {
	return []eLesenAusloeser{
		{model.EingabeSyntax, "", func(v string) (string, int64) { return `{` + v + `,"n1":x}`, eGrenze }},
		{model.EingabeKodierung, "", func(v string) (string, int64) { return `{` + v + `,"n1":"` + "\xff" + `"}`, eGrenze }},
		{model.EingabeTiefe, "", func(v string) (string, int64) {
			return `{` + v + `,"@z":` + strings.Repeat("[", 513) + strings.Repeat("]", 513) + `}`, eGrenze
		}},
		{model.EingabeZuGross, "", func(v string) (string, int64) {
			k := `{` + v + `,"n1":1}`
			return k + strings.Repeat(" ", 10), int64(len(k)) + 5
		}},
		{model.EingabeDoppelterName, "", func(v string) (string, int64) { return `{` + v + `,"n1":1,"n1":2}`, eGrenze }},
	}
}

// TestEingabeLesenUrsachen hält Zeile 15: jede Ursache des Lesens kommt als gleichnamige Ursache der Klasse `ungueltig` mit leerem
// Ort an, `zahl_zu_lang` mit dem Ort des Wertes. Menge 18 = 6 Ursachen × 3 Arten.
func TestEingabeLesenUrsachen(t *testing.T) {
	n := 0
	for _, art := range eArten() {
		for _, l := range eLesenAusloeserListe() {
			body, g := l.body(`"Id":1`)
			a := eA("Ws", art, body)
			a.grenze = g
			got, err := a.los()
			eFehler(t, string(l.ursache), got, err, l.ursache, "")
			n++
		}
		got, err := eA("Ws", art, `{"Id":1,"Dec":1e1024}`).los()
		eFehler(t, "zahl_zu_lang", got, err, model.EingabeZahlZuLang, "/Dec")
		n++
	}
	if n != 18 {
		t.Errorf("Menge %d, erwartet 18", n)
	}
}
