package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// Tests zu W-1 bis W-5 der zweiten Verifikation und zu den Mengen des Architect (SPEC-200 Nr. 2 d, 3, 4, 5, 8, 9, 13):
// jede Zusage läuft als Schleife über die Menge, die ihr Wortlaut meint; das Orakel steht im Test, von Hand gesetzt.
// Wo eine Funktion direkt aufrufbar ist, ruft der Test sie direkt, ohne Prozess. Zeichen jenseits von ASCII und jede
// Folge, die ein Schreibwerkzeug umsetzen könnte, entstehen aus Codepunkten oder Bytewerten (§3.18).

// geordnet liefert die Zeilen der Ausgabe von meldeBefunde ohne den Kopf der Ausgabe.
func geordnet(bf []befund) []string {
	var out, errb bytes.Buffer
	meldeBefunde(append([]befund{}, bf...), &out, &errb)
	if errb.Len() == 0 {
		return nil
	}
	var z []string
	for _, l := range strings.Split(strings.TrimSuffix(errb.String(), "\n"), "\n") {
		z = append(z, strings.TrimPrefix(l, ausgabeKopf))
	}
	return z
}

func gleichListe(a, b []string) bool {
	return len(a) == len(b) && strings.Join(a, "\x00") == strings.Join(b, "\x00")
}

// Z54 (Nr. 2 d): bei der Zeile 0 stehen `Kopf: <Schlüssel> fehlt` (in der Reihenfolge der Schlüssel) vor `version:
// Ausgabe nicht lesbar: …`, und dahinter die Befunde mit der Zeile 3. Menge: Software-Version {fehlt, leer, doppelt} ×
// Gültigkeitsbereich {fehlt, vorhanden} × Stand {fehlt, vorhanden} = 12, die Ausgabe von `version` ist Exit 1. Orakel von
// Hand: erst die `fehlt` der Schlüssel in Reihenfolge, dann `nicht lesbar`, dann `leer` bzw. `doppelt` (Zeile 3).
func TestHandbuchKopfGleichstandZeileNull(t *testing.T) {
	const nl = "version: Ausgabe nicht lesbar: Exit 1"
	m := neuesUmfeld(t, map[string]antwort{"version": {out: "odata-recorder dev\n", code: 1}})
	m.entferne("docs/user/beispiele")
	n := 0
	for _, sv := range []string{"fehlt", "leer", "doppelt"} {
		for _, gbFehlt := range []bool{true, false} {
			for _, stFehlt := range []bool{true, false} {
				zeilen := []string{"# H", ""}
				var soll []string
				switch sv {
				case "fehlt":
					soll = append(soll, "Kopf: Software-Version fehlt")
				case "leer":
					zeilen = append(zeilen, "Software-Version:")
				case "doppelt":
					zeilen = append(zeilen, "Software-Version: dev", "Software-Version: dev")
				}
				if gbFehlt {
					soll = append(soll, "Kopf: Gültigkeitsbereich fehlt")
				} else {
					zeilen = append(zeilen, "Gültigkeitsbereich: Binary")
				}
				if stFehlt {
					soll = append(soll, "Kopf: Stand fehlt")
				} else {
					zeilen = append(zeilen, "Stand: 2026-10-30")
				}
				soll = append(soll, nl)
				if sv != "fehlt" {
					soll = append(soll, "Kopf: Software-Version "+sv)
				}
				m.handbuch(strings.Join(zeilen, "\n") + "\n\n## A\n")
				code, _, errText := m.lauf()
				if code != exitBefund {
					t.Fatalf("%s %v %v: Exit %d %q", sv, gbFehlt, stFehlt, code, errText)
				}
				if ist := befundZeilen(t, errText); !gleichListe(ist, soll) {
					t.Fatalf("%s, Gültigkeitsbereich fehlt %v, Stand fehlt %v: %q, erwartet %q", sv, gbFehlt, stFehlt, ist, soll)
				}
				n++
			}
		}
	}
	if n != 12 {
		t.Fatalf("%d Elemente, erwartet 12", n)
	}
}

