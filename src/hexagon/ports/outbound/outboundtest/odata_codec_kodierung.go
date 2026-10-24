package outboundtest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"iter"
	"testing"
	"unicode/utf8"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Mengen der Zeile B3 (SPEC-126 Nr. 3): das teilweise BOM (A2), die gültigen Zeichen roh an den Grenzen der Lead-Bytes (GZ)
// und die Einzelverstöße der UTF-8-Folgen (U), Orakel utf8: alle Ein- und Zwei-Byte-Folgen, jedes Lead-Byte mit jedem
// Steuerzeichen, Drei- und Vier-Byte-Folgen nur über 24 Grenzbytes (Prüfgrenze und Folge im Plan, Zeile B3).

// b3A2Zahl ist die Zahl der Körper von A2 je Methode: 4 Präfixe der BOM am Ende des Stroms, 4 × 255 Folgebytes, die das BOM
// nicht fortsetzen, am Ende des Stroms und 3 vollständige BOM hinter einem Leerzeichen (b3A2); dahinter jedes Byte und das
// Ende des Stroms, also 4 × 255 × 256 weitere Körper (b3A2Hinten): 262 147 zusammen.
const b3A2Zahl = 4 + 4*255 + 3 + 4*255*256

// b3A2Probe ist die Zahl der Körper von A2 je Methode unter dem Detektor: die ersten 1027 ohne das Byte dahinter.
const b3A2Probe = 4 + 4*255 + 3

// b3GZZahl ist die Zahl der Folgen von GZ: 2 + 6 + 4 + 12 + 4 + 8 + 8 + 24 + 8 über die neun Intervalle der Tabelle von UTF-8.
const b3GZZahl = 76

// b3BomPraefix ist ein echtes Präfix eines der drei BOM (`EF BB BF`, `FF FE`, `FE FF`) mit dem Byte, das es zum BOM fortsetzt.
type b3BomPraefix struct {
	praefix string
	fort    byte
}

func b3Bom() []b3BomPraefix {
	return []b3BomPraefix{{"\xEF", 0xBB}, {"\xEF\xBB", 0xBF}, {"\xFF", 0xFE}, {"\xFE", 0xFF}}
}

// b3A2Hinten sind die Körper von A2 mit einem Folgebyte, das das BOM nicht fortsetzt, und danach jedem Byte 0x00 bis 0xFF
// (4 × 255 × 256 = 261 120); das Ende des Stroms hinter dem Folgebyte steht in b3A2.
func b3A2Hinten() iter.Seq[[]byte] {
	return func(yield func([]byte) bool) {
		for _, p := range b3Bom() {
			for c := range 256 {
				for d := range 256 {
					if byte(c) != p.fort && !yield([]byte(p.praefix+string([]byte{byte(c), byte(d)}))) {
						return
					}
				}
			}
		}
	}
}

// b3A2Pruefer prüft einen Körper von b3A2Hinten: `syntax` ohne Ort, gleich welches Byte folgt, in beiden Methoden.
func (s *jsuite) b3A2Pruefer(b []byte, _ int64) (klasse, abweichung string) {
	for _, einzel := range []bool{true, false} {
		r := s.aufruf(context.Background(), einzel, bytes.NewReader(b), s.modelle["R"], "Products", 1<<30)
		var df *model.DatenFehler
		if !errors.As(r.err, &df) || df.Ursache != model.DatenSyntax || df.Ort != "" {
			return "syntax", fmt.Sprintf("%s, Körper %x: Ergebnis %v, erwartet syntax", methodeName(einzel), b, r.err)
		}
	}
	return "syntax", ""
}

// b3A2 ist die Menge A2 je Methode: ein unvollständiges oder abweichendes BOM ist kein BOM, sondern `syntax`.
func b3A2() []jfall {
	var out []jfall
	for _, p := range b3Bom() {
		out = append(out, roh(fmt.Sprintf("B3 A2 Präfix %x am Ende", p.praefix), []byte(p.praefix), model.DatenSyntax)...)
		for c := range 256 {
			if byte(c) != p.fort {
				out = append(out, roh(fmt.Sprintf("B3 A2 %x und %#02x", p.praefix, c), []byte(p.praefix+string([]byte{byte(c)})), model.DatenSyntax)...)
			}
		}
	}
	for _, bom := range []string{"\xEF\xBB\xBF", "\xFF\xFE", "\xFE\xFF"} {
		out = append(out, roh(fmt.Sprintf("B3 A2 BOM %x hinter einem Leerzeichen", bom), []byte(" "+bom+jEinzelR), model.DatenSyntax)...)
	}
	return out
}

