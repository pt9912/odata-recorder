package model

import (
	"fmt"
	"strings"
)

// maxZahlLaenge ist die Länge in Zeichen, die der kanonische Text einer Zahl nicht übersteigt (SPEC-166 Nr. 7).
const maxZahlLaenge = 1024

// wertKlasse ist die Klasse, in der Werte und Typen vergleichbar sind (SPEC-166 Nr. 2). Der Nullwert ist keine Klasse.
type wertKlasse int

const (
	klasseNull wertKlasse = iota + 1
	klasseText
	klasseWahrheit
	klasseZahl
	klasseGUID
	klasseDatum
	klasseZeitpunkt
)

// formFehler ist ErrUngueltig mit einem festen Grund; der Fehler trägt nie einen Wert (SPEC-166 Nr. 7).
func formFehler(grund string) error { return fmt.Errorf("%w: %s", ErrUngueltig, grund) }

// klasseVonArt liefert die Klasse einer Art; false für ArtRedigiert und jede unbekannte Art.
func klasseVonArt(a WertArt) (wertKlasse, bool) {
	switch a {
	case ArtNull:
		return klasseNull, true
	case ArtString:
		return klasseText, true
	case ArtBoolean:
		return klasseWahrheit, true
	case ArtInt32, ArtInt64, ArtDecimal:
		return klasseZahl, true
	case ArtGUID:
		return klasseGUID, true
	case ArtDate:
		return klasseDatum, true
	case ArtDateTimeOffset:
		return klasseZeitpunkt, true
	}
	return 0, false
}

// klasseVonEdmTyp liefert die Klasse eines der acht EDM-Typen; false für jeden anderen Namen.
func klasseVonEdmTyp(t string) (wertKlasse, bool) {
	switch t {
	case "Edm.String":
		return klasseText, true
	case "Edm.Boolean":
		return klasseWahrheit, true
	case "Edm.Int32", "Edm.Int64", "Edm.Decimal":
		return klasseZahl, true
	case "Edm.Guid":
		return klasseGUID, true
	case "Edm.Date":
		return klasseDatum, true
	case "Edm.DateTimeOffset":
		return klasseZeitpunkt, true
	}
	return 0, false
}

// HatOrdnung sagt, ob Werte des EDM-Typs eine Ordnung haben: wahr für sieben der acht Typen, falsch für Edm.Guid
// und für jeden Namen außerhalb der acht (SPEC-166 Nr. 3 f).
func HatOrdnung(edmType string) bool {
	k, ok := klasseVonEdmTyp(edmType)
	return ok && k != klasseGUID
}

// WertAusLiteral bildet ein Literal des Filters auf einen Wert ab (SPEC-166 Nr. 1).
func WertAusLiteral(l Literal) (Wert, error) {
	switch l.Art {
	case LiteralZeichenkette:
		t, ok := zeichenketteOhneAnfuehrung(l.Text)
		if !ok {
			return Wert{}, formFehler("Zeichenkette")
		}
		return Wert{Art: ArtString, Text: t}, nil
	case LiteralWahrheitswert:
		return Wert{Art: ArtBoolean, Text: l.Text}, nil
	case LiteralZahl:
		return Wert{Art: ArtDecimal, Text: l.Text}, nil
	case LiteralGUID:
		return Wert{Art: ArtGUID, Text: l.Text}, nil
	case LiteralDatum:
		return Wert{Art: ArtDate, Text: l.Text}, nil
	case LiteralZeitpunkt:
		return Wert{Art: ArtDateTimeOffset, Text: l.Text}, nil
	}
	return Wert{}, formFehler("Art des Literals")
}

// zeichenketteOhneAnfuehrung löst die äußeren `'` und jedes verdoppelte `'` auf; ein einzelnes `'` im Inhalt
// oder ein fehlendes äußeres `'` ergibt false.
func zeichenketteOhneAnfuehrung(s string) (string, bool) {
	if len(s) < 2 || s[0] != '\'' || s[len(s)-1] != '\'' {
		return "", false
	}
	inhalt := s[1 : len(s)-1]
	var b strings.Builder
	for i := 0; i < len(inhalt); i++ {
		if inhalt[i] == '\'' {
			if i+1 >= len(inhalt) || inhalt[i+1] != '\'' {
				return "", false
			}
			i++
		}
		b.WriteByte(inhalt[i])
	}
	return b.String(), true
}