// Z55 (Nr. 2 d, 5 b, 5 d): die Info-Zeichenkette steht vor jedem anderen Befund desselben Zaunes. Menge: Info `text`
// zusammen mit {Beispiel ohne Bindungszeile, Bindung ungültig, Beispiel weicht ab, Datei fehlt} = 4; der Zaun steht in
// Zeile 10, genau zwei Zeilen, die Info zuerst, beide mit derselben Zeile.
func TestHandbuchInfoVorBindungsbefund(t *testing.T) {
	const info = "Zeile 10: Info-Zeichenkette text nicht zulässig"
	mit := neuesUmfeld(t, standardAntworten())
	ohne := ohneBeispiele(t)
	faelle := []struct {
		name  string
		m     umfeld
		prosa string
		soll  string
	}{
		{"ohne Bindungszeile", ohne, "Text.\n```text\n" + aufrufDatei + "```\n", "Zeile 10: Beispiel ohne Bindungszeile"},
		{"ungültig", ohne, "Datei: [beispiele/a.txt](beispiele/b.txt)\n```text\n" + aufrufDatei + "```\n",
			"Zeile 10: Bindung ungültig: Datei: [beispiele/a.txt](beispiele/b.txt)"},
		{"weicht ab", mit, binde("beispiele/aufruf.txt") + "```text\n" + aufrufDatei + "x\n```\n", "Zeile 10: Beispiel weicht von beispiele/aufruf.txt ab"},
		{"fehlt", ohne, binde("beispiele/fehlt.txt") + "```text\n" + aufrufDatei + "```\n", "Zeile 10: beispiele/fehlt.txt fehlt"},
	}
	for _, f := range faelle {
		f.m.handbuch(rumpf(f.prosa))
		code, _, errText := f.m.lauf()
		if code != exitBefund {
			t.Fatalf("%s: Exit %d %q", f.name, code, errText)
		}
		if ist := befundZeilen(t, errText); !gleichListe(ist, []string{info, f.soll}) {
			t.Fatalf("%s: %q, erwartet %q", f.name, ist, []string{info, f.soll})
		}
	}
}

// Z56 (Nr. 9 b, 13 a, 13 c): *n* eines Befundes von Nr. 9 b ist die Zeile, in der die öffnende Backtick-Folge steht;
// Zeichen von Listen, Zitaten und Tabellen ändern weder Absatz noch *n*; eine Folge ohne Gegenstück öffnet keinen Span;
// ein gebundener Zaun trägt keinen Span und trennt den Absatz. Menge: Lage des Spans × Absatzform, 23 Elemente. Die erste
// freie Zeile ist 9; jedes Element prüft die Liste der Befunde genau.
func TestHandbuchMakeZeile(t *testing.T) {
	m := ohneBeispiele(t)
	zielDateien(m)
	nix := func(z int) string { return fmt.Sprintf("Zeile %d: make-Ziel nix unbekannt", z) }
	nox := func(z int) string { return fmt.Sprintf("Zeile %d: make-Ziel nox unbekannt", z) }
	type element struct {
		name  string
		voll  string
		prosa string
		soll  []string
	}
	var es []element
	add := func(name, prosa string, soll ...string) {
		es = append(es, element{name: name, prosa: prosa, soll: soll})
	}
	add("1 eine Zeile", "Siehe `make nix`.\n", nix(9))
	add("2 zweite Zeile", "Eins.\nDann `make nix` hier.\n", nix(10))
	add("3 Span über zwei Zeilen", "Siehe `make\nnix` hier.\n", nix(9))
	add("4 zwei Spans", "Eins.\nDa `make\nnix` und\nweiter `make nox`.\n", nix(10), nox(12))
	add("5 ungepaarte Folge davor", "Eins `` hier.\nDann `make nix` hier.\n", nix(10))
	add("6 Absätze paaren nicht", "Eins `` hier.\n\nZwei `make nix` und ``.\n", nix(11))
	add("7 zwei Spans in einer Zeile", "`make nix` und `make nox`.\n", nix(9), nox(9))
	for n := 1; n <= 4; n++ {
		f := strings.Repeat("`", n)
		add(fmt.Sprintf("8a Folge der Länge %d", n), "Siehe "+f+"make nix"+f+" x\n", nix(9))
	}
	for n := 1; n <= 3; n++ {
		add(fmt.Sprintf("8b zu lange Folge zu %d", n), "Siehe "+strings.Repeat("`", n)+"make nix"+strings.Repeat("`", n+1)+" x\n")
	}
	for n := 2; n <= 3; n++ {
		add(fmt.Sprintf("8c zu kurze Folge zu %d", n), "Siehe "+strings.Repeat("`", n)+"make nix"+strings.Repeat("`", n-1)+" x\n")
	}
	add("9 Listenpunkt", "- Eins.\n- Dann `make nix`.\n", nix(10))
	add("10 Zitat", "> Eins.\n> `make nix`\n", nix(10))
	add("11 Tabelle", "| a | b |\n|---|---|\n| `make nix` | x |\n", nix(11))
	add("12 Überschrift", "## `make nix`\n", nix(9))
	es = append(es, element{name: "13 Kopf", voll: "# H\n\nSoftware-Version: dev\nGültigkeitsbereich: `make nix`\nStand: 2026-10-30\n\n## A\n", soll: []string{nix(4)}})
	n := 0
	for _, e := range es {
		h := e.voll
		if h == "" {
			h = rumpf(e.prosa)
		}
		m.handbuch(h)
		code, _, errText := m.lauf()
		if len(e.soll) == 0 {
			if code != exitOK {
				t.Fatalf("%s: Exit %d %q", e.name, code, errText)
			}
		} else if code != exitBefund {
			t.Fatalf("%s: Exit %d %q", e.name, code, errText)
		}
		if ist := befundZeilen(t, errText); !gleichListe(ist, e.soll) {
			t.Fatalf("%s: %q, erwartet %q", e.name, ist, e.soll)
		}
		n++
	}
	// 14: ein Span über einen gebundenen Zaun; 15: der Inhalt eines gebundenen Zaunes (Zitat) trägt keinen Span.
	mit := neuesUmfeld(t, standardAntworten())
	mit.handbuch(rumpf("Eins `make\n" + binde("beispiele/aufruf.txt") + zaunVon(aufrufDatei) + "nix` hier.\n"))
	if code, _, errText := mit.lauf(); code != exitOK {
		t.Fatalf("14 Span über den Zaun: Exit %d %q", code, errText)
	}
	n++
	zm, _ := bspZitat(t, "`make nix`")
	if code, _, errText := zm.lauf(); code != exitOK {
		t.Fatalf("15 Zitat: Exit %d %q", code, errText)
	}
	n++
	if n != 23 {
		t.Fatalf("%d Elemente, erwartet 23", n)
	}
}

