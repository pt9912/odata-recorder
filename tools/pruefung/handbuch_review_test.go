package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Tests zu den Befunden des Reviews und den Entscheidungen des Architect danach (SPEC-200).

const nbsp = "\xc2\xa0"

// Z44: ein vorhandener Eintrag, den man nicht lesen kann (ein Verzeichnis), ist `Handbuch nicht lesbar`, nie `Handbuch fehlt`.
func TestHandbuchVerzeichnisAlsHandbuch(t *testing.T) {
	m := neuesUmfeld(t, standardAntworten())
	m.entferne(handbuchPfad)
	if err := os.MkdirAll(filepath.Join(m.wurzel, handbuchPfad), 0o755); err != nil {
		t.Fatal(err)
	}
	code, _, errText := m.lauf()
	erwarte(t, code, errText, exitAufruf, "pruefung handbuch: Eingabe: Handbuch nicht lesbar: ")
	nichtEnthalten(t, errText, "Handbuch fehlt")
	if n := len(m.e.starts(t)); n != 0 || strings.Count(errText, "\n") != 1 {
		t.Fatalf("%d Starts, Ausgabe %q", n, errText)
	}
}

// Z45: scheitert das Anlegen des Verzeichnisses eines Laufs, ist es der Eingabefehler `Arbeitsverzeichnis nicht
// anlegbar`, an beiden Aufrufstellen (version über pruefeDokument, Beispiel über pruefeBeispiele).
func TestHandbuchLaufVerzeichnisNichtAnlegbar(t *testing.T) {
	m := neuesUmfeld(t, standardAntworten())
	fehlt := filepath.Join(t.TempDir(), "gibt-es-nicht")
	_, eingabe := pruefeDokument(m.u, m.e.bin, m.wurzel, fehlt, zeilenVon(sauber()))
	if len(eingabe) != 1 || !strings.HasPrefix(eingabe[0], "Arbeitsverzeichnis nicht anlegbar: ") {
		t.Fatalf("version: %q", eingabe)
	}
	_, eingabe = pruefeBeispiele(m.u, m.e.bin, m.wurzel, fehlt, []string{"beispiele/aufruf.txt"})
	if len(eingabe) != 1 || !strings.HasPrefix(eingabe[0], "Arbeitsverzeichnis nicht anlegbar: ") {
		t.Fatalf("Beispiel: %q", eingabe)
	}
	if n := len(m.e.starts(t)); n != 0 {
		t.Fatalf("%d Starts", n)
	}
}

// Z46: eine Kopfzeile mit leerem Schlüssel nach dem Trimmen ist `Zeile ohne Schlüssel`, nie `unbekannt`.
func TestHandbuchKopfLeererSchluessel(t *testing.T) {
	m := ohneBeispiele(t)
	for _, zeile := range []string{": x", "   : x", nbsp + ": x", ":"} {
		m.handbuch(kopfSauber + zeile + "\n\n## A\n")
		code, _, errText := m.lauf()
		erwarte(t, code, errText, exitBefund, "Zeile 6: Kopf: Zeile ohne Schlüssel")
		nichtEnthalten(t, errText, "unbekannt")
		if c := strings.Count(errText, "\n"); c != 1 {
			t.Fatalf("%q: %d Befunde", zeile, c)
		}
	}
}

// Z47: ein Span, der bis auf Leerraum `make` ist, ist kein Aufruf; Leerraum ist die Menge der 19 Zeichen (Nr. 3 b)
// und das Leerzeichen. Ein Span `make` und ein anderes Trennzeichen als U+0020 ist kein Aufruf (Nr. 14 d), `make nix`
// ist einer. Elemente: 3 Spans mit Leerzeichen + `makefile x`, `make-x` + 19 × 3 (`make`+Zeichen, `make `+Zeichen,
// Zeichen als Trenner vor `nix`) + `make nix` = 63.
func TestHandbuchMakeSpanOhneZweitesWort(t *testing.T) {
	m := ohneBeispiele(t)
	zielDateien(m)
	keinBefund := []string{"make", "make ", "make  ", "makefile x", "make-x"}
	for _, c := range leerraum19() {
		keinBefund = append(keinBefund, "make"+c, "make "+c, "make"+c+"nix")
	}
	n := 0
	for _, text := range keinBefund {
		m.handbuch(rumpf("Siehe `" + text + "` hier.\n"))
		if code, _, errText := m.lauf(); code != exitOK {
			t.Fatalf("%q: Exit %d %q", text, code, errText)
		}
		n++
	}
	m.handbuch(rumpf("Siehe `make nix` hier.\n"))
	code, _, errText := m.lauf()
	erwarte(t, code, errText, exitBefund, "make-Ziel nix unbekannt")
	if n+1 != 63 {
		t.Fatalf("%d Elemente, erwartet 63", n+1)
	}
}

