package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const (
	dExtra = "src/bootstrap/konfig_p.go"
	paket  = "package bootstrap\n\n"
)

func plus(l []string, x ...string) []string { return append(append([]string(nil), l...), x...) }

func art4(name string, zeile int) string {
	return fmt.Sprintf("%s in konfig*.go (Zeile %d), nicht in SPEC-110", name, zeile)
}

// Die 27 Schlüssel von SPEC-110 Nr. 2; `SERVICE_ODATA_VERSION` hat noch keinen Leser, C sind die übrigen 26.
func schluessel27() []string {
	return []string{"MODE", "LISTEN", "DATA_DIR", "SERVICE_ID", "SERVICE_ROOT_PATH", "UPSTREAM_URL", "SERVICE_ODATA_VERSION", "SERVICES_FILE",
		"REPLAY_STRATEGY", "REPLAY_ASSIGNMENT", "REPLAY_REDACTED_BODIES", "MATCH_VARY_HEADERS", "MATCH_INVALID_ENCODING", "MATCH_BODY_MODE",
		"REDACT_HEADERS", "REDACT_QUERY_PARAMS", "REDACT_JSON_PATHS", "MAX_CAPTURE_BODY", "UPSTREAM_TIMEOUT", "UPSTREAM_CA_FILE",
		"UPSTREAM_TLS_VERIFY", "UPSTREAM_TLS_INSECURE_SKIP_VERIFY", "SHUTDOWN_TIMEOUT", "RETENTION_MAX_AGE", "SQLITE_BUSY_TIMEOUT",
		"VOLATILE_HEADERS", "LINK_REWRITE"}
}

func schluessel26() []string { return ohne(schluessel27(), "SERVICE_ODATA_VERSION") }

// positionen sind die 30 Stellen, an denen ein Literal in `konfig*.go` stehen kann.
func positionen() []string {
	return []string{
		`func f() { g("NAME_X", 1, 2, 3) }`, `func f() { g(1, "NAME_X", 2, 3) }`, `func f() { g(1, 2, 3, "NAME_X") }`,
		`func f() { _ = p.env["NAME_X"] }`, `var l = []string{"NAME_X", "b", "c"}`, `var l = []string{"a", "b", "NAME_X"}`,
		`var m = map[string]int{"NAME_X": 1}`, `var m = map[int]string{1: "NAME_X"}`, `var s = T{F: "NAME_X"}`, `var s = T{"NAME_X", 1}`,
		`var l = [][]string{{"a"}, {"b", "NAME_X"}}`, `func f() string { return "NAME_X" }`, `func f() (string, error) { return "NAME_X", nil }`,
		`func f() { var x string; x = "NAME_X"; _ = x }`, `func f() { x := "NAME_X"; _ = x }`, `var x = "NAME_X"`,
		`func f() { var x = "NAME_X"; _ = x }`, `const x = "NAME_X"`, `func f() { const x = "NAME_X"; _ = x }`,
		`func f(s string) { switch s { case "NAME_X": } }`, `func f(s string) bool { return s == "NAME_X" }`, `func f(s string) string { return s + "NAME_X" }`,
		`var f = func() string { return "NAME_X" }`, `func f() { defer g("NAME_X") }`, `func f(c chan string) { c <- "NAME_X" }`,
		`func f() { _ = []byte("NAME_X") }`, "var x = `NAME_X`", `var x = "\x4eAME_X"`, `func (p *T) f() string { return "NAME_X" }`,
		"func f() {\n\tg(\n\t\t1,\n\t\t\"NAME_X\",\n\t)\n}",
	}
}

// TestKatSchlPositionen: jedes Literal der Form ist an jeder Stelle von `konfig*.go` ein Name (SPEC-201 Nr. 4 a). Menge 30
// Positionen, je in einer eigenen Datei: ohne den Namen in SPEC-110 genau Art 4 an der Zeile des Literals, mit dem Namen
// in SPEC-110 und im Handbuch `ok`.
func TestKatSchlPositionen(t *testing.T) {
	ps := positionen()
	zaehle(t, "Positionen", len(ps), 30)
	for i, p := range ps {
		src := paket + p + "\n"
		name := fmt.Sprintf("Position %d", i+1)
		rot := baumSchl(schluesselBasis(), schluesselBasis(), plus(schluesselBasis(), "NAME_X"))
		rot[dExtra] = src
		erwarteAus(t, name, rot, 1, katZeilen(katS, art4("NAME_X", zeileVon(t, src, "AME_X"))))
		gruen := baumSchl(schluesselBasis(), plus(schluesselBasis(), "NAME_X"), plus(schluesselBasis(), "NAME_X"))
		gruen[dExtra] = src
		erwarteOk(t, name+" grün", gruen)
	}
}