// kopfMitWert liest den Kopf eines Handbuchs mit dem Wert wert als Software-Version.
func kopfMitWert(wert string) kopf {
	return leseKopf(zerlege([]string{"# H", "", "Software-Version: " + wert, "Gültigkeitsbereich: B", "Stand: 2026-10-30"}))
}

// Z57 (Nr. 4 c): der Wert im Kopf muss gleich *V* sein. Menge: Paare (Kopf *K*, Ausgabe *V*), 21 Elemente: Schreibung (3),
// Teil: Präfix, Suffix, Mitte in beiden Richtungen (6), ein Zeichen anders an jeder Stelle von `1.2.3` (5), Leerraum im
// Inneren (2), Nicht-ASCII (2), gleich (3; der Wert im Kopf wird getrimmt). Orakel: bei Ungleich genau ein Befund mit
// dem getrimmten Wert des Kopfes und V, sonst keiner. Die Funktion läuft direkt.
func TestHandbuchVersionVergleich(t *testing.T) {
	nb := string(rune(0xA0))
	ae := string(rune(0xE4))
	paare := []struct {
		k, v   string
		gleich bool
		kopf   string // der Wert, wie ihn der Befund nennt, wenn er vom Eingabewert abweicht
	}{
		{"Dev", "dev", false, ""}, {"dev", "Dev", false, ""}, {"DEV", "dev", false, ""},
		{"1.0", "1.0.1", false, ""}, {"1.0.1", "1.0", false, ""}, {"1.0", "x1.0", false, ""},
		{"x1.0", "1.0", false, ""}, {"1.0", "x1.0y", false, ""}, {"x1.0y", "1.0", false, ""},
		{"x.2.3", "1.2.3", false, ""}, {"1x2.3", "1.2.3", false, ""}, {"1.x.3", "1.2.3", false, ""},
		{"1.2x3", "1.2.3", false, ""}, {"1.2.x", "1.2.3", false, ""},
		{"1.0 x", "1.0", false, ""}, {"1." + nb + "0", "1.0", false, ""},
		{"1.0" + ae, "1.0a", false, ""}, {"1.0a", "1.0" + ae, false, ""},
		{"1.0", "1.0", true, ""}, {"\t1.0\t", "1.0", true, ""}, {nb + "1.0" + nb, "1.0", true, ""},
	}
	if len(paare) != 21 {
		t.Fatalf("%d Elemente, erwartet 21", len(paare))
	}
	for _, p := range paare {
		k := kopfMitWert(p.k)
		bf := versionBefunde(ergebnis{stdout: "odata-recorder " + p.v + "\n"}, k)
		if p.gleich {
			if len(bf) != 0 {
				t.Fatalf("Kopf %q, Ausgabe %q: %q, erwartet kein Befund", p.k, p.v, geordnet(bf))
			}
			continue
		}
		soll := "Software-Version " + strings.TrimSpace(p.k) + " weicht von der Ausgabe von version ab (" + p.v + ")"
		if ist := geordnet(bf); !gleichListe(ist, []string{soll}) {
			t.Fatalf("Kopf %q, Ausgabe %q: %q, erwartet %q", p.k, p.v, ist, soll)
		}
	}
}