// b3Intervall ist ein Intervall der Lead-Bytes der Tabelle von UTF-8 mit den Lead-Bytes der Menge GZ und den Bereichen
// (unten, oben) der Fortsetzungsbytes.
type b3Intervall struct {
	leads  []byte
	folgen [][2]byte
}

func b3Intervalle() []b3Intervall {
	fort := [2]byte{0x80, 0xBF}
	return []b3Intervall{
		{[]byte{0x20, 0x7F}, nil},
		{[]byte{0xC2, 0xD0, 0xDF}, [][2]byte{fort}},
		{[]byte{0xE0}, [][2]byte{{0xA0, 0xBF}, fort}},
		{[]byte{0xE1, 0xE8, 0xEC}, [][2]byte{fort, fort}},
		{[]byte{0xED}, [][2]byte{{0x80, 0x9F}, fort}},
		{[]byte{0xEE, 0xEF}, [][2]byte{fort, fort}},
		{[]byte{0xF0}, [][2]byte{{0x90, 0xBF}, fort, fort}},
		{[]byte{0xF1, 0xF2, 0xF3}, [][2]byte{fort, fort, fort}},
		{[]byte{0xF4}, [][2]byte{{0x80, 0x8F}, fort, fort}},
	}
}

// b3GZ ist die Menge GZ: je Intervall das erste, das letzte und bei mehr als zwei Lead-Bytes ein inneres Lead-Byte, dahinter
// die untere und die obere Grenze jedes Fortsetzungsbereichs in allen Kombinationen.
func b3GZ() []string {
	var out []string
	for _, iv := range b3Intervalle() {
		for _, lead := range iv.leads {
			folgen := []string{string([]byte{lead})}
			for _, bereich := range iv.folgen {
				var weiter []string
				for _, f := range folgen {
					for _, b := range bereich {
						weiter = append(weiter, f+string([]byte{b}))
					}
				}
				folgen = weiter
			}
			out = append(out, folgen...)
		}
	}
	return out
}

// b3GZFaelle sind die Fälle von GZ je Methode: jede Folge roh im Wert einer Property ist gültig, der Text gleich.
func b3GZFaelle() []jfall {
	var out []jfall
	for _, seq := range b3GZ() {
		out = append(out, beideGut(fmt.Sprintf("B3 GZ %x", seq), "R", "Products", b3Orte(seq)["Wert"],
			pw("ID", model.ArtInt32, "1"), pw("Name", model.ArtString, seq))...)
	}
	return out
}

// Kopf und Fuß des Bodys der Menge U: die Folge steht roh als Wert der Property Name eines Elements.
const (
	b3Kopf = `{"@odata.context":"$metadata#Products","value":[{"ID":1,"Name":"`
	b3Fuss = `"}]}`
)

// b3Orakel ist das Orakel der Menge U (utf8): die Ursache, die ein Inhalt roh in einem String hat. Die erste Stelle entscheidet:
// ein Byte, das keine gültige UTF-8-Folge fortsetzt oder beginnt, ist `kodierung` (auch vor einem Steuerzeichen oder dem
// schließenden `"`, die angefangene Folge geht vor), ein ganzes Zeichen unter 0x20, `"` oder `\` ist `syntax`.
func b3Orakel(seq []byte) model.DatenUrsache {
	for i := 0; i < len(seq); {
		r, n := utf8.DecodeRune(seq[i:])
		switch {
		case r == utf8.RuneError && n <= 1:
			return model.DatenKodierung
		case n == 1 && (seq[i] < 0x20 || seq[i] == '"' || seq[i] == '\\'):
			return model.DatenSyntax
		}
		i += n
	}
	return ""
}