// TestKatSchlForm: ein Literal ist ein Name, wenn es ganz `^[A-Z][A-Z0-9_]+$` trifft (ASCII, mindestens zwei Zeichen).
// Menge 16 = 6 gültige und 10 ungültige Werte; die letzten beiden (MÖDE, MODE in Vollbreite) entstehen aus Bytewerten.
func TestKatSchlForm(t *testing.T) {
	gueltig := []string{"MO", "M1", "MODE_", "A_B", "X9_Y", "ZZ"}
	ungueltig := []string{"M", "mode", "Mode", "1MODE", "_MODE", "MODE-X", "MODE X", "MODE\n", "M\xc3\x96DE", "\xef\xbc\xad\xef\xbc\xaf\xef\xbc\xa4\xef\xbc\xa5"}
	zaehle(t, "Werte", len(gueltig)+len(ungueltig), 16)
	for _, n := range gueltig {
		m := baumSchl(schluesselBasis(), schluesselBasis(), plus(schluesselBasis(), n))
		m[dExtra] = paket + "var x = `" + n + "`\n"
		erwarteAus(t, n, m, 1, katZeilen(katS, art4(n, 3)))
	}
	for _, n := range ungueltig {
		m := baumSchl(schluesselBasis(), schluesselBasis(), schluesselBasis())
		m[dExtra] = paket + "var x = `" + n + "`\n"
		w := schreibe(t, m)
		liesZurueck(t, w, dExtra, "var x = `"+n+"`\n")
		c, o, e := fahre(w)
		gleichFahrt(t, fmt.Sprintf("%q", n), c, o, e, 0, "katalog: ok\n", "")
	}
}