// Z58 (Nr. 4 c): jede nicht leere Fehlerausgabe ist `Fehlerausgabe nicht leer`, und die Form der Standardausgabe ist
// `odata-recorder <V>` mit einem Zeilenumbruch. Menge: 25 Fehlerausgaben (die 19 Leerraum-Zeichen, Leerzeichen, `\n`,
// `\r\n`, `\n\n`, `x`, `x\n`) und die leere (Gutfall); dazu 7 Formen der Standardausgabe, die zu den 7 von
// `TestHandbuchVersionForm` kommen. Orakel von Hand; die Funktion läuft direkt, der Vergleich entfällt bei beiden.
func TestHandbuchVersionFehlerausgabeUndForm(t *testing.T) {
	k := kopfMitWert("dev")
	gut := ergebnis{stdout: "odata-recorder dev\n"}
	if bf := versionBefunde(gut, k); len(bf) != 0 {
		t.Fatalf("Gutfall: %q", geordnet(bf))
	}
	fehl := append(leerraum19(), " ", "\n", "\r\n", "\n\n", "x", "x\n")
	if len(fehl) != 25 {
		t.Fatalf("%d Fehlerausgaben, erwartet 25", len(fehl))
	}
	for _, f := range fehl {
		e := gut
		e.stderr = f
		ist := geordnet(versionBefunde(e, k))
		if !gleichListe(ist, []string{"version: Ausgabe nicht lesbar: Fehlerausgabe nicht leer"}) {
			t.Fatalf("Fehlerausgabe %q: %q", f, ist)
		}
	}
	formen := []string{
		"Odata-Recorder dev\n", "xodata-recorder dev\n", "odata-recorderx dev\n", "odata-recorder\n",
		"odata-recorder dev\r\n", "\nodata-recorder dev\n", "odata-recorder dev\nodata-recorder dev\n",
	}
	for _, s := range formen {
		ist := geordnet(versionBefunde(ergebnis{stdout: s}, k))
		if len(ist) != 1 || !strings.HasPrefix(ist[0], "version: Ausgabe nicht lesbar: Standardausgabe ") {
			t.Fatalf("Standardausgabe %q: %q", s, ist)
		}
	}
}

// Z59 (Nr. 2 d, 4 b): bei `doppelt` ist die Zeile die des ersten Vorkommens. Menge: Schlüssel (3) × Zahl der Vorkommen
// 2, 3, 4 × Lage der beiden anderen Schlüssel (vor dem Block: P = 2, hinter ihm: P = 0) = 18; dazu je Schlüssel für 2 und 3
// Vorkommen jede Belegung mit {gefüllt, leer} (4 und 8) = 36. Aufbau von Hand: Zeile 1 `# H`, Zeile 2 leer, dann P Zeilen,
// dann `f0, s, f1, s, …, s, fk`; die Zeile von fj ist 3 + P + 2j, die des ersten s ist 4 + P. Orakel: die Befunde in der
// Reihenfolge der Zeilen; der erste s steht hinter f0 und vor f1. In der Wertschleife erscheint nur `doppelt`.
func TestHandbuchKopfDoppeltErstesVorkommen(t *testing.T) {
	const fremd = "Zeile ohne Schlüssel"
	n := 0
	for _, s := range kopfSchluessel() {
		var andere []string
		for _, o := range kopfSchluessel() {
			if o != s {
				andere = append(andere, schluesselZeilen()[o])
			}
		}
		doppelt := "Kopf: " + s + " doppelt"
		for k := 2; k <= 4; k++ {
			for _, p := range []int{0, 2} {
				zeilen := []string{"# H", ""}
				if p == 2 {
					zeilen = append(zeilen, andere...)
				}
				var soll []string
				for j := 0; j <= k; j++ {
					zeilen = append(zeilen, fremd)
					soll = append(soll, fmt.Sprintf("Zeile %d: Kopf: %s", 3+p+2*j, fremd))
					if j == 0 {
						soll = append(soll, doppelt)
					}
					if j < k {
						zeilen = append(zeilen, schluesselZeilen()[s])
					}
				}
				if p == 0 {
					zeilen = append(zeilen, andere...)
				}
				ist := geordnet(leseKopf(zerlege(append(zeilen, "", "## A"))).befunde)
				if !gleichListe(ist, soll) {
					t.Fatalf("%s, %d Vorkommen, P = %d: %q, erwartet %q", s, k, p, ist, soll)
				}
				n++
			}
		}
		for k := 2; k <= 3; k++ {
			for maske := 0; maske < 1<<k; maske++ {
				zeilen := []string{"# H", ""}
				for v := 0; v < k; v++ {
					if maske&(1<<v) != 0 {
						zeilen = append(zeilen, s+":")
					} else {
						zeilen = append(zeilen, schluesselZeilen()[s])
					}
				}
				zeilen = append(append(zeilen, andere...), "", "## A")
				if ist := geordnet(leseKopf(zerlege(zeilen)).befunde); !gleichListe(ist, []string{doppelt}) {
					t.Fatalf("%s, %d Vorkommen, Belegung %b: %q, erwartet %q", s, k, maske, ist, doppelt)
				}
				n++
			}
		}
	}
	if n != 54 {
		t.Fatalf("%d Elemente, erwartet 54", n)
	}
}