// b3Pruefer prüft einen Body der Menge U: gültig heißt kein Fehler und der Text gleich der Folge, sonst die Ursache des Orakels.
func (s *jsuite) b3Pruefer(b []byte, _ int64) (klasse, abweichung string) {
	seq := b[len(b3Kopf) : len(b)-len(b3Fuss)]
	soll := b3Orakel(seq)
	r := s.aufruf(context.Background(), false, bytes.NewReader(b), s.modelle["R"], "Products", 1<<30)
	klasse = "gültig"
	if soll != "" {
		klasse = string(soll)
	}
	var df *model.DatenFehler
	switch {
	case soll == "" && r.err == nil && len(r.seite.Entities) == 1 && len(r.seite.Entities[0].Properties) == 2 &&
		r.seite.Entities[0].Properties[1].Wert.Text == string(seq):
		return klasse, ""
	case soll != "" && errors.As(r.err, &df) && df.Ursache == soll && df.Ort == "":
		return klasse, ""
	}
	return klasse, fmt.Sprintf("Folge %x: Ergebnis %v %+v, erwartet %s", seq, r.err, r.seite.Entities, klasse)
}

// b3Alphabet liefert die Bytes von bis bis (einschließlich) ohne die genannten, je als String von einem Byte.
func b3Alphabet(von, bis int, ohne ...byte) []string {
	var out []string
	for c := von; c <= bis; c++ {
		if !bytes.Contains(ohne, []byte{byte(c)}) {
			out = append(out, string([]byte{byte(c)}))
		}
	}
	return out
}

// b3LeadSteuer liefert die Menge U3: jedes Lead-Byte 0xC2 bis 0xF4 mit jedem Steuerzeichen 0x00 bis 0x1F dahinter.
func b3LeadSteuer() iter.Seq[[]byte] {
	return func(yield func([]byte) bool) {
		for lead := 0xC2; lead <= 0xF4; lead++ {
			for c := range 0x20 {
				if !yield([]byte(b3Kopf + string([]byte{byte(lead), byte(c)}) + b3Fuss)) {
					return
				}
			}
		}
	}
}

// jB3Mengen fährt die Mengen von B3: GZ und A2 stehen als Fälle in jfB3, hier die Zahlen und die Menge U. Unter dem Detektor
// läuft von U die Probe (Ein- und Zwei-Byte-Folgen, Drei-Byte-Folgen aus acht Grenzbytes); `make test` fährt die ganze Menge.
func jB3Mengen(t *testing.T, s *jsuite) {
	if n := len(b3A2()); n != 2*b3A2Probe {
		t.Errorf("B3 A2: %d Fälle, erwartet %d (%d Körper je Methode)", n, 2*b3A2Probe, b3A2Probe)
	}
	if !unterDetektor {
		klassen := gFahre(t, "B3 A2 hinten", b3A2Hinten(), s.b3A2Pruefer)
		var hinten int64
		for _, v := range klassen {
			hinten += v
		}
		gZaehle(t, "B3 A2", map[string]int64{"syntax": int64(len(b3A2())/2) + hinten}, b3A2Zahl)
	}
	if n := len(b3GZ()); n != b3GZZahl {
		t.Errorf("B3 GZ: %d Folgen, erwartet %d", n, b3GZZahl)
	}
	grenzBytes := []string{"\x20", "\x7F", "\x80", "\x8F", "\x90", "\x9F", "\xA0", "\xBF", "\xC0", "\xC1", "\xC2", "\xDF", "\xE0", "\xE1", "\xEC", "\xED",
		"\xEE", "\xEF", "\xF0", "\xF1", "\xF3", "\xF4", "\xF5", "\xFF"}
	drei, vier := grenzBytes, 4
	if unterDetektor {
		drei, vier = []string{"\x7F", "\x80", "\xBF", "\xC2", "\xE0", "\xED", "\xF0", "\xF4"}, 3
	}
	alle := b3Alphabet(0x00, 0xFF)
	zwei := b3Alphabet(0x20, 0xFF, '"', '\\')
	gZaehle(t, "B3 U Ein-Byte-Folgen", gFahre(t, "B3 U1", gFolgen(alle, 1, 1, b3Kopf, b3Fuss), s.b3Pruefer), 256)
	gZaehle(t, "B3 U Zwei-Byte-Folgen", gFahre(t, "B3 U2", gFolgen(zwei, 2, 2, b3Kopf, b3Fuss), s.b3Pruefer), 222*222)
	gZaehle(t, "B3 U Lead-Byte und Steuerzeichen", gFahre(t, "B3 U3", b3LeadSteuer(), s.b3Pruefer), 51*32)
	gZaehle(t, "B3 U Drei- und Vier-Byte-Folgen", gFahre(t, "B3 U4", gFolgen(drei, 3, vier, b3Kopf, b3Fuss), s.b3Pruefer),
		gFolge(len(drei), 3, vier))
}
