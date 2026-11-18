package urlparser_test

import (
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die Menge der Segmente eines Pfads von `$expand` (SPEC-262 Nr. 4 c, d), getrennt nach Zeichenmenge, Struktur, Wörtern und Länge.
// Das Orakel ist `exOrakel` (Tabelle der Zeichenklassen, kein Code des Lesers); jedes Segment steht an vier Stellen: als erstes,
// mittleres und letztes Segment eines Pfads aus drei Segmenten und als ganzes Element. Zeichen jenseits von ASCII entstehen aus
// Bytewerten (`AGENTS.md` §3.18).

// exSegmentStellen setzt ein Segment an die vier Stellen.
func exSegmentStellen(seg string) [4]string {
	return [4]string{seg + "/N/N", "N/" + seg + "/N", "N/N/" + seg, seg}
}

// exSegmentForm baut ein Segment aus k Namen, durch `.` getrennt; der Teil idx trägt den Inhalt, jeder andere ist `a`.
func exSegmentForm(k, idx int, inhalt string) string {
	teile := make([]string, k)
	for i := range teile {
		teile[i] = "a"
	}
	teile[idx] = inhalt
	return strings.Join(teile, ".")
}

// exZeichenMenge sind alle Zeichen, die ein Segment tragen kann, mit ihrer rohen und ihrer dekodierten Schreibweise: die 125 ASCII-
// Zeichen außer `,`, `/` und `(` roh (ohne Steuerzeichen, `&` und `%`: 90) und als `%XX` in beiden Schreibweisen (ohne `%00`: 248),
// dazu 7 Zeichen jenseits von ASCII in drei Schreibweisen (21).
func exZeichenMenge() []exSymbol {
	var s []exSymbol
	for b := range 128 {
		c := string([]byte{byte(b)})
		if c == "," || c == "/" || c == "(" {
			continue
		}
		if b >= 0x20 && b != 0x7f && c != "&" && c != "%" {
			s = append(s, exSymbol{c, c})
		}
		if b != 0 {
			s = append(s, exSymbol{exHex(c, true), c}, exSymbol{exHex(c, false), c})
		}
	}
	for _, z := range exNamenzeichen() {
		s = append(s, exSymbol{z.z, z.z}, exSymbol{exHex(z.z, true), z.z}, exSymbol{exHex(z.z, false), z.z})
	}
	return s
}

// exKlassenZaehler hält je Klasse der Antwort die Zahl der geprüften Fälle, damit eine Achse nicht allein aus einer Klasse besteht.
type exKlassenZaehler map[model.URLUrsache]int

func (k exKlassenZaehler) mindestens(t *testing.T, name string, n int) {
	t.Helper()
	for _, u := range []model.URLUrsache{"", model.URLWert, model.URLWertNichtUnterstuetzt} {
		if k[u] < n {
			t.Errorf("%s, Klasse %q: nur %d Fälle", name, u, k[u])
		}
	}
}

// TestExpandSegmentZeichen (Zeile 12, SPEC-262 Nr. 4 c, d): jedes Zeichen an jeder Stelle jedes Teils eines Segments hat die Klasse
// des Orakels. Achsen: 359 Zeichen (Schreibweisen eingerechnet) × 24 Formen (1 bis 3 Teile, der Teil mit dem Zeichen an jeder
// Stelle des Segments, Inhalt `c`, `ac`, `ca`, `aca`) × 4 Stellen im Pfad = 34 464 Fälle.
func TestExpandSegmentZeichen(t *testing.T) {
	p := exParser()
	zeichen := exZeichenMenge()
	exZaehlt(t, "Segmentzeichen", len(zeichen), 90+248+21)
	inhalte := []func(string) string{
		func(c string) string { return c },
		func(c string) string { return "a" + c },
		func(c string) string { return c + "a" },
		func(c string) string { return "a" + c + "a" },
	}
	n, formen := 0, 0
	klassen := exKlassenZaehler{}
	for k := 1; k <= 3; k++ {
		for idx := range k {
			for _, inhalt := range inhalte {
				formen++
				for _, z := range zeichen {
					roh := exSegmentStellen(exSegmentForm(k, idx, inhalt(z.roh)))
					dek := exSegmentStellen(exSegmentForm(k, idx, inhalt(z.dek)))
					for stelle := range roh {
						soll := exOrakel(dek[stelle])
						klassen[soll.u]++
						exPruefe(t, p, exQ(roh[stelle]), soll)
						n++
					}
				}
			}
		}
	}
	exZaehlt(t, "Segmentformen", formen, 24)
	exZaehlt(t, "Segmentzeichen × Form × Stelle", n, 359*24*4)
	klassen.mindestens(t, "Segmentzeichen", 100)
}

// TestExpandSegmentStruktur (Zeile 12, SPEC-262 Nr. 4 c, d): jede Folge der Länge 0 bis 4 über 11 Zeichen der Grammatik eines
// Segments (`a`, `Z`, `_`, `0`, `.`, `$`, `-`, `*`, ein Buchstabe, NBSP, eine Ziffer jenseits von ASCII) hat als Segment an jeder der
// vier Stellen die Klasse des Orakels. Menge: (1 + 11 + 11² + 11³ + 11⁴) Folgen × 4 Stellen = 64 420. Prüfgrenze: Länge 4; ein
// Verstoß, der erst ab Länge 5 greift, bleibt hier ungefangen (die Formen von `TestExpandSegmentZeichen` und die Wörter von
// `TestExpandSegmentWorte` tragen die längeren).
func TestExpandSegmentStruktur(t *testing.T) {
	p := exParser()
	alphabet := []string{"a", "Z", "_", "0", ".", "$", "-", "*", exE, exNBSP, exArab3}
	folgen := []string{""}
	vorige := []string{""}
	for range 4 {
		var naechste []string
		for _, v := range vorige {
			for _, a := range alphabet {
				naechste = append(naechste, v+a)
			}
		}
		folgen = append(folgen, naechste...)
		vorige = naechste
	}
	exZaehlt(t, "Segmentfolgen", len(folgen), 1+11+11*11+11*11*11+11*11*11*11)
	n := 0
	klassen := exKlassenZaehler{}
	for _, f := range folgen {
		for _, el := range exSegmentStellen(f) {
			soll := exOrakel(el)
			klassen[soll.u]++
			exPruefe(t, p, exQ(el), soll)
			n++
		}
	}
	exZaehlt(t, "Segmentfolgen × Stelle", n, 4*len(folgen))
	klassen.mindestens(t, "Segmentfolgen", 100)
}

// exWorte sind die Wörter mit `$`, die als Segment auftreten: die drei zulässigen (`$ref`, `$count`, `$value`) und 23 Systemnamen,
// die es nicht sind.
func exWorte() []string {
	return []string{"$ref", "$count", "$value", "$filter", "$select", "$orderby", "$top", "$skip", "$levels", "$expand", "$search",
		"$format", "$id", "$metadata", "$batch", "$entity", "$root", "$it", "$this", "$each", "$all", "$inlinecount", "$apply",
		"$compute", "$skiptoken", "$deltatoken"}
}

// exWortNachbarn sind das Wort und seine Nachbarn im Abstand eins: ein Zeichen gestrichen, ersetzt oder eingefügt (aus 7 Zeichen),
// jedes Zeichen nach dem `$` groß, das ganze Wort groß, dazu das Wort als Teil eines qualifizierten Namens.
func exWortNachbarn(w string) []string {
	ersatz := []string{"x", ".", "$", "0", "-", exE, exNBSP}
	v := []string{w, strings.ToUpper(w), w + ".a", "a." + w, "a." + w + ".a", w + "." + w}
	for i := 1; i < len(w); i++ {
		v = append(v, w[:i]+strings.ToUpper(w[i:i+1])+w[i+1:])
	}
	for i := range len(w) {
		v = append(v, w[:i]+w[i+1:])
	}
	for _, e := range ersatz {
		for i := range len(w) {
			v = append(v, w[:i]+e+w[i+1:])
		}
		for i := 0; i <= len(w); i++ {
			v = append(v, w[:i]+e+w[i:])
		}
	}
	return v
}

// TestExpandSegmentWorte (Zeile 12, SPEC-262 Nr. 4 c): Segmente mit `$` sind es genau für `$ref`, `$count` und `$value` in dieser
// Schreibweise; jedes andere Wort (die 26 von `exWorte` samt ihren Nachbarn im Abstand eins, auch mit `%24` für das `$`) macht das
// Element zu `wert`. Menge: 5720 Schreibweisen (die Wörter und Nachbarn, roh und mit `%24`, ohne Doppelte) × 4 Stellen = 22 880.
func TestExpandSegmentWorte(t *testing.T) {
	p := exParser()
	gesehen := map[string]bool{}
	n, zulaessig := 0, 0
	klassen := exKlassenZaehler{}
	for _, w := range exWorte() {
		for _, dek := range exWortNachbarn(w) {
			for _, roh := range []string{dek, strings.ReplaceAll(dek, "$", "%24")} {
				if gesehen[roh] {
					continue
				}
				gesehen[roh] = true
				if dek == "$ref" || dek == "$count" || dek == "$value" {
					zulaessig++
				}
				rohStellen, dekStellen := exSegmentStellen(roh), exSegmentStellen(dek)
				for stelle := range rohStellen {
					soll := exOrakel(dekStellen[stelle])
					klassen[soll.u]++
					exPruefe(t, p, exQ(rohStellen[stelle]), soll)
					n++
				}
			}
		}
	}
	exZaehlt(t, "Wörter ohne Doppelte", len(gesehen), 5720)
	exZaehlt(t, "Wörter × Stelle", n, 4*len(gesehen))
	exZaehlt(t, "zulässige Wörter", zulaessig, 6)
	klassen.mindestens(t, "Wörter", 10)
}

// TestExpandSegmentLaenge (Zeile 12, SPEC-262 Nr. 4 a, c): jeder Name eines Segments hat 1 bis 128 Zeichen (nicht Bytes), gleich in
// welchem Teil eines qualifizierten Namens er steht, und die Länge des Segments als Ganzes ist nicht begrenzt. Menge: 6 Paare aus
// Zahl der Teile (1 bis 3) und Teil mit dem langen Namen × Länge {1, 127, 128, 129} × Breite {1 Byte, 2 Bytes} × 4 Stellen = 192; dazu
// alle Teile 128 Zeichen lang (2 und 3 Teile) × 2 Breiten × 4 Stellen = 16.
func TestExpandSegmentLaenge(t *testing.T) {
	p := exParser()
	n := 0
	name := func(laenge int, breit bool) string {
		fuell, erstes := "a", "a"
		if breit {
			fuell, erstes = exE, exE
		}
		return erstes + strings.Repeat(fuell, laenge-1)
	}
	pruefe := func(seg string) {
		t.Helper()
		for _, el := range exSegmentStellen(seg) {
			exPruefe(t, p, exQ(el), exOrakel(el))
			n++
		}
	}
	for k := 1; k <= 3; k++ {
		for idx := range k {
			for _, laenge := range []int{1, 127, 128, 129} {
				for _, breit := range []bool{false, true} {
					pruefe(exSegmentForm(k, idx, name(laenge, breit)))
				}
			}
		}
	}
	for k := 2; k <= 3; k++ {
		for _, breit := range []bool{false, true} {
			teile := make([]string, k)
			for i := range teile {
				teile[i] = name(128, breit)
			}
			pruefe(strings.Join(teile, "."))
		}
	}
	exZaehlt(t, "Segmentlänge", n, 192+16)
}