// Z60 (Nr. 5 b: „jeder Zaun“): Menge Eigenschaft (5) × Stelle (4) = 20. Vier Zäune stehen hintereinander, nur der j-te
// ist fehlerhaft, die anderen sind sauber und binden `aufruf.txt`. Ein sauberer Block belegt 10 Zeilen (Bindungszeile,
// Zaun, 6 Zeilen, Zaun, Leerzeile); der j-te beginnt in Zeile 9 + 10(j − 1). Orakel: genau ein Befund, *n* von Hand.
func TestHandbuchZaunEigenschaftJeStelle(t *testing.T) {
	m := neuesUmfeld(t, standardAntworten())
	const bind = "Datei: [beispiele/aufruf.txt](beispiele/aufruf.txt)\n"
	sauberer := bind + "```\n" + aufrufDatei + "```\n\n"
	eigenschaften := []struct {
		name  string
		block string
		off   int // Zeilen vom Beginn des Blocks bis zum Zaun
		text  string
	}{
		{"Info", bind + "```text\n" + aufrufDatei + "```\n\n", 1, "Zeile %d: Info-Zeichenkette text nicht zulässig"},
		{"Bindungszeile fehlt", "```\n" + aufrufDatei + "```\n\n", 0, "Zeile %d: Beispiel ohne Bindungszeile"},
		{"Bindung ungültig", "Datei: [beispiele/aufruf.txt](beispiele/anderes.txt)\n```\n" + aufrufDatei + "```\n\n", 1,
			"Zeile %d: Bindung ungültig: Datei: [beispiele/aufruf.txt](beispiele/anderes.txt)"},
		{"Block ungleich", bind + "```\n" + aufrufDatei + "x\n```\n\n", 1, "Zeile %d: Beispiel weicht von beispiele/aufruf.txt ab"},
		{"Datei fehlt", "Datei: [beispiele/fehlt.txt](beispiele/fehlt.txt)\n```\n" + aufrufDatei + "```\n\n", 1, "Zeile %d: beispiele/fehlt.txt fehlt"},
	}
	n := 0
	for _, e := range eigenschaften {
		for j := 1; j <= 4; j++ {
			var b strings.Builder
			for i := 1; i <= 4; i++ {
				if i == j {
					b.WriteString(e.block)
				} else {
					b.WriteString(sauberer)
				}
			}
			m.handbuch(rumpf(b.String()))
			code, _, errText := m.lauf()
			soll := fmt.Sprintf(e.text, 9+10*(j-1)+e.off)
			if code != exitBefund {
				t.Fatalf("%s, Zaun %d: Exit %d %q", e.name, j, code, errText)
			}
			if ist := befundZeilen(t, errText); !gleichListe(ist, []string{soll}) {
				t.Fatalf("%s, Zaun %d: %q, erwartet %q", e.name, j, ist, soll)
			}
			n++
		}
	}
	if n != 20 {
		t.Fatalf("%d Elemente, erwartet 20", n)
	}
}