// TestKatSchlGrenze: ein Name im Körper jeder Deklaration `func nichtGelieferte` ist N, außerhalb davon C (Nr. 4 a). Menge 14
// Fälle (a bis n); `AUSSEN_X` mit Zeile im Handbuch ist C ohne K (Art 4), `INNEN_X` ist N: seine Zeile im Handbuch
// ist Art 10, und er ist nie in C (kein Art 4, kein Art 8).
func TestKatSchlGrenze(t *testing.T) {
	faelle := []struct {
		name, src, src2 string
		innen           []string // Namen in N: die Tabelle nennt sie, erwartet Art 10
		aussen, zeile   string   // Name in C ohne K (Art 4) und die Nadel seiner Zeile
	}{
		{"a Rückgabe einer Liste", `func nichtGelieferte() []string { return []string{"INNEN_X"} }`, "", []string{"INNEN_X"}, "", ""},
		{"b Argument", `func nichtGelieferte() { g("INNEN_X") }`, "", []string{"INNEN_X"}, "", ""},
		{"c Funktionsliteral", "func nichtGelieferte() {\n\tf := func() string { return \"INNEN_X\" }\n\t_ = f\n}", "", []string{"INNEN_X"}, "", ""},
		{"d Empfänger", `func (p *T) nichtGelieferte() []string { return []string{"INNEN_X"} }`, "", []string{"INNEN_X"}, "", ""},
		{"e Typparameter", `func nichtGelieferte[T any]() []string { return []string{"INNEN_X"} }`, "", []string{"INNEN_X"}, "", ""},
		{"f zwei Dateien", `func nichtGelieferte() { g("INNEN_X") }`, `func nichtGelieferte() { g("INNEN_Y") }`, []string{"INNEN_X", "INNEN_Y"}, "", ""},
		{"g Körper und außerhalb", "func nichtGelieferte() { g(\"BEIDE_X\") }\n\nfunc other() { g(\"BEIDE_X\") }", "", []string{"BEIDE_X"}, "BEIDE_X", "func other"},
		{"h davor", "func vor() { g(\"AUSSEN_X\") }\n\nfunc nichtGelieferte() { g(\"INNEN_X\") }", "", []string{"INNEN_X"}, "AUSSEN_X", "AUSSEN_X"},
		{"i dahinter", `func nichtGelieferte() { g("INNEN_X") }; var x = "AUSSEN_X"`, "", []string{"INNEN_X"}, "AUSSEN_X", "AUSSEN_X"},
		{"j ohne e", `func nichtGeliefert() { g("AUSSEN_X") }`, "", nil, "AUSSEN_X", "AUSSEN_X"},
		{"k großes N", `func NichtGelieferte() { g("AUSSEN_X") }`, "", nil, "AUSSEN_X", "AUSSEN_X"},
		{"l Variable", `var nichtGelieferte = func() { g("AUSSEN_X") }`, "", nil, "AUSSEN_X", "AUSSEN_X"},
		{"m ohne Körper", "func nichtGelieferte()\n\nvar x = \"AUSSEN_X\"", "", nil, "AUSSEN_X", "AUSSEN_X"},
		{"n ohne Deklaration", `var x = "AUSSEN_X"`, "", nil, "AUSSEN_X", "AUSSEN_X"},
	}
	zaehle(t, "Fälle", len(faelle), 14)
	for _, f := range faelle {
		w := plus(schluesselBasis(), f.innen...)
		if f.aussen == "AUSSEN_X" {
			w = plus(w, "AUSSEN_X")
		}
		m := baumSchl(schluesselBasis(), schluesselBasis(), w)
		m[dKonfig] = konfigLade(schluesselBasis())
		m["src/bootstrap/konfig_g.go"] = paket + f.src + "\n"
		if f.src2 != "" {
			m["src/bootstrap/konfig_h.go"] = paket + f.src2 + "\n"
		}
		var soll []string
		if f.aussen != "" {
			soll = append(soll, art4(f.aussen, zeileVon(t, m["src/bootstrap/konfig_g.go"], f.zeile)))
		}
		for _, n := range f.innen { // Art 10 steht hinter Art 4, die Namen sind in Bytefolge genannt
			soll = append(soll, fmt.Sprintf("Handbuch Zeile %d nennt %s, den das Binary ablehnt", zeileVon(t, m[dHandbuch], "`"+n+"`"), n))
		}
		if soll == nil {
			erwarteOk(t, f.name, m)
			continue
		}
		erwarteAus(t, f.name, m, 1, katZeilen(katS, soll...))
	}
}

// TestKatSchlAbgelehnt: ein Name der Tabelle, der ein Name aus N ist (Gleichheit) oder an einem Präfix aus N beginnt, ist
// Art 10 und nie Art 8; ein Name außerhalb von C und N ist Art 8, auch einer aus K ohne C (Nr. 4 a, 4 c, 7). Menge 24 =
// 6 Elemente von N (4 Präfixe, 2 Namen) × 3 Formen (selbst, mit Anhang `_X`, um ein Zeichen gekürzt) + 6 Art-8-Namen
// (K ohne C, in keiner Quelle; je an der ersten, mittleren und letzten Stelle der Tabelle).
func TestKatSchlAbgelehnt(t *testing.T) {
	praefixe, namen := []string{"ADMIN_", "HYBRID_", "FALLBACK_", "LISTEN_TLS_"}, []string{"STRICT_COVERAGE", "PUBLIC_BASE_URL"}
	konfig := konfigLade(schluesselBasis()) + "\nfunc nichtGelieferte() (praefixe, namen []string) {\n\treturn []string{\"ADMIN_\", \"HYBRID_\", \"FALLBACK_\", \"LISTEN_TLS_\"}, []string{\"STRICT_COVERAGE\", \"PUBLIC_BASE_URL\"}\n}\n"
	abgelehnt := "Handbuch Zeile %d nennt %s, den das Binary ablehnt"
	unbekannt := "Handbuch Zeile %d nennt %s, der nicht in konfig*.go steht"
	n := 0
	pruefe := func(name string, w []string, text string) {
		m := baumSchl(schluesselBasis(), plus(schluesselBasis(), "SERVICE_ODATA_VERSION"), w)
		m[dKonfig] = konfig
		z := zeileVon(t, m[dHandbuch], "`"+name+"`")
		erwarteAus(t, name, m, 1, katZeilen(katS, fmt.Sprintf(text, z, name)))
		n++
	}
	for _, p := range praefixe {
		pruefe(p, plus(schluesselBasis(), p), abgelehnt)
		pruefe(p+"_X", plus(schluesselBasis(), p+"_X"), abgelehnt)
		pruefe(p[:len(p)-1], plus(schluesselBasis(), p[:len(p)-1]), unbekannt)
	}
	for _, e := range namen {
		pruefe(e, plus(schluesselBasis(), e), abgelehnt)
		pruefe(e+"_X", plus(schluesselBasis(), e+"_X"), unbekannt)
		pruefe(e[:len(e)-1], plus(schluesselBasis(), e[:len(e)-1]), unbekannt)
	}
	for _, x := range []string{"SERVICE_ODATA_VERSION", "GAR_NICHT"} {
		pruefe(x, plus([]string{x}, schluesselBasis()...), unbekannt)
		pruefe(x, append(plus(schluesselBasis()[:1], x), schluesselBasis()[1:]...), unbekannt)
		pruefe(x, plus(schluesselBasis(), x), unbekannt)
	}
	zaehle(t, "Fälle", n, 24)
}