// pruefeWert prüft Art und Form eines Wertes und liefert seine Klasse (SPEC-166 Nr. 7). ArtNull liest den Text nicht,
// ArtString nimmt jeden.
func pruefeWert(w Wert) (wertKlasse, error) {
	k, ok := klasseVonArt(w.Art)
	if !ok {
		return 0, formFehler("Art des Wertes")
	}
	var formOK bool
	switch w.Art {
	case ArtBoolean:
		formOK = w.Text == "true" || w.Text == "false"
	case ArtInt32, ArtInt64, ArtDecimal:
		_, _, _, formOK = zerlegeZahl(w.Text)
	case ArtGUID:
		formOK = guidForm(w.Text)
	case ArtDate:
		formOK = datumForm(w.Text)
	case ArtDateTimeOffset:
		formOK = zeitpunktForm(w.Text)
	default:
		formOK = true
	}
	if !formOK {
		return 0, formFehler("Form des Wertes")
	}
	return k, nil
}

func ziffer(c byte) bool { return c >= '0' && c <= '9' }

func nurZiffern(s string) bool {
	for i := 0; i < len(s); i++ {
		if !ziffer(s[i]) {
			return false
		}
	}
	return true
}

// zerlegeZahl zerlegt den kanonischen Text einer Zahl in Vorzeichen, ganzen Teil und Nachkommastellen
// (SPEC-166 Nr. 7): `-?(0|[1-9][0-9]*)`, höchstens ein `.` mit Bruchteil ohne Null am Ende, `-0` nicht, höchstens
// 1024 Zeichen.
func zerlegeZahl(s string) (negativ bool, ganz, bruch string, ok bool) {
	if len(s) == 0 || len(s) > maxZahlLaenge {
		return false, "", "", false
	}
	t, negativ := strings.CutPrefix(s, "-")
	ganz, bruch, hatBruch := strings.Cut(t, ".")
	if !ganzteilForm(ganz) || (hatBruch && !bruchteilForm(bruch)) {
		return false, "", "", false
	}
	if negativ && ganz == "0" && !hatBruch {
		return false, "", "", false
	}
	return negativ, ganz, bruch, true
}

// ganzteilForm prüft `0` oder eine Ziffernfolge ohne führende Null.
func ganzteilForm(g string) bool {
	return g != "" && nurZiffern(g) && (g[0] != '0' || len(g) == 1)
}

// bruchteilForm prüft eine nicht leere Ziffernfolge ohne Null am Ende.
func bruchteilForm(b string) bool {
	return b != "" && nurZiffern(b) && b[len(b)-1] != '0'
}

func istHexKlein(c byte) bool { return ziffer(c) || (c >= 'a' && c <= 'f') }

// guidForm prüft die 36 Zeichen `8-4-4-4-12` aus Kleinbuchstaben und Ziffern.
func guidForm(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if s[i] != '-' {
				return false
			}
		} else if !istHexKlein(s[i]) {
			return false
		}
	}
	return true
}

// datumForm prüft `JJJJ-MM-TT` aus Ziffern.
func datumForm(s string) bool {
	return len(s) == 10 && s[4] == '-' && s[7] == '-' && nurZiffern(s[:4]) && nurZiffern(s[5:7]) && nurZiffern(s[8:])
}

// zeitpunktForm prüft `JJJJ-MM-TTThh:mm:ssZ`, mit Bruchteil von 1 bis 12 Ziffern ohne Null am Ende.
func zeitpunktForm(s string) bool {
	if len(s) < 20 || s[len(s)-1] != 'Z' || s[10] != 'T' || !datumForm(s[:10]) || !uhrzeitForm(s[11:19]) {
		return false
	}
	rest := s[19 : len(s)-1]
	if rest == "" {
		return true
	}
	bruch, ok := strings.CutPrefix(rest, ".")
	return ok && len(bruch) <= 12 && bruchteilForm(bruch)
}

// uhrzeitForm prüft `hh:mm:ss` aus Ziffern.
func uhrzeitForm(s string) bool {
	return len(s) == 8 && s[2] == ':' && s[5] == ':' && nurZiffern(s[:2]) && nurZiffern(s[3:5]) && nurZiffern(s[6:])
}

// vergleicheZahl ordnet zwei kanonische Zahlen exakt: Vorzeichen, Stellen des ganzen Teils, Ziffern, Nachkommastellen.
func vergleicheZahl(a, b string) int {
	an, ag, ab, _ := zerlegeZahl(a)
	bn, bg, bb, _ := zerlegeZahl(b)
	switch {
	case an && !bn:
		return -1
	case !an && bn:
		return 1
	}
	c := len(ag) - len(bg)
	if c == 0 {
		c = strings.Compare(ag, bg)
	}
	if c == 0 {
		c = strings.Compare(ab, bb)
	}
	if an {
		return -sign(c)
	}
	return sign(c)
}