// Z61 (Nr. 5 b, 5 c: „beginnt mit `Datei: [`“): Vorsatz vor der Bindungszeile {Leerzeichen, zwei Leerzeichen, Tabulator,
// NBSP, `x`, `x `, `- `, `> `, `Datei: `} (9) × {vor einem Zaun, ohne Zaun} = 18. Vor dem Zaun, im Handbuch ohne
// Beispiel-Verzeichnis: genau `Zeile 10: Beispiel ohne Bindungszeile`, nie „Bindung ungültig“; ohne Zaun kein Befund.
func TestHandbuchDateiPraefixAmAnfang(t *testing.T) {
	m := ohneBeispiele(t)
	vorsaetze := []string{" ", "  ", "\t", string(rune(0xA0)), "x", "x ", "- ", "> ", "Datei: "}
	n := 0
	for _, v := range vorsaetze {
		zeile := v + "Datei: [beispiele/aufruf.txt](beispiele/aufruf.txt)\n"
		m.handbuch(rumpf(zeile + zaunVon(aufrufDatei)))
		code, _, errText := m.lauf()
		if code != exitBefund {
			t.Fatalf("%q vor dem Zaun: Exit %d %q", v, code, errText)
		}
		if ist := befundZeilen(t, errText); !gleichListe(ist, []string{"Zeile 10: Beispiel ohne Bindungszeile"}) {
			t.Fatalf("%q vor dem Zaun: %q", v, ist)
		}
		m.handbuch(rumpf(zeile + "Text.\n"))
		if code, _, errText := m.lauf(); code != exitOK {
			t.Fatalf("%q ohne Zaun: Exit %d %q", v, code, errText)
		}
		n += 2
	}
	if n != 18 {
		t.Fatalf("%d Elemente, erwartet 18", n)
	}
}

// Z62 (Nr. 13 b: „kein Buchstabe“): jede der 104 Formen × 26 Nachbarn: Nicht-Buchstabe davor (`0`, `9`, `-`, `/`, `.`,
// `,`, `)`, `€`, Absatzanfang; 9), dasselbe dahinter (9), Buchstabe davor (`x`, `ä`, `ß`, `я`; 4), Buchstabe dahinter (4)
// = 2704, ohne die 8 Elemente „Buchstabe dahinter“ der beiden Kennungsformen von C4 (`adr-0001`, `adr-idx-0004`: der Treffer
// ist `adr`, der Nachbar grenzt nicht an ihn) = 2696. Orakel von Hand: Treffer genau dann, wenn (davor kein Buchstabe oder
// das Muster beginnt mit `\p{L}*`) und (dahinter kein Buchstabe oder das Muster endet auf `\p{L}*`); ein Präfix der Form
// (`in `, `wie `) steht vor dem Nachbarn. Die Funktion läuft direkt.
func TestHandbuchTrefferNachbar(t *testing.T) {
	keinBuchstabe := []string{"0", "9", "-", "/", ".", ",", ")", string(rune(0x20AC)), ""}
	buchstabe := []string{"x", string(rune(0xE4)), string(rune(0xDF)), string(rune(0x44F))}
	n := 0
	for _, p := range orakel() {
		vornOffen := strings.HasPrefix(p.ausdruck, `\p{L}*`)
		hintenOffen := strings.HasSuffix(p.ausdruck, `\p{L}*`)
		for _, form := range p.formen {
			k := kern(form)
			pre := strings.TrimSuffix(form, k)
			pruefe := func(davor, dahinter string, soll bool) {
				prosa := pre + davor + k + dahinter
				if davor == "" {
					prosa = pre + k + dahinter
				}
				bf := mengenBefunde([]absatz{{zeile: 9, zeilen: []string{prosa}}})
				gefunden := false
				for _, b := range bf {
					if strings.HasPrefix(b.text, "Zeile 9"+trefferZeile(p.id)) {
						gefunden = true
					}
				}
				if gefunden != soll {
					t.Fatalf("%s, Form %q mit %q davor und %q dahinter: Treffer %v, erwartet %v", p.id, form, davor, dahinter, gefunden, soll)
				}
				n++
			}
			for _, d := range keinBuchstabe {
				pruefe(d, "", true)
				pruefe("", d, true)
			}
			for _, b := range buchstabe {
				pruefe(b, "", vornOffen)
				if p.id == "C4" && strings.Contains(form, "-") {
					continue
				}
				pruefe("", b, hintenOffen)
			}
		}
	}
	if n != 2696 {
		t.Fatalf("%d Elemente, erwartet 2696", n)
	}
}