// TestKatSchlDateien: gelesen wird ein Eintrag, der eine Datei ist, mit dem Namen `konfig`, beliebigem Rest und `.go`, nicht
// auf `_test.go`, nicht rekursiv, in Bytefolge der Namen; die Zeile eines Namens an mehreren Stellen ist die des ersten
// Vorkommens (Nr. 4 a). Menge 13 = 5 gelesene + 6 nicht gelesene Dateinamen, ein Verzeichnis `konfig_d.go`, ein Verweis auf
// eine Datei; dazu derselbe Name in zwei Dateien in 2 Reihenfolgen.
func TestKatSchlDateien(t *testing.T) {
	text := paket + "var x = \"NAME_F\"\n"
	gelesen := map[string]string{
		"konfig_x.go": text, "konfigx.go": text, "konfig_test_x.go": text,
		"konfig_b.go": "//go:build ignore\n\n" + text, "konfig.go": konfigGo(plus(schluesselBasis(), "NAME_F")),
	}
	nicht := []string{"Konfig.go", "konfig_test.go", "konfig_x_test.go", "konfig.go.bak", "xkonfig.go", "unter/konfig.go", "konfig_d.go/x.go"}
	zaehle(t, "Dateien", len(gelesen)+len(nicht), 12)
	rot := func(name string, m map[string]string, zeile int) {
		erwarteAus(t, name, m, 1, katZeilen(katS, art4("NAME_F", zeile)))
	}
	for d, inhalt := range gelesen {
		m := baumSchl(schluesselBasis(), schluesselBasis(), plus(schluesselBasis(), "NAME_F"))
		m["src/bootstrap/"+d] = inhalt
		rot(d, m, zeileVon(t, inhalt, "NAME_F"))
	}
	for _, d := range nicht {
		m := baumSchl(schluesselBasis(), schluesselBasis(), schluesselBasis())
		m["src/bootstrap/"+d] = text
		erwarteOk(t, d, m)
	}
	// Ein Verweis auf eine Datei gilt als sein Ziel.
	m := baumSchl(schluesselBasis(), schluesselBasis(), plus(schluesselBasis(), "NAME_F"))
	m["src/ziel.txt"] = text
	w := schreibe(t, m)
	if err := os.Symlink(filepath.Join(w, "src", "ziel.txt"), filepath.Join(w, "src", "bootstrap", "konfig_l.go")); err != nil {
		t.Fatal(err)
	}
	c, o, e := fahre(w)
	gleichFahrt(t, "Verweis", c, o, e, 1, "", katZeilen(katS, art4("NAME_F", 3)))
	// Ein Verweis ins Leere ist ein Eingabefehler dieser Datei.
	w = schreibe(t, baumSchl(schluesselBasis(), schluesselBasis(), schluesselBasis()))
	if err := os.Symlink("nirgends", filepath.Join(w, "src", "bootstrap", "konfig_l.go")); err != nil {
		t.Fatal(err)
	}
	c, o, e = fahre(w)
	gleichFahrt(t, "Verweis ins Leere", c, o, e, 2, "", eingaben("konfig*.go nicht lesbar: konfig_l.go: no such file or directory"))
	// Derselbe Name in zwei Dateien: die Zeile ist die des ersten Vorkommens in der Bytefolge der Dateinamen.
	vorn, hinten := paket+"var x = \"NAME_F\"\n", paket+"\n\nvar x = \"NAME_F\"\n"
	for i, paar := range [][2]string{{vorn, hinten}, {hinten, vorn}} {
		m := baumSchl(schluesselBasis(), schluesselBasis(), plus(schluesselBasis(), "NAME_F"))
		m["src/bootstrap/konfig_a.go"], m["src/bootstrap/konfig_b.go"] = paar[0], paar[1]
		rot(fmt.Sprintf("zwei Dateien, Reihenfolge %d", i+1), m, zeileVon(t, paar[0], "NAME_F"))
	}
}