// Z21 (Rand): eine Ziffernfolge, deren Wert in einer 64-Bit-Zahl überläuft, ist kein Ausgangswert (2^64 ist kein Ausgangswert).
func TestHandbuchAusgangswertUeberlauf(t *testing.T) {
	folgen := []string{"18446744073709551616", "18446744073709551621", "99999999999999999999", "4294967296", "1000", "100000"}
	for _, z := range folgen {
		erwarteUngueltig(t, "k.txt", bspDatei(praef, nil, nil, z+"\n"))
	}
}

// Z15 (Rand): Text hinter der Klammer der Bindungszeile und eine Endung jenseits von .txt sind Verstöße der Form.
func TestHandbuchBindungszeileRestUndEndung(t *testing.T) {
	zaun := zaunVon(aufrufDatei)
	m := neuesUmfeld(t, standardAntworten())
	b := "Datei: [beispiele/aufruf.txt](beispiele/aufruf.txt)"
	var zeilen []string
	for _, rest := range []string{" x", "x", " ", ")", "]", nbsp} {
		zeilen = append(zeilen, b+rest)
	}
	for _, name := range []string{"a.tat", "a.txx", "a.tx", "a.txtx", "a.txt.md", "a.TXT", "a.txt "} {
		zeilen = append(zeilen, "Datei: [beispiele/"+name+"](beispiele/"+name+")")
	}
	for _, z := range zeilen {
		m.handbuch(rumpf(z + "\n" + zaun))
		code, _, errText := m.lauf()
		erwarte(t, code, errText, exitBefund, "Bindung ungültig: "+z)
	}
}

// Z10 (Form): die Ausgabe von version hat sechs Teile; fehlt einer, ist es `Ausgabe nicht lesbar` und kein Vergleich.
func TestHandbuchVersionForm(t *testing.T) {
	ausgaben := []string{
		"odata-recorder dev", "odata-recorder \n", "odata-recorder  dev\n", "odata-recorder dev\n\n",
		"odata-recorder\tdev\n", "odata-recorder dev \n", "odata-recorder" + nbsp + "dev\n",
	}
	for _, out := range ausgaben {
		m := neuesUmfeld(t, map[string]antwort{"version": {out: out}, "": {out: "odata-recorder dev\n"}})
		m.entferne("docs/user/beispiele")
		m.handbuch(rumpf("Text.\n"))
		code, _, errText := m.lauf()
		erwarte(t, code, errText, exitBefund, "version: Ausgabe nicht lesbar: ")
		nichtEnthalten(t, errText, "weicht von der Ausgabe")
	}
}

// Z17 (Block 1 bis 3): jeder Block, der eine Datei bindet, gleicht ihr; Abweichungsart × Stelle des Blocks.
func TestHandbuchBloeckeJeDatei(t *testing.T) {
	abw := []string{
		strings.Replace(aufrufDatei, "odata-recorder dev", "odata-recorder dex", 1),
		aufrufDatei + "x\n",
		strings.TrimSuffix(aufrufDatei, "0\n"),
		aufrufDatei + "\n",
		strings.TrimSuffix(aufrufDatei, "\n") + " \n",
		strings.Replace(aufrufDatei, "--- Standardausgabe", "--- standardausgabe", 1),
	}
	m := neuesUmfeld(t, standardAntworten())
	b := binde("beispiele/aufruf.txt")
	n := 0
	for _, a := range abw {
		for stelle := 0; stelle < 3; stelle++ {
			var h strings.Builder
			h.WriteString(rumpf(""))
			for i := 0; i < 3; i++ {
				blk := aufrufDatei
				if i == stelle {
					blk = a
				}
				h.WriteString(b + zaunVon(blk) + "\n")
			}
			m.handbuch(h.String())
			code, _, errText := m.lauf()
			if c := strings.Count(errText, "Beispiel weicht von beispiele/aufruf.txt ab"); code != exitBefund || c != 1 {
				t.Fatalf("Abweichung %q im Block %d: Exit %d, %d Befunde: %q", a, stelle+1, code, c, errText)
			}
			n++
		}
	}
	if n != 18 {
		t.Fatalf("%d Elemente, erwartet 18", n)
	}
}

// Z50: „getrimmt“ (Nr. 3 b) gilt an sechs Stellen: Wert und Schlüssel des Kopfes, leere Zeile des Kopfes, Absatzgrenze,
// Schluss-Zaun und Info-Zeichenkette, jede mit allen 19 Leerraum-Zeichen (leerraum19; 114 Elemente).
func TestHandbuchGetrimmtAnSechsStellen(t *testing.T) {
	m := neuesUmfeld(t, standardAntworten())
	n := 0
	for _, c := range leerraum19() {
		faelle := map[string]string{
			"Kopf-Wert":       strings.Replace(sauber(), "Stand: 2026-10-30", "Stand:"+c+"2026-10-30"+c, 1),
			"Kopf-Schlüssel":  strings.Replace(sauber(), "Stand: 2026-10-30", "Stand"+c+": 2026-10-30", 1),
			"leere Kopfzeile": strings.Replace(sauber(), "Software-Version: dev\n", "Software-Version: dev\n"+c+"\n", 1),
			"Absatzgrenze":    sauber() + "\nDas ist noch\n" + c + "\nnicht hier.\n",
			"Schluss-Zaun":    rumpf(binde("beispiele/aufruf.txt") + "```\n" + aufrufDatei + "```" + c + "\n"),
			"Info":            rumpf(binde("beispiele/aufruf.txt") + "```" + c + "\n" + aufrufDatei + "```\n"),
		}
		if len(faelle) != 6 {
			t.Fatalf("%d Stellen, erwartet 6", len(faelle))
		}
		for name, h := range faelle {
			m.handbuch(h)
			if code, _, errText := m.lauf(); code != exitOK {
				t.Fatalf("%s mit %q: Exit %d %q", name, c, code, errText)
			}
			n++
		}
	}
	if n != 114 {
		t.Fatalf("%d Elemente, erwartet 114", n)
	}
}