// Z63 (Nr. 9 b: „beginnt das zweite Wort mit `-` oder enthält `=`“): der Span `make W build` in einem Satz; Zuweisung (11):
// `X=1`, `x=1`, `VERSION=1`, `a1=1`, `A_B=1`, `a.b=1`, `a-b=1`, `=1`, `x=`, `=`, `x=y=z`; Option (11): `-j`, `-n`, `-C`,
// `-f`, `-s`, `-k`, `--help`, `--version`, `-`, `--`, `-j4`; dazu 5 Gutfälle = 27. Orakel: genau `make-Aufruf ohne Ziel an
// erster Stelle`, nie `unbekannt`. Die Funktion läuft direkt.
func TestHandbuchMakeZuweisungUndOption(t *testing.T) {
	ziele := map[string]bool{"build": true, "docs-check": true, "lint": true}
	zuweisung := []string{"X=1", "x=1", "VERSION=1", "a1=1", "A_B=1", "a.b=1", "a-b=1", "=1", "x=", "=", "x=y=z"}
	option := []string{"-j", "-n", "-C", "-f", "-s", "-k", "--help", "--version", "-", "--", "-j4"}
	n := 0
	for _, w := range append(append([]string{}, zuweisung...), option...) {
		span := "make " + w + " build"
		bf := makeBefunde([]absatz{{zeile: 9, zeilen: []string{"Siehe `" + span + "`."}}}, ziele)
		soll := "Zeile 9: make-Aufruf ohne Ziel an erster Stelle: " + span
		if ist := geordnet(bf); !gleichListe(ist, []string{soll}) {
			t.Fatalf("%q: %q, erwartet %q", w, ist, soll)
		}
		n++
	}
	gut := [][]string{{"Siehe `make build -j`."}, {"Siehe `make build X=1`."}, {"Siehe `make build nix`."}, {"Siehe `make docs-check`."}, {"Siehe `make lint", "X=1`."}}
	for _, g := range gut {
		if ist := geordnet(makeBefunde([]absatz{{zeile: 9, zeilen: g}}, ziele)); len(ist) != 0 {
			t.Fatalf("%q: %q, erwartet kein Befund", g, ist)
		}
		n++
	}
	if n != 27 {
		t.Fatalf("%d Elemente, erwartet 27", n)
	}
}

// Z64 (Nr. 8: „`Makefile` oder `harness/mk/*.mk`“ bzw. „die `Dockerfile`“): 19 Elemente, jedes in einer eigenen Wurzel mit
// einem Eintrag in docs/user/beispiele. Image erlaubt (4): `Makefile`, `harness/mk/{a,b,c}.mk`. Image nicht gedeckt (6):
// die Zeichenfolge nur in `d-check.mk`, `repo.mk`, `harness/mk/sub/x.mk`, `harness/mk/x.txt`, `harness/mk/x.mk.bak`,
// `Dockerfile`. Pfad nicht gedeckt (6): `-o /out/odata-recorder` nur in `Makefile`, `harness/mk/x.mk`, `d-check.mk`,
// `repo.mk`, `Dockerfile.bak`, `docker/Dockerfile`. Eintragsarten (3): `x.txt`, Unterverzeichnis, `notiz.md` als einziger
// Eintrag, das Makefile ohne Image. Die Funktion läuft direkt.
func TestHandbuchAufrufpraefixQuellen(t *testing.T) {
	const image, pfad = "-t odata-recorder:build", "-o /out/odata-recorder"
	neu := func(eintrag string, dateien map[string]string) string {
		m := umfeld{t: t, wurzel: t.TempDir()}
		if eintrag == "sub/" {
			m.schreibe("docs/user/beispiele/sub/y.txt", "")
		} else {
			m.schreibe("docs/user/beispiele/"+eintrag, "")
		}
		for rel, inhalt := range dateien {
			m.schreibe(rel, inhalt)
		}
		return m.wurzel
	}
	texte := func(bf []befund) []string {
		var z []string
		for _, b := range bf {
			z = append(z, b.text)
		}
		return z
	}
	n := 0
	for _, quelle := range []string{"Makefile", "harness/mk/a.mk", "harness/mk/b.mk", "harness/mk/c.mk"} {
		w := neu("x.txt", map[string]string{quelle: image + "\n", "Dockerfile": pfad + "\n"})
		if ist := texte(praefixBefunde(w)); len(ist) != 0 {
			t.Fatalf("Image in %s: %q, erwartet kein Befund", quelle, ist)
		}
		n++
	}
	for _, quelle := range []string{"d-check.mk", "repo.mk", "harness/mk/sub/x.mk", "harness/mk/x.txt", "harness/mk/x.mk.bak", "Dockerfile"} {
		d := map[string]string{"Makefile": "build: ## x\n", "Dockerfile": pfad + "\n", quelle: image + "\n"}
		if quelle == "Dockerfile" {
			d[quelle] = image + "\n" + pfad + "\n"
		}
		if ist := texte(praefixBefunde(neu("x.txt", d))); !gleichListe(ist, []string{"Aufrufpräfix nicht gedeckt: Image"}) {
			t.Fatalf("Image nur in %s: %q", quelle, ist)
		}
		n++
	}
	for _, quelle := range []string{"Makefile", "harness/mk/x.mk", "d-check.mk", "repo.mk", "Dockerfile.bak", "docker/Dockerfile"} {
		d := map[string]string{"Makefile": image + "\n", "Dockerfile": "FROM x\n"}
		d[quelle] = pfad + "\n"
		if quelle == "Makefile" {
			d[quelle] = image + "\n" + pfad + "\n"
		}
		if ist := texte(praefixBefunde(neu("x.txt", d))); !gleichListe(ist, []string{"Aufrufpräfix nicht gedeckt: Pfad"}) {
			t.Fatalf("Pfad nur in %s: %q", quelle, ist)
		}
		n++
	}
	for _, eintrag := range []string{"x.txt", "sub/", "notiz.md"} {
		d := map[string]string{"Makefile": "build: ## x\n", "Dockerfile": pfad + "\n"}
		if ist := texte(praefixBefunde(neu(eintrag, d))); !gleichListe(ist, []string{"Aufrufpräfix nicht gedeckt: Image"}) {
			t.Fatalf("Eintrag %s: %q", eintrag, ist)
		}
		n++
	}
	if n != 19 {
		t.Fatalf("%d Elemente, erwartet 19", n)
	}
}