// TestKatQuelleEingabe: Eingabefehler von `konfig*.go` und `cli.go` (Nr. 4 a, 5 a, 8): Exit 2, Text bytegleich, kein Befund.
// Menge 16 = konfig 10 (Verzeichnis fehlt; `src/bootstrap` ist eine Datei; kein Go in 3 Formen; zwei Dateien kein Go; kein
// Schlüssel in 3 Formen; eine Datei kein Go bei sonst keinem Namen) + cli 6 (fehlt, Verzeichnis, kein Go in 2 Formen,
// kein Kommando in 2 Formen).
func TestKatQuelleEingabe(t *testing.T) {
	klammer, literal, ohnePaket := paket+"func f() {\n", paket+"var x = \"abc\n", "func f() {}\n"
	nurN := paket + konfigN
	kg := func(d string, zeile int) string {
		return fmt.Sprintf("konfig*.go nicht lesbar: %s: kein Go (Zeile %d)", d, zeile)
	}
	faelle := []struct {
		name string
		bau  func(m map[string]string)
		soll []string
	}{
		{"konfig: Verzeichnis fehlt", func(m map[string]string) { delete(m, dKonfig) }, []string{"konfig*.go fehlt: src/bootstrap"}},
		{"konfig: Verzeichnis ist eine Datei", func(m map[string]string) { delete(m, dKonfig); m["src/bootstrap"] = "x" }, []string{"konfig*.go fehlt: src/bootstrap"}},
		{"konfig: Klammer fehlt", func(m map[string]string) { m[dKonfig] = klammer }, []string{kg("konfig.go", 3)}},
		{"konfig: Literal ohne Ende", func(m map[string]string) { m[dKonfig] = literal }, []string{kg("konfig.go", 3)}},
		{"konfig: ohne Paketzeile", func(m map[string]string) { m[dKonfig] = ohnePaket }, []string{kg("konfig.go", 1)}},
		{"konfig: zwei Dateien kein Go", func(m map[string]string) {
			m["src/bootstrap/konfig_b.go"], m["src/bootstrap/konfig_a.go"] = ohnePaket, klammer
		}, []string{kg("konfig_a.go", 3), kg("konfig_b.go", 1)}},
		{"konfig: Datei ohne Deklaration", func(m map[string]string) { m[dKonfig] = "package bootstrap\n" }, []string{"konfig*.go: kein Schlüssel gefunden"}},
		{"konfig: nur Kommentare", func(m map[string]string) { m[dKonfig] = paket + "// MODE\n/* LISTEN */\n" }, []string{"konfig*.go: kein Schlüssel gefunden"}},
		{"konfig: nur Namen aus N", func(m map[string]string) { m[dKonfig] = nurN }, []string{"konfig*.go: kein Schlüssel gefunden"}},
		{"konfig: eine Datei kein Go, sonst kein Name", func(m map[string]string) {
			m[dKonfig] = nurN
			m["src/bootstrap/konfig_b.go"] = ohnePaket
		}, []string{kg("konfig_b.go", 1)}},
		{"cli: fehlt", func(m map[string]string) { delete(m, dCli) }, []string{"cli.go fehlt: src/adapters/driving/cli/cli.go"}},
		{"cli: Verzeichnis", func(m map[string]string) { delete(m, dCli); m[dCli+"/x"] = "" }, []string{"cli.go nicht lesbar: is a directory"}},
		{"cli: Klammer fehlt", func(m map[string]string) { m[dCli] = "package cli\n\nfunc f() {\n" }, []string{"cli.go nicht lesbar: kein Go (Zeile 3)"}},
		{"cli: Literal ohne Ende", func(m map[string]string) { m[dCli] = "package cli\n\nvar x = \"abc\n" }, []string{"cli.go nicht lesbar: kein Go (Zeile 3)"}},
		{"cli: leere Datei", func(m map[string]string) { m[dCli] = "package cli\n" }, []string{"cli.go: kein Kommando gefunden"}},
		{"cli: nur der Programmname", func(m map[string]string) { m[dCli] = "package cli\n\nconst n = \"odata-recorder\"\n" }, []string{"cli.go: kein Kommando gefunden"}},
	}
	zaehle(t, "Fälle", len(faelle), 16)
	for _, f := range faelle {
		m := grund(codeListe()[:3], nil)
		f.bau(m)
		erwarteEingabe(t, f.name, m, f.soll...)
	}
}