// Z38 (innerhalb einer Nummer): nach Zeile aufsteigend, dann nach Name; die Umkehr jedes Schlüssels ist rot.
func TestHandbuchReihenfolgeInnerhalbEinerNummer(t *testing.T) {
	m := ohneBeispiele(t)
	m.handbuch(rumpf("Das ist geplant.\n\nDas ist künftig.\n"))
	_, _, errText := m.lauf()
	i, j := strings.Index(errText, "Planung A1"), strings.Index(errText, "Planung A4")
	if i < 0 || j < 0 || i > j {
		t.Fatalf("Zeilen nicht aufsteigend: %q", errText)
	}
	m = neuesUmfeld(t, standardAntworten())
	m.entferne("docs/user/beispiele/aufruf.txt")
	m.schreibe("docs/user/beispiele/a.txt", bspDatei(praef, []string{"x"}, nil, "0\n"))
	m.schreibe("docs/user/beispiele/b.txt", bspDatei(praef+" foo", []string{"y"}, nil, "0\n"))
	m.handbuch(rumpf(binde("beispiele/b.txt") + zaunVon(bspDatei(praef+" foo", []string{"y"}, nil, "0\n")) + "\n" +
		binde("beispiele/a.txt") + zaunVon(bspDatei(praef, []string{"x"}, nil, "0\n"))))
	_, _, errText = m.lauf()
	i, j = strings.Index(errText, "a.txt: Standardausgabe weicht ab"), strings.Index(errText, "b.txt: Standardausgabe weicht ab")
	if i < 0 || j < 0 || i > j {
		t.Fatalf("Namen nicht aufsteigend: %q", errText)
	}
}

// Z49 (Nr. 7 e): ein Kindprozess von 6 s, der die Ströme nach dem Ende des Binary offen hält, hält den Lauf höchstens
// eine Sekunde auf; der Test fängt eine Wartezeit ab 2,5 s je Lauf (ein WaitDelay zwischen 1 s und 2,5 s bleibt grün, weil
// die Last des Prüfhosts keine engere Grenze trägt). Ausgabe und Ausgangswert gelten. Jede Aufrufstelle für sich.
func TestHandbuchKindprozessHaeltDenStrom(t *testing.T) {
	ok := antwort{out: "odata-recorder dev\n"}
	mitKind := ok
	mitKind.kind = "6"
	for _, stelle := range []string{"version", "Beispiel"} {
		a := map[string]antwort{"version": ok, "": ok}
		if stelle == "version" {
			a["version"] = mitKind
		} else {
			a[""] = mitKind
		}
		m := neuesUmfeld(t, a)
		t0 := time.Now()
		code, _, errText := m.lauf()
		if dauer := time.Since(t0); dauer > 2500*time.Millisecond {
			t.Fatalf("%s: Dauer %v mit Kindprozess von 6 s", stelle, dauer)
		}
		if code != exitOK {
			t.Fatalf("%s: Exit %d %q", stelle, code, errText)
		}
	}
}

// F-6: das Wort ADR trifft C4 an den sieben Stellen und in der Schreibung der Titel; Adresse trifft nicht.
func TestHandbuchAdrWort(t *testing.T) {
	m := neuesUmfeld(t, standardAntworten())
	n := 0
	for _, form := range []string{"ADR", "ADRs", "ADR-Index", "Adr", "adrs"} {
		for name, h := range stellungen(form) {
			m.handbuch(h)
			code, _, errText := m.lauf()
			if code != exitBefund || !strings.Contains(errText, trefferZeile("C4")+"adr") {
				t.Fatalf("%q an der Stelle %s: Exit %d %q", form, name, code, errText)
			}
			n++
		}
	}
	for _, form := range []string{"Adresse", "Adressen", "Adressat", "Madrid"} {
		for name, h := range stellungen(form) {
			m.handbuch(h)
			if code, _, errText := m.lauf(); code != exitOK {
				t.Fatalf("%q an der Stelle %s: Exit %d %q", form, name, code, errText)
			}
			n++
		}
	}
	if n != 63 {
		t.Fatalf("%d Elemente, erwartet 63", n)
	}
}