// Z67 (R3, Nr. 3 b, 3 c): ein `\r` gehört nicht zur Zeile, wenn ein `\n` unmittelbar folgt; am Dateiende ohne folgendes
// `\n` ist es ein Zeichen der letzten Zeile. Menge: sieben Formen des Architect (`x`, `x\n`, `x\r\n` ohne Befund; `x\r`,
// `x\r\r\n`, `x\r\r`, `x\ry\r\n` mit je einem Befund U+000D) und drei Formen am Zeilenende vor und nach einer weiteren
// Zeile (`a\n\r`, `\r`, `a\r\nb\r`): 10 Elemente. Orakel: die Zeilen von Hand.
func TestHandbuchCRAmDateiende(t *testing.T) {
	faelle := []struct {
		in      string
		zeilen  []string
		befunde []string
	}{
		{"x", []string{"x"}, nil},
		{"x\n", []string{"x"}, nil},
		{"x\r\n", []string{"x"}, nil},
		{"x\r", []string{"x\r"}, []string{"Zeile 1: Zeichen U+000D nicht zulässig"}},
		{"x\r\r\n", []string{"x\r"}, []string{"Zeile 1: Zeichen U+000D nicht zulässig"}},
		{"x\r\r", []string{"x\r\r"}, []string{"Zeile 1: Zeichen U+000D nicht zulässig"}},
		{"x\ry\r\n", []string{"x\ry"}, []string{"Zeile 1: Zeichen U+000D nicht zulässig"}},
		{"a\n\r", []string{"a", "\r"}, []string{"Zeile 2: Zeichen U+000D nicht zulässig"}},
		{"\r", []string{"\r"}, []string{"Zeile 1: Zeichen U+000D nicht zulässig"}},
		{"a\r\nb\r", []string{"a", "b\r"}, []string{"Zeile 2: Zeichen U+000D nicht zulässig"}},
	}
	if len(faelle) != 10 {
		t.Fatalf("%d Elemente, erwartet 10", len(faelle))
	}
	for _, f := range faelle {
		zeilen := zeilenVon(f.in)
		if fmt.Sprintf("%q", zeilen) != fmt.Sprintf("%q", f.zeilen) {
			t.Fatalf("%q: Zeilen %q, erwartet %q", f.in, zeilen, f.zeilen)
		}
		var ist []string
		for _, b := range zeichenBefunde(zeilen) {
			ist = append(ist, b.text)
		}
		if !gleichListe(ist, f.befunde) {
			t.Fatalf("%q: Befunde %q, erwartet %q", f.in, ist, f.befunde)
		}
	}
}