// koeder sind die 11 Köder der Zeile `2. **Schlüssel dieser Version.**` aus SPEC-110: die ersten neun tragen `KOEDER_X`
// an einer Stelle, die nicht zu K gehört; die letzten beiden ändern K nicht.
func koeder(s string) []func(m map[string]string) {
	zeile := "2. **" + katS + " dieser Version.** Gepr\xc3\xbcft werden `KOEDER_X`.\n"
	dash := "\xe2\x80\x94"
	ersetze := func(neu string) func(m map[string]string) {
		return func(m map[string]string) { m[dSpez] = strings.Replace(m[dSpez], s, neu, 1) }
	}
	return []func(m map[string]string){
		ersetze(strings.Replace(s, "Version.** Gepr", "Version.** `KOEDER_X` Gepr", 1)),
		ersetze(strings.Replace(s, "`STRICT_COVERAGE`", "`KOEDER_X`", 1)),
		ersetze(s + "  " + zeile),
		ersetze("### SPEC-1100 " + dash + " x\n\n" + zeile + "\n" + s),
		ersetze("### SPEC-110 - x\n\n" + zeile + "\n" + s),
		ersetze(s + "```\n" + zeile + "```\n"),
		ersetze(s + "\n### N\xc3\xa4chster\n\n" + zeile),
		func(m map[string]string) { m[dSpez] += "\n## 6. Mehr\n\n" + zeile },
		ersetze(strings.Replace(s, "`MODE`", "`MODE`, `KOEDER_X und mehr`", 1)),
		ersetze(strings.Replace(s, "`MODE`", "`MODE`, `MODE`", 1)),
		func(m map[string]string) { m[dSpez] = strings.Replace(m[dSpez], s, "", 1) + "\n" + s + "```\nrest\n" },
	}
}

// TestKatSchlK: SPEC-110 Nr. 2 liefert K (Nr. 4 b). Menge 38 = 27 Namen (je einer aus K entfernt: Art 4 für die 26 aus C,
// kein Befund für `SERVICE_ODATA_VERSION`) + 11 Köder (neun lesen `KOEDER_X` zu Unrecht, zwei ändern nichts).
func TestKatSchlK(t *testing.T) {
	c26 := schluessel26()
	for _, e := range schluessel27() {
		m := baumSchl(c26, ohne(schluessel27(), e), c26)
		if e == "SERVICE_ODATA_VERSION" {
			erwarteOk(t, e, m)
			continue
		}
		erwarteAus(t, e, m, 1, katZeilen(katS, art4(e, zeileVon(t, m[dKonfig], "\""+e+"\""))))
	}
	k := koeder(spec110(schluessel27()))
	zaehle(t, "Köder", len(k), 11)
	for i, bau := range k {
		m := baumSchl(plus(c26, "KOEDER_X"), schluessel27(), plus(c26, "KOEDER_X"))
		if i >= 9 {
			m = baumSchl(c26, schluessel27(), c26)
		}
		bau(m)
		name := fmt.Sprintf("Köder %d", i+1)
		if i >= 9 {
			erwarteOk(t, name, m)
			continue
		}
		erwarteAus(t, name, m, 1, katZeilen(katS, art4("KOEDER_X", zeileVon(t, m[dKonfig], "KOEDER_X"))))
	}
}