func sign(c int) int {
	switch {
	case c < 0:
		return -1
	case c > 0:
		return 1
	}
	return 0
}

// vergleicheZeitpunkt ordnet zwei kanonische Zeitpunkte: bis zur Sekunde nach dem Text, danach den Bruchteil als
// Dezimalbruch (ohne Null am Ende ordnet der Text der Ziffern, ein Präfix ist kleiner).
func vergleicheZeitpunkt(a, b string) int {
	if c := strings.Compare(a[:19], b[:19]); c != 0 {
		return sign(c)
	}
	return sign(strings.Compare(bruchteil(a), bruchteil(b)))
}

func bruchteil(z string) string {
	r := z[19 : len(z)-1]
	if r == "" {
		return ""
	}
	return r[1:]
}

// vergleicheGleichartig ordnet zwei geprüfte Werte derselben Klasse außer Guid.
func vergleicheGleichartig(k wertKlasse, a, b string) int {
	switch k {
	case klasseZahl:
		return vergleicheZahl(a, b)
	case klasseZeitpunkt:
		return vergleicheZeitpunkt(a, b)
	case klasseWahrheit:
		return sign(wahrheitsRang(a) - wahrheitsRang(b))
	}
	return sign(strings.Compare(a, b))
}

// pruefePaar prüft zwei Werte und ihre Vergleichbarkeit (SPEC-166 Nr. 2, 7) und liefert ihre Klassen.
func pruefePaar(links, rechts Wert) (kl, kr wertKlasse, err error) {
	kl, err = pruefeWert(links)
	if err != nil {
		return 0, 0, err
	}
	kr, err = pruefeWert(rechts)
	if err != nil {
		return 0, 0, err
	}
	if kl != klasseNull && kr != klasseNull && kl != kr {
		return 0, 0, formFehler("Klassen nicht vergleichbar")
	}
	return kl, kr, nil
}

// Vergleiche wertet `links op rechts` aus; das Ergebnis ist nie WahrheitUnbekannt (SPEC-166 Nr. 3, 4).
func Vergleiche(op VergleichOp, links, rechts Wert) (Wahrheit, error) {
	switch op {
	case VergleichEq, VergleichNe, VergleichGt, VergleichGe, VergleichLt, VergleichLe:
	default:
		return WahrheitUnbekannt, formFehler("Operator")
	}
	kl, kr, err := pruefePaar(links, rechts)
	if err != nil {
		return WahrheitUnbekannt, err
	}
	ordnung := op != VergleichEq && op != VergleichNe
	if ordnung && (kl == klasseGUID || kr == klasseGUID) {
		return WahrheitUnbekannt, formFehler("Ordnung auf Guid")
	}
	var c int
	switch {
	case kl == klasseNull && kr == klasseNull:
		c = 0
	case kl == klasseNull || kr == klasseNull:
		// genau ein null: nur ne ist wahr, in jeder Reihenfolge
		return wahr(op == VergleichNe), nil
	default:
		c = vergleicheGleichartig(kl, links.Text, rechts.Text)
	}
	return wahr(opTrifft(op, c)), nil
}

func wahr(b bool) Wahrheit {
	if b {
		return WahrheitWahr
	}
	return WahrheitFalsch
}

// opTrifft sagt, ob der Operator für das Vorzeichen eines Vergleichs gilt.
func opTrifft(op VergleichOp, c int) bool {
	switch op {
	case VergleichEq:
		return c == 0
	case VergleichNe:
		return c != 0
	case VergleichGt:
		return c > 0
	case VergleichGe:
		return c >= 0
	case VergleichLt:
		return c < 0
	}
	return c <= 0
}

// OrdneWerte ordnet zwei vergleichbare Werte und liefert -1, 0 oder 1; ArtNull ist kleiner als jeder Wert und gleich
// ArtNull (SPEC-166 Nr. 6). Guid hat keine Ordnung.
func OrdneWerte(links, rechts Wert) (int, error) {
	kl, kr, err := pruefePaar(links, rechts)
	if err != nil {
		return 0, err
	}
	if kl == klasseGUID || kr == klasseGUID {
		return 0, formFehler("Ordnung auf Guid")
	}
	switch {
	case kl == klasseNull && kr == klasseNull:
		return 0, nil
	case kl == klasseNull:
		return -1, nil
	case kr == klasseNull:
		return 1, nil
	}
	return vergleicheGleichartig(kl, links.Text, rechts.Text), nil
}

// wahrheitsRang ordnet false vor true.
func wahrheitsRang(s string) int {
	if s == "true" {
		return 1
	}
	return 0
}