// TestKatSchlKEingabe: Eingabefehler von SPEC-110 Nr. 2 (Nr. 4 b, 8). Menge 8 = Abschnitt fehlt, Zeile fehlt,
// `Geprüft werden` fehlt, Zeile ohne Span, Zeile nur mit Spans anderer Form (alle `Schlüssel fehlen`), Abschnitt doppelt,
// Zeile doppelt (beide `mehrdeutig`), Datei fehlt; dazu ein Lauf mit Fehler in §4 und in SPEC-110 (§4 steht zuerst).
func TestKatSchlKEingabe(t *testing.T) {
	s := spec110(schluesselBasis())
	spans := "`MODE`, `LISTEN`, `DATA_DIR`"
	zeile := "2. **" + katS + " dieser Version.**"
	fehlen, mehr := "SPEC-110: "+katS+" fehlen", "SPEC-110: "+katS+" mehrdeutig"
	ersetze := func(alt, neu string) func(m map[string]string) {
		return func(m map[string]string) { m[dSpez] = strings.Replace(m[dSpez], alt, neu, 1) }
	}
	faelle := []struct {
		name string
		bau  func(m map[string]string)
		soll string
	}{
		{"Abschnitt fehlt", ersetze(s, ""), fehlen},
		{"Zeile fehlt", ersetze(zeile, "2. **Anderes.**"), fehlen},
		{"Geprüft werden fehlt", ersetze("Gepr\xc3\xbcft werden", "Genannt werden"), fehlen},
		{"Zeile ohne Span", ersetze(spans, "nichts"), fehlen},
		{"Spans anderer Form", ersetze(spans, "`mode`, `Listen`"), fehlen},
		{"Abschnitt doppelt", func(m map[string]string) { m[dSpez] += "\n" + s }, mehr},
		{"Zeile doppelt", ersetze(s, s+zeile+" Gepr\xc3\xbcft werden `MODE`.\n"), mehr},
		{"Datei fehlt", func(m map[string]string) { delete(m, dSpez) }, "spezifikation.md fehlt: spec/spezifikation.md"},
	}
	zaehle(t, "Fälle", len(faelle), 8)
	for _, f := range faelle {
		m := grund(codeListe()[:3], nil)
		f.bau(m)
		erwarteEingabe(t, f.name, m, f.soll)
	}
	m := grund(codeListe()[:3], nil)
	ersetze(s, "")(m)
	ersetze("| ID | Code | Bedingung | Aktion |", "| ID | Code | Bedingung |")(m)
	erwarteEingabe(t, "§4 und SPEC-110", m, "§4: Kopf weicht ab", fehlen)
}

// TestKatSchlCgegenK: jeder Name aus C steht in K (Art 4), ein Name aus K ohne C ist kein Befund (Nr. 1, 4 b, 7). Menge 54 =
// 27 Namen × 2 Richtungen.
func TestKatSchlCgegenK(t *testing.T) {
	alle := schluessel27()
	zaehle(t, "Namen", len(alle), 27)
	for _, e := range alle {
		m := baumSchl(alle, ohne(alle, e), alle)
		erwarteAus(t, e+": C ohne K", m, 1, katZeilen(katS, art4(e, zeileVon(t, m[dKonfig], "\""+e+"\""))))
		erwarteOk(t, e+": K ohne C", baumSchl(ohne(alle, e), alle, ohne(alle, e)))
	}
}

// TestKatSchlPflicht: jeder Name aus H = C hat eine Zeile in der Tabelle Einstellungen, sonst Art 7; ein Name aus K ohne C
// verlangt keine (Nr. 4 c, 7). Menge 30 = 26 Namen aus C (je Zeile entfernt: genau ein Art 7) + `SERVICE_ODATA_VERSION` ohne
// Zeile (kein Befund) + 3 Orte, die die Zeile nicht ersetzen (Prosa, Zaun, Tabelle Kommandos).
func TestKatSchlPflicht(t *testing.T) {
	c26 := schluessel26()
	for _, e := range c26 {
		erwarteAus(t, e, baumSchl(c26, schluessel27(), ohne(c26, e)), 1, katZeilen(katS, e+" fehlt im Handbuch"))
	}
	erwarteOk(t, "SERVICE_ODATA_VERSION", baumSchl(c26, schluessel27(), c26))
	fehlt := katZeilen(katS, "MODE fehlt im Handbuch")
	w := ohne(schluesselBasis(), "MODE")
	prosa := baumSchl(schluesselBasis(), schluesselBasis(), w)
	prosa[dHandbuch] = strings.Replace(prosa[dHandbuch], "## Einstellungen\n\nText.\n", "## Einstellungen\n\nDer Schl\xc3\xbcssel `MODE` steht in Prosa.\n", 1)
	erwarteAus(t, "Prosa", prosa, 1, fehlt)
	zaun := baumSchl(schluesselBasis(), schluesselBasis(), w)
	zaun[dHandbuch] = strings.Replace(zaun[dHandbuch], "## Einstellungen\n\nText.\n", "## Einstellungen\n\nText.\n\n```\n| `MODE` | x |\n```\n", 1)
	erwarteAus(t, "Zaun", zaun, 1, fehlt)
	komm := baumSchl(schluesselBasis(), schluesselBasis(), w)
	nachTabelleK(komm, "| `MODE` | x |\n")
	erwarteAus(t, "Kommandos", komm, 1, fehlt+katZeilen("Kommandos", fmt.Sprintf("Form: Handbuch Zeile %d: \"MODE\" hat nicht die Form", zeileVon(t, komm[dHandbuch], "`MODE`"))))
}

// TestKatSchlHandbuch: die Tabelle Einstellungen (Nr. 4 c, 6 c, 6 d). Menge 11 = Zelle ohne Span, mit zwei Spans, mit Text
// hinter dem Span (Art 2) + Namen `mode`, `ADMIN_*`, `start`, `MODE X` (Art 1) + Name in zwei Zeilen (Art 3) + Tabelle im
// Zaun + Zeile mit Namen anderer Form in der Tabelle Kommandos (kein Befund der Schlüssel) + Tabelle ohne Datenzeile.
func TestKatSchlHandbuch(t *testing.T) {
	zelle := "Zelle: Handbuch Zeile %d: erste Zelle ist nicht genau ein Name in Inline-Code"
	n := 0
	for _, z := range []string{"kein Name", "`a` `b`", "`MODE` Text"} {
		m := baumSchl(schluesselBasis(), schluesselBasis(), schluesselBasis())
		nachTabelleS(m, "| "+z+" | x |\n")
		erwarteAus(t, z, m, 1, katZeilen(katS, fmt.Sprintf(zelle, zeileVon(t, m[dHandbuch], z))))
		n++
	}
	for _, x := range []string{"mode", "ADMIN_*", "start", "MODE X"} {
		z := "`" + x + "`"
		m := baumSchl(schluesselBasis(), schluesselBasis(), schluesselBasis())
		nachTabelleS(m, "| "+z+" | x |\n")
		erwarteAus(t, x, m, 1, katZeilen(katS, fmt.Sprintf("Form: Handbuch Zeile %d: %s hat nicht die Form", zeileVon(t, m[dHandbuch], z), strconv.Quote(x))))
		n++
	}
	doppelt := baumSchl(schluesselBasis(), schluesselBasis(), schluesselBasis())
	nachTabelleS(doppelt, "| `MODE` | x |\n| `MODE` | y |\n")
	erwarteAus(t, "doppelt", doppelt, 1, katZeilen(katS, fmt.Sprintf("MODE doppelt in Handbuch, erstes Vorkommen Zeile %d", zeileVon(t, doppelt[dHandbuch], "`MODE`"))))
	zaun := baumSchl(schluesselBasis(), schluesselBasis(), schluesselBasis())
	nachTabelleS(zaun, "\n```\n| `a` | x |\n| `OTHER_X` | y |\n```\n")
	erwarteOk(t, "Zaun", zaun)
	komm := baumSchl(schluesselBasis(), schluesselBasis(), schluesselBasis())
	nachTabelleK(komm, "| `OTHER_X` | y |\n")
	erwarteAus(t, "Kommandos", komm, 1, katZeilen("Kommandos", fmt.Sprintf("Form: Handbuch Zeile %d: \"OTHER_X\" hat nicht die Form", zeileVon(t, komm[dHandbuch], "`OTHER_X`"))))
	leer := baumSchl(schluesselBasis(), schluesselBasis(), nil)
	erwarteAus(t, "ohne Datenzeile", leer, 1, katZeilen(katS, "DATA_DIR fehlt im Handbuch", "LISTEN fehlt im Handbuch", "MODE fehlt im Handbuch"))
	zaehle(t, "Fälle", n+4, 11)
}
